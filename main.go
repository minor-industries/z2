package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/codelab/cmd/z2/database"
	handler2 "github.com/minor-industries/codelab/cmd/z2/handler"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"github.com/minor-industries/codelab/cmd/z2/source/replay"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"os"
	"time"

	"tinygo.org/x/bluetooth"
)

var opts struct {
	Replay bool `long:"replay"`
}

func run() error {
	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	db, err := database.Get(os.ExpandEnv("$HOME/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get database")
	}
	_ = db

	allSeries, err := database.LoadAllSeries(db)
	if err != nil {
		return errors.Wrap(err, "load series")
	}

	br := broker.NewBroker()
	go br.Start()

	errCh := make(chan error)
	go publishPrometheusMetrics(errCh, br)

	ctx, cancel := context.WithCancel(context.Background())
	handler, err := handler2.NewBikeHandler(
		db,
		cancel,
		ctx,
		allSeries,
		br,
	)
	if err != nil {
		return errors.Wrap(err, "new handler")
	}
	go handler.Monitor()

	go func() {
		errCh <- serve(db, br, allSeries)
	}()

	go func() {
		if opts.Replay {
			err = replay.Run(ctx, "raw.txt", handler.Handle)
		} else {
			err = source.Run(ctx, address, map[source.CBKey]func(time.Time, []byte) error{
				source.CBKey{
					bluetooth.ServiceUUIDFitnessMachine,
					bluetooth.CharacteristicUUIDIndoorBikeData,
				}: handler.Handle,
			})
		}
		errCh <- err
	}()

	return <-errCh
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
