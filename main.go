package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/static"
	"github.com/minor-industries/z2/variables"
	"github.com/pkg/errors"
	webview "github.com/webview/webview_go"
	"net/http"
	"os"
)

type Config struct {
	Source            string   `toml:"source"`
	ReplayDB          string   `toml:"replay_db"`
	Port              int      `toml:"port"`
	HeartrateMonitors []string `toml:"heartrate_monitors"`
	StaticPath        string   `toml:"static_path"`
	RemoveDB          bool     `toml:"remove_db"`
	Webview           bool     `toml:"webview"`
}

func run() error {
	opts := Config{
		Port:    8077,
		Webview: true,
	}

	cfgFile := os.ExpandEnv("$HOME/.z2/config.toml")
	if _, err := toml.DecodeFile(cfgFile, &opts); err != nil {
		return errors.Wrap(err, "read config file")
	}

	if opts.Source == "" {
		return errors.New("missing source")
	}

	gin.SetMode(gin.ReleaseMode)

	errCh := make(chan error)

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

	vars := variables.NewCache()
	z2App := app.NewApp(graph, vars, opts.Source)
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

	apiHandler := &ApiServer{db: db, vars: vars}
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

	z2App.Run()

	if opts.Webview {
		w := webview.New(true)
		errCh2 := make(chan error)
		go func() {
			err = <-errCh
			w.Terminate()
			errCh2 <- err
		}()

		runWebview(w, errCh, fmt.Sprintf("http://localhost:%d", opts.Port))
		return <-errCh2
	} else {
		return <-errCh
	}
}

func runWebview(
	w webview.WebView,
	ch chan error,
	url string,
) {
	defer w.Destroy()
	w.SetTitle("z2")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate(url)
	w.Run()
	ch <- nil
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
