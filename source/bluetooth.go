package source

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
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
	Services() []bluetooth.UUID
	Characteristics() []bluetooth.UUID
}

var enableOnce sync.Once

func Run(
	ctx context.Context,
	errCh chan error,
	address string,
	src Source,
	connectCallback func(),
	messageCallback MessageCallback,
) error {
	var adapter = bluetooth.DefaultAdapter

	var err error
	enableOnce.Do(func() {
		fmt.Println("enabling")
		err = adapter.Enable()
	})
	if err != nil {
		return errors.Wrap(err, "enable adapter")
	}

	ch := make(chan bluetooth.ScanResult, 1)

	fmt.Println("scanning...", address)
	err = adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		//fmt.Println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if strings.ToLower(result.Address.String()) == strings.ToLower(address) {
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

		fmt.Println("connected to", result.Address.String())
		if connectCallback != nil {
			connectCallback()
		}
	}

	fmt.Println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices(src.Services())
	if err != nil {
		return errors.Wrap(err, "discover services")
	}

	buf := make([]byte, 255)
	for _, srvc := range srvcs {
		fmt.Println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics(src.Characteristics())
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

			fmt.Println("enabling notifications", srvc.UUID().String(), char.UUID().String())
			{
				// capture loop variables for use in closure below
				svUUID := srvc.UUID()
				chUUID := char.UUID()
				if err := char.EnableNotifications(func(buf []byte) {
					err := messageCallback(time.Now(), svUUID, chUUID, buf)
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
