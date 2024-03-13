package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	handler2 "github.com/minor-industries/codelab/cmd/z2/handler"
	"github.com/minor-industries/codelab/cmd/z2/rtgraph/database"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"github.com/minor-industries/codelab/cmd/z2/source/replay"
	"github.com/minor-industries/platform/common/broker"
	"github.com/pkg/errors"
	"os"
)

var opts struct {
	Replay bool   `long:"replay"`
	Source string `long:"source" required:"true" env:"SOURCE"`
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
	//src := &handler2.BikeSource{}
	srcAddr, src := getSource(opts.Source)
	fmt.Printf("looking for %s at address %s\n", opts.Source, srcAddr)

	handler, err := handler2.NewBikeHandler(
		db,
		src,
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
			err = replay.Run(
				ctx,
				errCh,
				"raw.txt",
				handler.Handle,
			)
		} else {
			err = source.Run(
				ctx,
				errCh,
				srcAddr,
				src,
				handler.Handle,
			)
		}
		errCh <- err
	}()

	return <-errCh
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
