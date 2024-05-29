package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/static"
	"github.com/pkg/errors"
	"math"
	"net/http"
	"os"
	"time"
)

var opts struct {
	Source string `long:"source" required:"true" env:"SOURCE"`

	ReplayDB   string `long:"replay-db"`
	NoReplayDB string `long:"no-replay-db"`

	Port int `long:"port" default:"8077" env:"PORT"`

	HeartrateMonitors []string `long:"hrm" env:"HRM"`

	StaticPath string `long:"static-path" required:"false"`

	RemoveDB bool `long:"remove-db" required:"false"`
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	errCh := make(chan error)

	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	dbPath := os.ExpandEnv("$HOME/z2.db")

	if opts.RemoveDB {
		_ = os.Remove(dbPath)
	}

	db, err := database.Get(dbPath, errCh)
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

	router := graph.GetEngine()

	if opts.StaticPath != "" {
		router.Static("/static", opts.StaticPath)
	} else {
		router.StaticFS("/static", http.FS(static.FS))
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static/index.html")
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	apiHandler := &ApiServer{db: db}
	router.Any("/twirp/api.Calendar/*Method", gin.WrapH(api.NewCalendarServer(apiHandler, nil)))
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

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

	go func() {
		const (
			target       = 41.5
			allowedError = 0.1
			errorSteps   = 5
			stepSize     = allowedError / errorSteps

			minMaxWindow = 1.0
		)

		now := time.Now()
		msgCh := make(chan *messages.Data)
		go graph.Subscribe(&subscription.Request{
			Series: []string{
				"bike_instant_speed | gt 0 | avg 10m triangle",
			},
			WindowSize:  uint64((10 * time.Minute).Milliseconds()),
			LastPointMs: 0,
			MaxGapMs:    uint64((5 * time.Second).Milliseconds()),
		}, now, msgCh)

		for m := range msgCh {
			for _, s := range m.Series {
				//fmt.Println(i, time.Now().UnixMilli(), s.Timestamps, s.Values)
				ts := time.UnixMilli(s.Timestamps[0])
				value := s.Values[0]

				e := value - target
				steps := int(math.Round(math.Abs(e) / stepSize))
				steps = min(steps, errorSteps)
				steps = -sign(e) * steps
				fmt.Println(value, e, e/stepSize, steps)

				if err := graph.CreateValue("bike_instant_speed_min", ts, target-minMaxWindow/2.0); err != nil {
					panic(err)
				}
				if err := graph.CreateValue("bike_instant_speed_max", ts, target+minMaxWindow/2.0); err != nil {
					panic(err)
				}
			}
		}
	}()

	return <-errCh
}

func sign(x float64) int {
	if x < 0 {
		return -1
	}
	return 1
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
