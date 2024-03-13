package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	handler2 "github.com/minor-industries/codelab/cmd/z2/handler"
	"github.com/minor-industries/codelab/cmd/z2/rtgraph"
	"github.com/minor-industries/codelab/cmd/z2/source"
	"github.com/minor-industries/codelab/cmd/z2/source/replay"
	"github.com/pkg/errors"
	"os"
)

var opts struct {
	Replay bool   `long:"replay"`
	Source string `long:"source" required:"true" env:"SOURCE"`
}

func run() error {
	errCh := make(chan error)

	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	graph, err := rtgraph.New(
		os.ExpandEnv("$HOME/z2.db"),
		errCh,
	)
	if err != nil {
		return errors.Wrap(err, "new graph")
	}

	ctx, cancel := context.WithCancel(context.Background())
	//src := &handler2.BikeSource{}
	srcAddr, src := getSource(opts.Source)
	fmt.Printf("looking for %s at address %s\n", opts.Source, srcAddr)

	handler, err := handler2.NewBikeHandler(
		graph,
		src,
		cancel,
		ctx,
	)
	if err != nil {
		return errors.Wrap(err, "new handler")
	}
	go handler.Monitor()

	go func() {
		errCh <- graph.RunServer("0.0.0.0:8077")
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
