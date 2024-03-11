package source

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
	"tinygo.org/x/bluetooth"
)

type CBKey [2]bluetooth.UUID
type Callbacks map[CBKey]func(t time.Time, msg []byte) error

func Run(
	ctx context.Context,
	address string,
	callbacks Callbacks,
) error {
	var adapter = bluetooth.DefaultAdapter

	fmt.Println("enabling")

	if err := adapter.Enable(); err != nil {
		return errors.Wrap(err, "enable adapter")
	}

	ch := make(chan bluetooth.ScanResult, 1)

	fmt.Println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		//fmt.Println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == address {
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

	fmt.Println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{
		bluetooth.ServiceUUIDFitnessMachine,
	})
	if err != nil {
		return errors.Wrap(err, "discover services")
	}

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

			cb, ok := callbacks[CBKey{srvc.UUID(), char.UUID()}]
			if ok {
				fmt.Println("enabling notifications for", srvc.UUID().String(), char.UUID().String())
				if err := char.EnableNotifications(func(buf []byte) {
					cb(time.Now(), buf)
				}); err != nil {
					return errors.Wrap(err, "enable notifications")
				}
			}
		}
	}

	<-ctx.Done()

	err = device.Disconnect()
	if err != nil {
		fmt.Println(err)
	}

	return errors.New("exited")
}
