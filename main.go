package main

import (
	"context"
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

	//conn.DiscoverServices()
	return nil
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
