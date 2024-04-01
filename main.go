package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/rtgraph"
	handler2 "github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/html"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/replay"
	"github.com/pkg/errors"
	"os"
)

var opts struct {
	Source string `long:"source" required:"true" env:"SOURCE"`

	Replay   bool   `long:"replay"`
	ReplayDB string `long:"replay-db"`
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	errCh := make(chan error)

	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	avg := func(name string, seconds uint) rtgraph.ComputedReq {
		return rtgraph.ComputedReq{
			SeriesName: name,
			Function:   "avg",
			Seconds:    seconds,
		}
	}

	graph, err := rtgraph.New(
		os.ExpandEnv("$HOME/z2.db"),
		errCh,
		[]string{
			"bike_instant_speed",
			"bike_instant_cadence",
			"bike_total_distance",
			"bike_resistance_level",
			"bike_instant_power",
			"bike_total_energy",
			"bike_energy_per_hour",
			"bike_energy_per_minute",
			"bike_heartrate",

			"rower_stroke_count",
			"rower_power",
			"rower_speed",
			"rower_spm",
		},
		[]rtgraph.ComputedReq{
			avg("bike_instant_speed", 30),
			avg("bike_instant_cadence", 30),
			avg("bike_instant_power", 30),
			avg("bike_instant_speed", 900),

			avg("rower_power", 30),
			avg("rower_spm", 30),
			avg("rower_power", 900),
		},
	)
	if err != nil {
		return errors.Wrap(err, "new graph")
	}

	graph.StaticFiles(html.FS,
		"index.html", "text/html",
		"bike.html", "text/html",
		"rower.html", "text/html",
	)

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
		if opts.ReplayDB != "" {
			err = replay.RunDB(
				ctx,
				errCh,
				os.ExpandEnv(opts.ReplayDB),
				handler.Handle,
			)
		} else if opts.Replay {
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
