package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/source"
	"github.com/minor-industries/codelab/cmd/bike/source/replay"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"os"
	"time"

	"tinygo.org/x/bluetooth"
)

func run() error {
	db, err := database.Get(os.ExpandEnv("$HOME/bike.db"))
	if err != nil {
		return errors.Wrap(err, "get database")
	}
	_ = db

	br := broker.NewBroker()
	go br.Start()

	ctx, cancel := context.WithCancel(context.Background())
	handler := &bikeHandler{
		t0:     time.Now(),
		cancel: cancel,
		ctx:    ctx,
		broker: br,
	}
	go handler.Monitor()

	errCh := make(chan error)

	go func() {
		errCh <- serve(nil, br)
	}()

	go func() {
		if false {
			err = source.Run(ctx, address, map[source.CBKey]func([]byte){
				source.CBKey{
					bluetooth.ServiceUUIDFitnessMachine,
					bluetooth.CharacteristicUUIDIndoorBikeData,
				}: handler.Handle,
			})
		} else {
			err = replay.Run(ctx, "raw.txt", handler.Handle)
		}
		errCh <- err
	}()

	return <-errCh
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
