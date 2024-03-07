package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/evt"
	"github.com/google/uuid"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/pkg/errors"
	"os"
	"time"
)

const deviceName = "go-bike"

type subscriptionCallback func(service *ble.Service, ch *ble.Characteristic, req []byte)

func run() error {
	db, err := database.Get(os.ExpandEnv("$HOME/bike.db"))
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	allSeries, err := database.LoadSeries(db)
	if err != nil {
		return errors.Wrap(err, "load series")
	}

	fp, err := os.OpenFile(os.ExpandEnv("$HOME/raw.txt"), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return errors.Wrap(err, "open")
	}

	cb := func(service *ble.Service, ch *ble.Characteristic, req []byte) {
		t := time.Now()
		fmt.Println(service.UUID, ch.UUID, hex.Dump(req))
		datum := parser.ParseIndoorBikeData(req)
		line := fmt.Sprintf("%d %s\n", t.UnixMilli(), hex.EncodeToString(req))
		_, err := fp.WriteString(line)
		if err != nil {
			panic(errors.Wrap(err, "write line"))
		}
		datum.AllPresentFields(func(seriesName string, value float64) {
			series, ok := allSeries[seriesName]
			if !ok {
				panic(fmt.Errorf("unknown database series: %s", seriesName))
			}
			tx := db.Create(&database.Value{
				ID:        uuid.New(),
				Timestamp: t,
				Value:     value,
				Series:    series,
			})
			if tx.Error != nil {
				panic(errors.Wrap(tx.Error, "create value"))
			}
		})
	}

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
		err = connectLoop(disconnectCh, cb)
	default:
		err = errors.New("unknown verb")
	}

	return err
}

func connectLoop(
	disconnectCh chan error,
	callback subscriptionCallback,
) error {
	for {
		err := connectAndSubscribe(disconnectCh, callback)
		if err != nil {
			fmt.Println("connect error:", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func connectAndSubscribe(
	disconnectCh chan error,
	callback subscriptionCallback,
) error {
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
				sub(conn, service, ch, callback)
			}
		}
	}

	select {
	case err = <-disconnectCh:
	case <-topCtx.Done():
	}

	return err
}

func sub(
	conn ble.Client,
	service *ble.Service,
	ch *ble.Characteristic,
	callback subscriptionCallback,
) {
	fmt.Println("  subscribe", service.UUID, ch.UUID)
	err := conn.Subscribe(ch, false, func(req []byte) {
		callback(service, ch, req)
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
