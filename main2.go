package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"tinygo.org/x/bluetooth"
)

func connectAddress() string {
	return "FC:38:34:32:0D:69"
}

var adapter = bluetooth.DefaultAdapter

func main() {
	fmt.Println("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	ch := make(chan bluetooth.ScanResult, 1)

	// Start scanning.
	fmt.Println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		fmt.Println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == connectAddress() {
			adapter.StopScan()
			ch <- result
		}
	})

	adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if !connected {
			panic("not connected")
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			fmt.Println(err.Error())
			return
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

			if srvc.UUID() == bluetooth.ServiceUUIDFitnessMachine &&
				char.UUID() == bluetooth.CharacteristicUUIDIndoorBikeData {
				if err := char.EnableNotifications(func(buf []byte) {
					fmt.Println(time.Now().String(), "notify", hex.Dump(buf))
				}); err != nil {
					fmt.Println("error enabling notifications:", err.Error())
				}
			}
		}
	}

	select {}

	err = device.Disconnect()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("done")
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
