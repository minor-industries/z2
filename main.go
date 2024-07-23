package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/app"
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
	"github.com/pkg/errors"
	webview "github.com/webview/webview_go"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	Source            string   `toml:"source"`
	ReplayDB          string   `toml:"replay_db"`
	Port              int      `toml:"port"`
	HeartrateMonitors []string `toml:"heartrate_monitors"`
	StaticPath        string   `toml:"static_path"`
	RemoveDB          bool     `toml:"remove_db"`
	Webview           bool     `toml:"webview"`
	XRes              int      `toml:"xres"`
	YRes              int      `toml:"yres"`
	Scan              bool     `toml:"scan"`
	Audio             string   `toml:"audio"`

	Devices map[string]string `toml:"devices"`
}

//go:embed templates/*.html
var templatesFS embed.FS

func run() error {
	opts := Config{
		Port:    8077,
		Webview: true,
		XRes:    1132,
		YRes:    700,
		Audio:   "browser",
	}

	cfgFile := os.ExpandEnv("$HOME/.z2/config.toml")
	if _, err := toml.DecodeFile(cfgFile, &opts); err != nil {
		return errors.Wrap(err, "read config file")
	}

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

	errCh := make(chan error)

	dbPath := os.ExpandEnv("$HOME/.z2/z2.db")

	if opts.RemoveDB {
		_ = os.Remove(dbPath)
	}

	db, err := database.Get(dbPath, errCh)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	if err := db.GetORM().AutoMigrate(
		&data.Variable{},
		&data.RawValue{},
	); err != nil {
		return errors.Wrap(err, "automigrate")
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

	vars, err := variables.NewCache(db)
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

	apiHandler := app.NewApiServer(db, vars)
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

		runWebview(
			&opts,
			w,
			errCh,
			fmt.Sprintf("http://localhost:%d/%s.html", opts.Port, opts.Source),
		)
		return <-errCh2
	} else {
		return <-errCh
	}
}

func getSource(opts Config) (string, source.Source, error) {
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
	opts *Config,
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
