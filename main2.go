package main

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"

	"tinygo.org/x/bluetooth"
)

func connectAddress() string {
	return "FC:38:34:32:0D:69"
}

type cbKey [2]bluetooth.UUID
type Callbacks map[cbKey]func(msg []byte)

func run2(
	adapter *bluetooth.Adapter,
	callbacks Callbacks,
) error {
	ch := make(chan bluetooth.ScanResult, 1)

	// Start scanning.
	fmt.Println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.Address.String() == connectAddress() {
			fmt.Println("found device:", result.Address.String(), result.RSSI, result.LocalName())
			adapter.StopScan()
			ch <- result
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			return errors.Wrap(err, "connect")
		}

		fmt.Println("connected to ", result.Address.String())
	}

	// get services
	fmt.Println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{
		bluetooth.ServiceUUIDFitnessMachine,
	})
	must("discover services", err)

	// buffer to retrieve characteristic data
	buf := make([]byte, 255)

	for _, srvc := range srvcs {
		fmt.Println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{
			bluetooth.CharacteristicUUIDIndoorBikeData,
		})
		if err != nil {
			fmt.Println(err)
		}
		for _, char := range chars {
			fmt.Println("-- 16 bit", srvc.Is16Bit())
			fmt.Println("-- characteristic", char.UUID().String())
			n, err := char.Read(buf)
			if err != nil {
				fmt.Println("    ", err.Error())
			} else {
				fmt.Println("    data bytes", strconv.Itoa(n))
				fmt.Println("    value =", string(buf[:n]))
			}

			cb, ok := callbacks[cbKey{srvc.UUID(), char.UUID()}]
			if ok {
				fmt.Println("enabling notifications for", srvc.UUID().String(), char.UUID().String())
				if err := char.EnableNotifications(cb); err != nil {
					return errors.Wrap(err, "enable notifications")
				}
			}
		}
	}

	<-time.After(10 * time.Second)

	err = device.Disconnect()
	if err != nil {
		fmt.Println(err)
	}

	return errors.New("exited")
}

func main() {
	var adapter = bluetooth.DefaultAdapter

	fmt.Println("enabling")

	adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		fmt.Println("connect handler: connected:", connected)
	})

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	t0 := time.Now()
	for {
		err := run2(adapter, map[cbKey]func([]byte){
			cbKey{
				bluetooth.ServiceUUIDFitnessMachine,
				bluetooth.CharacteristicUUIDIndoorBikeData,
			}: func(msg []byte) {
				dt := time.Now().Sub(t0).Seconds()
				fmt.Printf("%7.2f bikedata: %s\n", dt, hex.EncodeToString(msg))
			},
		})
		fmt.Println("run exited, error:", err)
		time.Sleep(500 * time.Millisecond)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
