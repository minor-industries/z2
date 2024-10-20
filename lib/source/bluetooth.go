//go:build !wasm

package source

import (
	"context"
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"reflect"
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

type DeviceInfo struct {
	Address string
	Kind    string
	Name    string
}

type Device struct {
	Info     DeviceInfo
	Source   Source
	Callback MessageCallback
}

type foundDevice struct {
	Device Device
	Result bluetooth.ScanResult
}

func Connect(
	ctx context.Context,
	errCh chan error,
	devices []Device,
	disconnect chan struct{},
) {
	var adapter = bluetooth.DefaultAdapter

	addresses := set.Set[string]{}
	kinds := set.Set[string]{}

	for _, device := range devices {
		if device.Info.Address != strings.ToLower(device.Info.Address) {
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

		fmt.Printf("looking for %s (%s) at %s\n", device.Info.Kind, device.Info.Name, device.Info.Address)

		addresses.Add(device.Info.Address)
		kinds.Add(device.Info.Kind)
	}

	deviceMap := map[string]Device{}
	for _, device := range devices {
		deviceMap[strings.ToLower(device.Info.Address)] = device
	}

	var err error
	enableOnce.Do(func() {
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
				// we're not looking for this address
				return
			}

			if !addresses.Has(scannedAddr) {
				// we've already found this address
				return
			}

			if !kinds.Has(dev.Info.Kind) {
				// we've already found one of these kinds of devices
				return
			}

			fmt.Printf("found %s (%s) at %s\n", dev.Info.Kind, dev.Info.Name, dev.Info.Address)
			addresses.Delete(dev.Info.Address)
			kinds.Delete(dev.Info.Kind)

			ch <- foundDevice{
				Device: dev,
				Result: result,
			}

			if kinds.Len() == 0 {
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

			err := handleDevice(errCh, adapter, result, disconnect)
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

func handleDevice(
	errCh chan error,
	adapter *bluetooth.Adapter,
	found foundDevice,
	disconnect chan struct{},
) error {
	device, err := adapter.Connect(found.Result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return errors.Wrap(err, "connect")
	}

	dev := found.Device
	fmt.Printf("connected to %s (%s) at %s\n", dev.Info.Kind, dev.Info.Name, dev.Info.Address)

	src := found.Device.Source
	srvcs, err := device.DiscoverServices(lo.Map(src.Services(), func(item UUID, _ int) bluetooth.UUID {
		return mustParseUUID(item)
	}))
	if err != nil {
		return errors.Wrap(err, "discover services")
	}

	for _, srvc := range srvcs {
		//fmt.Println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics(lo.Map(src.Characteristics(), func(item UUID, _ int) bluetooth.UUID {
			return mustParseUUID(item)
		}))
		if err != nil {
			fmt.Println(err)
		}
		for _, char := range chars {
			//fmt.Println("-- 16 bit", srvc.Is16Bit())
			//fmt.Println("-- characteristic", char.UUID().String())

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

	go func() {
		<-disconnect
		fmt.Printf("disconnect %s\n", reflect.TypeOf(found.Device.Source).String())
		err := device.Disconnect()
		if err != nil {
			errCh <- errors.Wrap(err, "disconnect")
		}
	}()

	return nil
}
