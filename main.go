//go:build !wasm

package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/app"
	handler2 "github.com/minor-industries/z2/app/handler"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/bike"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/multi"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/source/rower"
	"github.com/minor-industries/z2/static/dist"
	"github.com/minor-industries/z2/variables"
	"github.com/minor-industries/z2/workouts"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	webview "github.com/webview/webview_go"
	"html/template"
	"net/http"
	"os"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/env_web.js
var envWebJS []byte

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

	samples := backends.Samples.(*sqlite.Backend) // TODO

	go samples.RunWriter(errCh)
	go backends.RawValues.RunWriter(errCh)

	if opts.Scan {
		return source.Scan()
	}

	if opts.ReplayDB != "" {
		// Don't write back raw values if we're replaying raw values
		opts.WriteRawValues = false
	}

	gin.SetMode(gin.ReleaseMode)

	if err := backends.RawValues.GetORM().AutoMigrate(
		&data.RawValue{},
	); err != nil {
		return errors.Wrap(err, "automigrate")
	}

	if err := samples.GetORM().AutoMigrate(
		&data.Variable{},
	); err != nil {
		return errors.Wrap(err, "automigrate")
	}

	graph, err := rtgraph.New(
		backends.Samples,
		errCh,
		rtgraph.Opts{},
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "new graph")
	}

	vars, err := variables.NewCache(variables.NewSQLiteStorage(samples.GetORM()))
	if err != nil {
		return errors.Wrap(err, "new cache")
	}

	br := broker.NewBroker()
	go br.Start()

	router := gin.New()
	router.Use(gin.Recovery())
	skipLogging := []string{"/metrics"}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogging...))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dist/html/index.html")
	})

	graph.SetupServer(router.Group("/rtgraph"))

	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/workouts.html", func(c *gin.Context) {
		ref := c.Query("ref")

		cfg, ok := app.Configs[ref]
		if !ok {
			_ = c.Error(errors.New("unknown config"))
			return
		}

		data, err := workouts.GenerateData(samples.GetORM(), ref, cfg.PaceMetric)
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
		router.Static("/dist", opts.StaticPath)
	} else {
		router.StaticFS("/dist", http.FS(dist.FS))
	}

	router.GET("/env.js", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/javascript", envWebJS)
	})

	apiHandler := handler2.NewApiServer(backends, vars)
	router.Any("/twirp/api.Api/*Method", gin.WrapH(api.NewApiServer(apiHandler, nil)))

	router.POST(
		"/twirp/calendar.Calendar/*Method",
		gin.WrapH(calendar.NewCalendarServer(apiHandler, nil)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sources, err := SetupSources(opts.Devices)
	if err != nil {
		return errors.Wrap(err, "setup source")
	}

	mainHandler := handler.NewHandler(
		graph,
		backends.RawValues,
		sources.multi,
		opts.WriteRawValues,
		cancel,
		ctx,
	)
	go mainHandler.Monitor() // TODO: should monitor each source independently

	go func() {
		errCh <- router.Run(fmt.Sprintf("0.0.0.0:%d", opts.Port))
	}()

	if opts.ReplayDB != "" {
		go func() {
			errCh <- replay.FromDatabase(
				ctx,
				os.ExpandEnv(opts.ReplayDB),
				mainHandler.Handle,
			)
		}()
	} else {
		for i := range sources.connect {
			go func(i int) {
				errCh <- source.Run(
					ctx,
					errCh,
					sources.addrs[i],
					sources.connect[i],
					mainHandler.Handle,
				)
			}(i)
		}
	}

	z2App := app.NewApp(graph, vars, br, sources.primaryKind, opts.Audio)
	z2App.Run()

	if opts.Webview {
		return errors.New("webview not implemented")
		//w := webview.New(true)
		//errCh2 := make(chan error)
		//go func() {
		//	err = <-errCh
		//	w.Terminate()
		//	errCh2 <- err
		//}()
		//
		//runWebview(
		//	opts,
		//	w,
		//	errCh,
		//	fmt.Sprintf("http://localhost:%d/%s.html", opts.Port, opts.Source),
		//)
		//return <-errCh2
	} else {
		return <-errCh
	}
}

type SourceInfo struct {
	primary     source.Source
	primaryKind string
	multi       source.Source
	connect     []source.Source
	addrs       []string
}

func SetupSources(devices []cfg.Device) (*SourceInfo, error) {
	sources := map[string]source.Source{}
	var primarySources []source.Source
	result := &SourceInfo{}

	for _, dev := range devices {
		_, hasHandler := sources[dev.Kind]
		if dev.Disable == true && hasHandler {
			continue
		}

		var handler source.Source
		switch dev.Kind {
		case "bike":
			handler = &bike.BikeSource{}
			result.primaryKind = dev.Kind
			primarySources = append(primarySources, handler)
		case "rower":
			handler = rower.NewRowerSource()
			result.primaryKind = dev.Kind
			primarySources = append(primarySources, handler)
		case "hrm":
			handler = &heartrate.Source{}
		default:
			return nil, fmt.Errorf("unknown device kind: %s", dev.Kind)
		}

		if !hasHandler {
			sources[dev.Kind] = handler
		}

		if dev.Disable {
			continue
		}

		fmt.Printf("looking for %s at address %s\n", dev.Kind, dev.Addr)
		result.connect = append(result.connect, handler)
		result.addrs = append(result.addrs, dev.Addr)
	}

	switch len(primarySources) {
	case 0:
		return nil, errors.New("no primary devices found")
	case 1:
		result.primary = primarySources[0]
	default:
		return nil, errors.New("found multiple primary devices")
	}

	var srcs []source.Source
	for _, s := range sources {
		srcs = append(srcs, s)
	}
	var err error
	result.multi, err = multi.NewSource(srcs)
	if err != nil {
		return nil, errors.Wrap(err, "new multi-source")
	}

	return result, nil
}

func getBackends(opts *cfg.Config) (handler.Backends, error) {
	backends := handler.Backends{}

	var err error

	backends.Samples, err = getBackend(opts)
	if err != nil {
		return handler.Backends{}, errors.Wrap(err, "get backend")
	}

	backends.RawValues, err = getBackend(opts)
	if err != nil {
		return handler.Backends{}, errors.Wrap(err, "get backend")
	}

	return backends, err
}

func getBackend(opts *cfg.Config) (*sqlite.Backend, error) {
	dbPath := os.ExpandEnv(opts.DBPath)

	if opts.RemoveDB {
		_ = os.Remove(dbPath)
	}

	db, err := sqlite.Get(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "get database")
	}

	return db, nil
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
