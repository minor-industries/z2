package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/evt"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/pkg/errors"
	"os"
	"time"
)

const deviceName = "go-bike"

func run() error {
	disconnectCh := make(chan error)

	d, err := linux.NewDeviceWithName(
		deviceName,
		ble.OptDisconnectHandler(func(complete evt.DisconnectionComplete) {
			disconnectCh <- fmt.Errorf("disconnect %d", complete.Reason())
		}),
	)
	if err != nil {
		return errors.Wrap(err, "new device")
	}
	ble.SetDefaultDevice(d)

	//TODO: signal handler

	switch os.Args[1] {
	case "scan":
		err = scan()
	case "connect":
		err = connectLoop(disconnectCh)
	default:
		err = errors.New("unknown verb")
	}

	return err
}

func connectLoop(disconnectCh chan error) error {
	for {
		err := connectAndSubscribe(disconnectCh)
		if err != nil {
			fmt.Println("connect error:", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func connectAndSubscribe(disconnectCh chan error) error {
	fmt.Println("connecting...")
	topCtx, topCancel := context.WithCancel(context.Background())
	defer topCancel()

	ctx, connectCancel := context.WithTimeout(topCtx, time.Minute)

	conn, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		if a.LocalName() == "IC Bike" {
			fmt.Println(a.LocalName(), a.Addr(), a.RSSI(), a.Connectable())
			return a.Connectable()
		}
		return false
	})
	connectCancel()
	if err != nil {
		return errors.Wrap(err, "connect")
	}

	//fmt.Println(conn.Name(), conn.Addr())

	p, err := conn.DiscoverProfile(false)
	if err != nil {
		return errors.Wrap(err, "discover profile")
	}

	for _, service := range p.Services {
		//fmt.Println(i, service.UUID.String())
		for _, ch := range service.Characteristics {
			//fmt.Println(" ", j, ch.UUID.String())
			if service.UUID.String() == "1826" && ch.UUID.String() == "2ad2" {
				sub(conn, service, ch)
			}
		}
	}

	select {
	case err = <-disconnectCh:
	case <-topCtx.Done():
	}

	return err
}

func sub(conn ble.Client, service *ble.Service, ch *ble.Characteristic) {
	fmt.Println("  subscribe", service.UUID, ch.UUID)
	err := conn.Subscribe(ch, false, func(req []byte) {
		fmt.Println(service.UUID, ch.UUID, hex.Dump(req))
		datum := parser.ParseIndoorBikeData(req)
		datum.AllPresentFields(func(series string, value float64) {
			fmt.Println(series, value)
		})
	})
	if err != nil {
		fmt.Println("error:", errors.Wrap(err, "subscribe"))
	}
}

func scan() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	//ble.WithSigHandler(ctx, cancelFunc)
	err := ble.Scan(
		ctx,
		false,
		func(a ble.Advertisement) {
			fmt.Println(a.RSSI(), a.Addr(), a.LocalName())
		},
		func(a ble.Advertisement) bool {
			return true
		},
	)
	if err != nil {
		return errors.Wrap(err, "scan")
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
