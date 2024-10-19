//go:build !wasm

package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/bike"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/multi"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/source/rower"
	"github.com/minor-industries/z2/time_series"
	"github.com/minor-industries/z2/variables"
	"github.com/pkg/errors"
	webview "github.com/webview/webview_go"
	"os"
)

func run() error {
	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	if opts.ConfigFile == "" {
		opts.ConfigFile = cfg.DefaultConfigPath
	}

	opts, err := cfg.Load(opts.ConfigFile)
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
	)
	if err != nil {
		return errors.Wrap(err, "new graph")
	}

	vars, err := variables.NewCache(variables.NewSQLiteStorage(samples.GetORM()))
	if err != nil {
		return errors.Wrap(err, "new cache")
	}

	time_series.SetupGraphFunctions(graph, vars)

	br := broker.NewBroker()
	go br.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allSources := map[string]source.Source{
		"bike":  &bike.BikeSource{},
		"rower": rower.NewRowerSource(),
		"hrm":   &heartrate.Source{},
	}

	multiSource := multi.NewSource()
	for _, s := range allSources {
		err := multiSource.Add(s)
		if err != nil {
			return errors.Wrap(err, "add source")
		}
	}

	mainHandler := handler.NewHandler(
		graph,
		backends.RawValues,
		multiSource,
		opts.WriteRawValues,
		cancel,
		ctx,
	)

	sources, err := setupSources(opts.Devices, mainHandler)
	if err != nil {
		return errors.Wrap(err, "setup source")
	}

	disconnect := make(chan struct{})

	router := gin.New()
	err = setupRoutes(router, opts, graph, br, backends, vars, disconnect)
	if err != nil {
		return errors.Wrap(err, "setup routes")
	}

	go func() {
		errCh <- router.Run(fmt.Sprintf("0.0.0.0:%d", opts.Port))
	}()

	go func() {
		if opts.ReplayDB != "" {
			fmt.Println("using replay database", opts.ReplayDB)
			go mainHandler.Monitor(disconnect) // TODO: should monitor each source independently
			errCh <- replay.FromDatabase(
				ctx,
				opts.ReplayDB,
				mainHandler.Handle,
			)
		} else if len(sources.connect) > 0 {
			source.Connect(ctx, errCh, sources.connect, disconnect)
			fmt.Println("all devices found, source.Connect finished")
			go mainHandler.Monitor(disconnect) // TODO: should monitor each source independently
		} else {
			fmt.Println("no sources found")
		}
	}()

	if sources.primary != nil {
		z2App := app.NewApp(graph, vars, br, sources.primaryKind, opts.Audio)
		z2App.Run()
	}

	if opts.Webview {
		if sources.primary == nil {
			return errors.New("unknown primary source")
		}

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
			fmt.Sprintf("http://localhost:%d/z2/pages/%s.html", opts.Port, sources.primaryKind),
		)
		return <-errCh2
	} else {
		return <-errCh
	}
}

type SourceInfo struct {
	primary     source.Source
	primaryKind string
	connect     []source.Device
}

func setupSources(devices []cfg.Device, mainHandler *handler.Handler) (*SourceInfo, error) {
	var primarySources []source.Source
	result := &SourceInfo{}

	for _, dev := range devices {
		if dev.Disable == true {
			continue
		}

		var src source.Source
		switch dev.Kind {
		case "bike":
			src = &bike.BikeSource{}
			result.primaryKind = dev.Kind
			primarySources = append(primarySources, src)
		case "rower":
			src = rower.NewRowerSource()
			result.primaryKind = dev.Kind
			primarySources = append(primarySources, src)
		case "hrm":
			src = &heartrate.Source{}
		default:
			return nil, fmt.Errorf("unknown device kind: %s", dev.Kind)
		}

		result.connect = append(result.connect, source.Device{
			Source:   src,
			Address:  dev.Addr,
			Kind:     dev.Kind,
			Name:     dev.Name,
			Callback: mainHandler.Handle,
		})
	}

	switch len(primarySources) {
	case 0:
		// pass
	case 1:
		result.primary = primarySources[0]
	default:
		return nil, errors.New("found multiple primary devices")
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
	if opts.RemoveDB {
		_ = os.Remove(opts.DBPath)
	}

	db, err := sqlite.Get(opts.DBPath)
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

var opts struct {
	ConfigFile string `long:"config-file" description:"config file path"`
}

func main() {
	err := run()
	fmt.Println("run exited, error:", err)
}
