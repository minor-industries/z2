package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/pkg/errors"
	"os"
	"time"
)

const deviceName = "go-bike"

func run() error {
	d, err := linux.NewDeviceWithName(deviceName)
	if err != nil {
		return errors.Wrap(err, "new device")
	}
	ble.SetDefaultDevice(d)

	//TODO: signal handler

	switch os.Args[1] {
	case "scan":
		err = scan()
	case "connect":
		err = connect()
	default:
		err = errors.New("unknown verb")
	}

	return err
}

func connect() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		return a.LocalName() == "IC Bike"
	})
	if err != nil {
		return errors.Wrap(err, "connect")
	}

	fmt.Println(conn.Name(), conn.Addr())

	p, err := conn.DiscoverProfile(false)
	if err != nil {
		return errors.Wrap(err, "discover profile")
	}

	for i, service := range p.Services {
		fmt.Println(i, service.UUID.String())
		for j, ch := range service.Characteristics {
			fmt.Println(" ", j, ch.UUID.String())
			if service.UUID.String() == "1826" && ch.UUID.String() == "2ad2" {
				sub(conn, service, ch)
			}
		}
	}

	select {}
	//conn.DiscoverServices()
	return nil
}

func sub(conn ble.Client, service *ble.Service, ch *ble.Characteristic) {
	fmt.Println("  subscribe", service.UUID, ch.UUID)
	err := conn.Subscribe(ch, false, func(req []byte) {
		fmt.Println(service.UUID, ch.UUID, hex.Dump(req))
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
