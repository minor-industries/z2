//go:build !wasm

package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
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
	"github.com/samber/lo"
	webview "github.com/webview/webview_go"
	"os"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/env_gin.js
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

	sources, err := setupSources(opts.Devices)
	if err != nil {
		return errors.Wrap(err, "setup source")
	}

	mainHandler := handler.NewHandler(
		graph,
		backends.RawValues,
		multiSource,
		opts.WriteRawValues,
		cancel,
		ctx,
	)
	go mainHandler.Monitor() // TODO: should monitor each source independently

	router := gin.New()
	setupRoutes(router, opts, graph, br, backends, vars)

	go func() {
		errCh <- router.Run(fmt.Sprintf("0.0.0.0:%d", opts.Port))
	}()

	go func() {
		if opts.ReplayDB != "" {

			errCh <- replay.FromDatabase(
				ctx,
				os.ExpandEnv(opts.ReplayDB),
				mainHandler.Handle,
			)
		} else {
			devices := lo.Map(sources.connect, func(src source.Source, index int) source.Device {
				return source.Device{
					Address:  sources.addrs[index],
					Source:   src,
					Callback: mainHandler.Handle,
				}
			})

			source.Run(ctx, errCh, devices)
			fmt.Println("all devices found, source.Run exited")
		}
	}()

	if sources.primary != nil {
		z2App := app.NewApp(graph, vars, br, sources.primaryKind, opts.Audio)
		z2App.Run()
	}

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
	connect     []source.Source
	addrs       []string
}

func setupSources(devices []cfg.Device) (*SourceInfo, error) {
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

		fmt.Printf("looking for %s at address %s\n", dev.Kind, dev.Addr)
		result.connect = append(result.connect, src)
		result.addrs = append(result.addrs, dev.Addr)
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
