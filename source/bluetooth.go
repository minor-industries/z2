//go:build !wasm

package source

import (
	"context"
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"tinygo.org/x/bluetooth"
)

func mustParseUUID(s UUID) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(string(s))
	if err != nil {
		panic(err)
	}
	return uuid
}

var enableOnce sync.Once

type Device struct {
	Address  string
	Source   Source
	Callback MessageCallback
}

type foundDevice struct {
	Device Device
	Result bluetooth.ScanResult
}

func Run(
	ctx context.Context,
	errCh chan error,
	devices []Device,
) {
	var adapter = bluetooth.DefaultAdapter

	addresses := set.Set[string]{}
	for _, device := range devices {
		if device.Address != strings.ToLower(device.Address) {
			errCh <- errors.New("device addresses must be given in lowercase")
			return
		}

		for _, svc := range device.Source.Services() {
			if string(svc) != strings.ToLower(string(svc)) {
				errCh <- errors.New("service UUIDs must be given in lowercase")
				return
			}
		}

		for _, ch := range device.Source.Characteristics() {
			if string(ch) != strings.ToLower(string(ch)) {
				errCh <- errors.New("characteristic UUIDs must be given in lowercase")
				return
			}
		}

		addresses.Add(device.Address)
	}

	deviceMap := map[string]Device{}
	for _, device := range devices {
		deviceMap[strings.ToLower(device.Address)] = device
	}

	var err error
	enableOnce.Do(func() {
		fmt.Println("enabling bluetooth adapter")
		err = adapter.Enable()
	})
	if err != nil {
		errCh <- errors.Wrap(err, "enable adapter")
		return
	}

	ch := make(chan foundDevice, 10)

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			//fmt.Println("scanned device:", result.Address.String(), result.RSSI, result.LocalName())
			scannedAddr := strings.ToLower(result.Address.String())

			dev, ok := deviceMap[scannedAddr]
			if !ok {
				return
			}

			if !addresses.Has(scannedAddr) {
				return
			}

			fmt.Println("scannedAddr", dev.Address, reflect.TypeOf(dev.Source).String())
			addresses.Delete(dev.Address)

			ch <- foundDevice{
				Device: dev,
				Result: result,
			}

			if addresses.Len() == 0 {
				err := adapter.StopScan()
				if err != nil {
					errCh <- errors.Wrap(err, "stop scan")
				}
				close(ch)
			}
		})

		if err != nil {
			errCh <- err
			return
		}
	}()

	for {
		select {
		case result, ok := <-ch:
			if !ok {
				return
			}
			err := handleDevice(errCh, adapter, result)
			if err != nil {
				errCh <- err
				return
			}
		case <-ctx.Done():
			// TODO: cleanup? what cleanup?
			return
		}
	}
}

func handleDevice(errCh chan error, adapter *bluetooth.Adapter, found foundDevice) error {
	device, err := adapter.Connect(found.Result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return errors.Wrap(err, "connect")
	}

	fmt.Println("connected to", found.Result.Address.String(), reflect.TypeOf(found.Device.Source).String())

	src := found.Device.Source
	srvcs, err := device.DiscoverServices(lo.Map(src.Services(), func(item UUID, _ int) bluetooth.UUID {
		return mustParseUUID(item)
	}))
	if err != nil {
		return errors.Wrap(err, "discover services")
	}

	buf := make([]byte, 255)
	for _, srvc := range srvcs {
		fmt.Println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics(lo.Map(src.Characteristics(), func(item UUID, _ int) bluetooth.UUID {
			return mustParseUUID(item)
		}))
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
					err := found.Device.Callback(time.Now(), UUID(svUUID.String()), UUID(chUUID.String()), buf)
					if err != nil {
						errCh <- errors.Wrap(err, "callback")
					}
				}); err != nil {
					return errors.Wrap(err, "enable notifications")
				}
			}
		}
	}

	return nil
}
