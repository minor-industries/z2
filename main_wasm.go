package main

import (
	"context"
	"encoding/json"
	"fmt"
	backupCfg "github.com/minor-industries/backup/cfg"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/capacitor_sqlite"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/app/time_series"
	"github.com/minor-industries/z2/cfg"
	"github.com/minor-industries/z2/lib/source"
	"github.com/minor-industries/z2/lib/source/bike"
	"github.com/minor-industries/z2/lib/source/heartrate"
	"github.com/minor-industries/z2/lib/source/multi"
	"github.com/minor-industries/z2/lib/source/replay"
	"github.com/minor-industries/z2/lib/source/rower"
	"github.com/minor-industries/z2/lib/sync"
	"github.com/minor-industries/z2/lib/variables"
	"github.com/minor-industries/z2/server/api"
	"github.com/minor-industries/z2/server/wasm"
	"github.com/pkg/errors"
	"syscall/js"
	"time"
)

func run() error {
	fmt.Println("in wasm run()")
	opts := cfg.Config{
		DBPath:         "",
		ReplayDB:       "",
		Port:           0,
		StaticPath:     "",
		RemoveDB:       false,
		Webview:        false,
		XRes:           0,
		YRes:           0,
		Scan:           false,
		Audio:          "",
		WriteRawValues: false,
		Devices: []cfg.Device{
			{Kind: "bike"},
		},
		Backup: backupCfg.BackupConfig{
			ResticPath: "",
			SourceHost: "",
			Targets:    nil,
		},
		Sync: cfg.SyncConfig{
			Host:     "",
			Database: "",
			Days:     0,
		},
		SyncServer: cfg.SyncServerConfig{
			Enable:    false,
			Databases: nil,
		},
	}
	_ = opts
	errCh := make(chan error)

	dbManager := js.Global().Get("dbManager")

	db, err := capacitor_sqlite.NewDatabaseManagerWrapper(dbManager)
	if err != nil {
		return errors.Wrap(err, "new db manager")
	}

	graph, err := rtgraph.New(
		db,
		errCh,
		rtgraph.Opts{},
	)
	if err != nil {
		return errors.Wrap(err, "new rtgraph")
	}

	vars, err := variables.NewCache(&variables.NullStorage{})
	noErr(err)

	time_series.SetupGraphFunctions(graph, vars)

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

	btHandler := app.NewBTHandler(graph, nil, nil, multiSource, opts.WriteRawValues, cancel, ctx)

	br := broker.NewBroker()
	go br.Start()

	apiHandler := api.NewApiServer(
		app.Backends{
			Samples: db,
		},
		vars,
		make(chan struct{}), // TODO: handle disconnect in wasm builds
	)

	goWasmApi := wasm.NewApiWasm(apiHandler)
	jsWasmApi := map[string]any{
		"addMarker":       js.FuncOf(goWasmApi.AddMarker),
		"deleteRange":     js.FuncOf(goWasmApi.DeleteRange),
		"updateVariables": js.FuncOf(goWasmApi.UpdateVariables),
		"readVariables":   js.FuncOf(goWasmApi.ReadVariables),
		"loadMarkers":     js.FuncOf(goWasmApi.LoadMarkers),
	}

	goWasmCalendar := wasm.NewCalendarWasm(apiHandler)
	jsWasmCalendar := map[string]any{
		"getEvents": js.FuncOf(goWasmCalendar.GetEvents),
	}

	printErr := func(err error) {
		js.Global().Get("console").Call("error", err.Error())
	}

	js.Global().Set("z2GoWasm", map[string]any{
		"apiClient":      jsWasmApi,
		"calendarClient": jsWasmCalendar,

		"z2": map[string]any{
			"startReplay": js.FuncOf(func(this js.Value, args []js.Value) any {
				kind := args[0].String()
				go func() {
					// should we protect this with a sync.Once?
					fmt.Println("start replay:", kind)
					err := replay.FromFile(
						ctx,
						fmt.Sprintf("%s.gob", kind),
						btHandler.Handle,
					)
					if err != nil {
						printErr(errors.Wrap(err, "start replay"))
					}
				}()
				return js.Undefined()
			}),

			"startApp": js.FuncOf(func(this js.Value, args []js.Value) any {
				go func() {
					// should we protect this with a sync.Once?
					z2App := app.NewApp(graph, vars, br, "bike", "browser")
					z2App.Run()
				}()
				return js.Undefined()
			}),

			"streamEvents": js.FuncOf(func(this js.Value, args []js.Value) any {
				path := args[0].String()
				callback := args[1]
				go func() {
					// TODO: should we share code with the gin handler?
					switch path {
					case "/events", "events":
						ch := br.Subscribe()
						defer br.Unsubscribe(ch)
						for m := range ch {
							switch msg := m.(type) {
							case *app.PlaySound:
								callback.Invoke(msg.Sound)
							}
						}
					default:
						printErr(fmt.Errorf("unknown events path: %s", path))
					}
				}()
				return js.Undefined()
			}),

			"subscribe": js.FuncOf(func(this js.Value, args []js.Value) any {
				now := time.Now()
				msgs := make(chan *messages.Data)

				req := subscription.Request{}
				err := json.Unmarshal([]byte(args[0].String()), &req)
				if err != nil {
					panic(errors.Wrap(err, "unmarshal"))
				}

				go graph.Subscribe(&req, now, msgs)

				callback := args[1]

				go func() {
					for data := range msgs {
						binmsg, err := data.MarshalMsg(nil)
						if err != nil {
							panic(err)
						}

						uint8Array := js.Global().Get("Uint8Array").New(len(binmsg))
						js.CopyBytesToJS(uint8Array, binmsg)

						callback.Invoke(uint8Array)
					}
				}()

				return js.Null()
			}),

			"createValue": js.FuncOf(func(this js.Value, args []js.Value) any {
				seriesName := args[0].String()
				ts := time.UnixMilli(int64(args[1].Int()))
				value := args[2].Float()

				if err := graph.CreateValue(seriesName, ts, value); err != nil {
					panic(errors.Wrap(err, "create value"))
				}

				return js.Undefined()
			}),

			"handleBTMsg": wasm.HandleBTMsg(btHandler),

			"triggerSync": js.FuncOf(func(this js.Value, args []js.Value) any {
				host := args[0].String()
				lookbackDays := args[1].Int()
				database := args[2].String()
				logCallback := args[3]

				go func() {
					syncClient := sync.NewClient(host, database)
					err := sync.Sync(db, syncClient, lookbackDays, func(s string) {
						fmt.Println("sync:", s)
						logCallback.Invoke(js.ValueOf(s))
					})
					if err != nil {
						printErr(errors.Wrap(err, "sync"))
					}
				}()
				return js.Undefined()
			}),
		},
	})

	eventInit := js.Global().Get("CustomEvent").New("wasmReady")
	js.Global().Get("document").Call("dispatchEvent", eventInit)

	for range time.NewTicker(time.Minute).C {
	}

	return nil // should never get here
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
