package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/broker"
	"github.com/minor-industries/rtgraph/database/capacitor_sqlite"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/app/handler"
	"github.com/minor-industries/z2/cfg"
	handler2 "github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"github.com/minor-industries/z2/source/bike"
	"github.com/minor-industries/z2/source/heartrate"
	"github.com/minor-industries/z2/source/multi"
	"github.com/minor-industries/z2/source/replay"
	"github.com/minor-industries/z2/source/rower"
	"github.com/minor-industries/z2/time_series"
	"github.com/minor-industries/z2/variables"
	"github.com/minor-industries/z2/wasm"
	"github.com/pkg/errors"
	"syscall/js"
	"time"
)

func run() error {
	fmt.Println("in wasm run()")
	opts := cfg.Config{
		Devices: []cfg.Device{
			{Kind: "bike"},
		},
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
		ResticPath:     "",
		BackupHost:     "",
		Backups:        nil,
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
		[]string{},
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

	btHandler := handler2.NewHandler(
		graph,
		nil,
		multiSource,
		opts.WriteRawValues,
		cancel,
		ctx,
	)

	wasm.HandleBTMsg(btHandler)

	js.Global().Set("createValue", js.FuncOf(func(this js.Value, args []js.Value) any {
		seriesName := args[0].String()
		ts := time.UnixMilli(int64(args[1].Int()))
		value := args[2].Float()

		if err := graph.CreateValue(seriesName, ts, value); err != nil {
			panic(errors.Wrap(err, "create value"))
		}

		return js.Undefined()
	}))

	js.Global().Set("subscribe", js.FuncOf(func(this js.Value, args []js.Value) any {
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
	}))

	apiHandler := handler.NewApiServer(handler2.Backends{
		Samples: db,
	}, vars)

	goWasmApi := wasm.NewApiWasm(apiHandler)
	jsWasmApi := map[string]any{
		"addMarker":       js.FuncOf(goWasmApi.AddMarker),
		"deleteRange":     js.FuncOf(goWasmApi.DeleteRange),
		"updateVariables": js.FuncOf(goWasmApi.UpdateVariables),
		"readVariables":   js.FuncOf(goWasmApi.ReadVariables),
		"loadMarkers":     js.FuncOf(goWasmApi.LoadMarkers),
	}
	js.Global().Set("goWasmApi", jsWasmApi)

	goWasmCalendar := wasm.NewCalendarWasm(apiHandler)
	jsWasmCalendar := map[string]any{
		"getEvents": js.FuncOf(goWasmCalendar.GetEvents),
	}
	js.Global().Set("goWasmCalendar", jsWasmCalendar)

	eventInit := js.Global().Get("CustomEvent").New("wasmReady")
	js.Global().Get("document").Call("dispatchEvent", eventInit)

	br := broker.NewBroker()
	go br.Start()

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
		},
	})

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
