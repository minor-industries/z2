package source

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
	"tinygo.org/x/bluetooth"
)

type MessageCallback func(
	t time.Time,
	service bluetooth.UUID,
	characteristic bluetooth.UUID,
	msg []byte,
) error

type Message struct {
	Timestamp      time.Time
	Service        bluetooth.UUID
	Characteristic bluetooth.UUID
	Msg            []byte
}

type Value struct {
	Name      string
	Timestamp time.Time
	Value     float64
}

type Source interface {
	Convert(msg Message) []Value
	SubscriptionFilter(
		Service bluetooth.UUID,
		Characteristic bluetooth.UUID,
	) bool
}

func Run(
	ctx context.Context,
	errCh chan error,
	address string,
	src Source,
	callback MessageCallback,
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

			if src.SubscriptionFilter(srvc.UUID(), char.UUID()) {
				fmt.Println("enabling notifications", srvc.UUID().String(), char.UUID().String())
				if err := char.EnableNotifications(func(buf []byte) {
					err := callback(time.Now(), srvc.UUID(), char.UUID(), buf)
					if err != nil {
						errCh <- errors.Wrap(err, "callback")
					}
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
