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
	"github.com/minor-industries/z2/variables"
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

type App struct {
	Graph *rtgraph.Graph
	Vars  *variables.Cache
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	app := &App{}

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

	app.Graph, err = rtgraph.New(
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

	router := app.Graph.GetEngine()

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

	app.Vars = variables.NewCache()

	apiHandler := &ApiServer{db: db, vars: app.Vars}
	router.Any("/twirp/api.Calendar/*Method", gin.WrapH(api.NewCalendarServer(apiHandler, nil)))
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	//src := &mainHandler.BikeSource{}
	srcAddr, src := getSource(opts.Source)
	fmt.Printf("looking for %s at address %s\n", opts.Source, srcAddr)

	mainHandler, err := handler.NewBikeHandler(
		app.Graph,
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
				app.Graph,
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
		errCh <- app.Graph.RunServer(fmt.Sprintf("0.0.0.0:%d", opts.Port))
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

	go app.ComputeBounds()
	go app.ComputePace()

	return <-errCh
}

func (app *App) ComputeBounds() {
	const (
		allowedError = 0.5
		errorSteps   = 5
		stepSize     = allowedError / errorSteps

		minMaxWindow = 1.0
		outStepSize  = minMaxWindow / 2 / errorSteps
	)

	now := time.Now()
	msgCh := make(chan *messages.Data)
	go app.Graph.Subscribe(&subscription.Request{
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

			target, _ := app.Vars.GetOne("bike_target_speed")

			e := value - target
			steps := int(math.Round(math.Abs(e) / stepSize))
			steps = min(steps, errorSteps)
			steps = -sign(e) * steps

			outAdjust := float64(steps) * outStepSize

			minTarget := target + outAdjust - minMaxWindow/2.0
			maxTarget := target + minMaxWindow/2.0 + outAdjust

			fmt.Println(value, e, e/stepSize, steps)

			if err := app.Graph.CreateValue(
				"bike_instant_speed_min",
				ts,
				minTarget,
			); err != nil {
				panic(err)
			}

			if err := app.Graph.CreateValue(
				"bike_instant_speed_max",
				ts,
				maxTarget,
			); err != nil {
				panic(err)
			}
		}
	}
}

func (app *App) ComputePace() {
	now := time.Now()
	msgCh := make(chan *messages.Data)
	go app.Graph.Subscribe(&subscription.Request{
		Series: []string{
			"bike_instant_speed | avg 30s triangle",
			"bike_instant_speed_min",
			"bike_instant_speed_max",
		},
		WindowSize:  uint64((30 * time.Second).Milliseconds()),
		LastPointMs: 0,
		MaxGapMs:    uint64((5 * time.Second).Milliseconds()),
	}, now, msgCh)

	var value, minTarget, maxTarget float64

	for m := range msgCh {
		if len(m.Series) == 0 {
			continue
		}

		for _, s := range m.Series {
			switch s.Pos {
			case 0:
				value = s.Values[0]
				fmt.Println("value", value)
			case 1:
				minTarget = s.Values[0]
				fmt.Println("minTarget", minTarget)
			case 2:
				maxTarget = s.Values[0]
				fmt.Println("maxTarget", maxTarget)
			}
		}

		if value == 0 || minTarget == 0 || maxTarget == 0 {
			continue
		}

		var state string
		switch {
		case value > maxTarget:
			state = "too-fast"
		case value < minTarget:
			state = "too-slow"
		default:
			state = "fairway"
		}

		fmt.Println(state)
	}
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
