package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/bike"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/source/rower"
	"github.com/minor-industries/z2/static"
	"github.com/minor-industries/z2/variables"
	"github.com/minor-industries/z2/workouts"
	"github.com/pkg/errors"
	webview "github.com/webview/webview_go"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed templates/*.html
var templatesFS embed.FS

func run() error {
	opts, err := cfg.Load(cfg.DefaultConfigPath)
	if err != nil {
		return errors.Wrap(err, "load config file")
	}

	errCh := make(chan error)

	backends, err := getBackends(opts)
	if err != nil {
		return errors.Wrap(err, "get backends")
	}

	go backends.Samples.RunWriter(errCh)
	go backends.RawValues.RunWriter(errCh)

	if opts.Scan {
		return source.Scan()
	}

	if opts.Source == "" {
		// temporary hack for supporting desktop launchers
		switch filepath.Base(os.Args[0]) {
		case "z2-bike":
			opts.Source = "bike"
		case "z2-rower":
			opts.Source = "rower"
		default:
			return errors.New("missing source")
		}
	}

	gin.SetMode(gin.ReleaseMode)

	if err := backends.RawValues.GetORM().AutoMigrate(
		&data.RawValue{},
	); err != nil {
		return errors.Wrap(err, "automigrate")
	}

	if err := backends.Samples.GetORM().AutoMigrate(
		&data.Variable{},
	); err != nil {
		return errors.Wrap(err, "automigrate")
	}

	graph, err := rtgraph.New(
		backends.Samples,
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

	vars, err := variables.NewCache(backends.Samples)
	if err != nil {
		return errors.Wrap(err, "new cache")
	}

	br := broker.NewBroker()
	go br.Start()

	z2App := app.NewApp(graph, vars, br, opts.Source, opts.Audio)
	router := graph.GetEngine()

	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/bike.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "bike.html", gin.H{
			"Title": "Bike",
		})
	})

	router.GET("/rower.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "rower.html", gin.H{
			"Title": "Rower",
		})
	})

	router.GET("/data-bike.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "data-bike.html", gin.H{
			"Title": "Bike",
		})
	})

	router.GET("/data-rower.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "data-rower.html", gin.H{
			"Title": "Rower",
		})
	})

	router.GET("/bike-presets.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "bike-presets.html", gin.H{
			"Title": "Bike",
		})
	})

	router.GET("/workouts.html", func(c *gin.Context) {
		ref := c.Query("ref")

		cfg, ok := app.Configs[ref]
		if !ok {
			_ = c.Error(errors.New("unknown config"))
			return
		}

		data, err := workouts.GenerateData(backends.Samples.GetORM(), ref, cfg.PaceMetric)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.HTML(http.StatusOK, "workouts.html", gin.H{
			"Title": ref + " workouts",
			"Data":  data,
			"Ref":   ref,
		})
	})

	setupSse(br, router)

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

	apiHandler := app.NewApiServer(backends, vars)
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

	router.POST(
		"/twirp/calendar.Calendar/*Method",
		gin.WrapH(calendar.NewCalendarServer(apiHandler, nil)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//src := &mainHandler.BikeSource{}
	srcAddr, src, err := getSource(opts)
	if err != nil {
		return errors.Wrap(err, "get source")
	}
	fmt.Printf("looking for %s at address %s\n", opts.Source, srcAddr)

	mainHandler, err := handler.NewBikeHandler(
		graph,
		backends,
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
				backends,
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

		runWebview(
			opts,
			w,
			errCh,
			fmt.Sprintf("http://localhost:%d/%s.html", opts.Port, opts.Source),
		)
		return <-errCh2
	} else {
		return <-errCh
	}
}

func getBackends(opts *cfg.Config) (handler.Backends, error) {
	backends := handler.Backends{}

	var err error

	backends.Samples, err = getBackend(opts, "$HOME/.z2/z2.db")
	if err != nil {
		return handler.Backends{}, errors.Wrap(err, "get backend")
	}

	backends.RawValues, err = getBackend(opts, "$HOME/.z2/z2-bt.db")
	if err != nil {
		return handler.Backends{}, errors.Wrap(err, "get backend")
	}

	return backends, err
}

func getBackend(opts *cfg.Config, path string) (*database.Backend, error) {
	dbPath := os.ExpandEnv(path)

	if opts.RemoveDB {
		_ = os.Remove(dbPath)
	}

	db, err := database.Get(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "get database")
	}

	return db, nil
}

func getSource(opts *cfg.Config) (string, source.Source, error) {
	switch opts.Source {
	case "bike", "rower":
	default:
		return "", nil, errors.Errorf("unknown source: %s", opts.Source)
	}

	hwid, ok := opts.Devices[opts.Source]
	if !ok {
		return "", nil, errors.Errorf("no device hardware ID specified for [%s]", opts.Source)
	}

	switch opts.Source {
	case "bike":
		return hwid, &bike.BikeSource{}, nil
	case "rower":
		return hwid, rower.NewRowerSource(), nil
	default:
		panic(fmt.Errorf("unknown source: %s", opts.Source))
	}
}

func runWebview(
	opts *cfg.Config,
	w webview.WebView,
	ch chan error,
	url string,
) {
	defer w.Destroy()
	w.SetTitle("z2")
	w.SetSize(opts.XRes, opts.YRes, webview.HintNone)
	w.Navigate(url)
	w.Run()
	ch <- nil
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
