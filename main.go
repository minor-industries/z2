package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/html"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/replay"
	"github.com/pkg/errors"
	"os"
)

var opts struct {
	Source string `long:"source" required:"true" env:"SOURCE"`

	Replay   bool   `long:"replay"`
	ReplayDB string `long:"replay-db"`

	Port int `long:"port" default:"8077" env:"PORT"`

	HeartrateMonitors []string `long:"hrm" env:"HRM"`
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	errCh := make(chan error)

	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	db, err := database.Get(os.ExpandEnv("$HOME/z2.db"), errCh)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	graph, err := rtgraph.New(
		db,
		errCh,
		rtgraph.Opts{},
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

			"heartrate",
		},
	)
	if err != nil {
		return errors.Wrap(err, "new graph")
	}

	graph.StaticFiles(html.FS,
		"index.html", "text/html",
		"bike.html", "text/html",
		"rower.html", "text/html",
		"handlebars.js", "application/javascript",
		"data.js", "application/javascript",
		"data.html", "text/html",
	)

	router := graph.GetEngine()

	apiServer := api.NewCalendarServer(&ApiServer{}, nil)
	router.Any("/twirp/api.Calendar/*Method", gin.WrapH(apiServer))

	ctx, cancel := context.WithCancel(context.Background())
	//src := &mainHandler.BikeSource{}
	srcAddr, src := getSource(opts.Source)
	fmt.Printf("looking for %s at address %s\n", opts.Source, srcAddr)

	mainHandler, err := handler.NewBikeHandler(
		graph,
		db,
		src,
		cancel,
		ctx,
	)
	if err != nil {
		return errors.Wrap(err, "new handler")
	}
	go mainHandler.Monitor()

	setupHRMs := func() {
		// TODO: this needs a lot of reorganization/cleanup
		for _, addr := range opts.HeartrateMonitors {
			addr := addr
			hrmSrc := &heartrate.Source{}
			h, err := handler.NewBikeHandler(
				graph,
				db,
				hrmSrc,
				cancel,
				ctx,
			)
			if err != nil {
				errCh <- errors.Wrap(err, "new handler")
				return
			}

			go func() {
				errCh <- source.Run(
					ctx,
					errCh,
					addr,
					hrmSrc,
					nil,
					h.Handle,
				)
			}()
		}
	}

	go func() {
		errCh <- graph.RunServer(fmt.Sprintf("0.0.0.0:%d", opts.Port))
	}()

	go func() {
		if opts.ReplayDB != "" {
			go setupHRMs()
			err = replay.RunDB(
				ctx,
				errCh,
				os.ExpandEnv(opts.ReplayDB),
				mainHandler.Handle,
			)
		} else if opts.Replay {
			err = replay.Run(
				ctx,
				errCh,
				"raw.txt",
				mainHandler.Handle,
			)
		} else {
			err = source.Run(
				ctx,
				errCh,
				srcAddr,
				src,
				func() {
					go setupHRMs()
				},
				mainHandler.Handle,
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
