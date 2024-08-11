package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database/capacitor_sqlite"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"syscall/js"
	"time"
)

func run() error {
	fmt.Println("in wasm run()")
	opts := cfg.Config{
		Source:            "bike",
		ReplayDB:          "",
		Port:              0,
		HeartrateMonitors: nil,
		StaticPath:        "",
		RemoveDB:          false,
		Webview:           false,
		XRes:              0,
		YRes:              0,
		Scan:              false,
		Audio:             "",
		WriteRawValues:    false,
		ResticPath:        "",
		BackupHost:        "",
		Backups:           nil,
		Devices:           nil,
	}
	_ = opts
	errCh := make(chan error)

	dbManager := js.Global().Get("dbManager")
	fmt.Println(dbManager.String())

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
		go graph.Subscribe(&subscription.Request{
			Series:      []string{"heartrate"},
			WindowSize:  0,
			LastPointMs: 0,
			Date:        "",
		}, now, msgs)

		callback := args[0]

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

	eventInit := js.Global().Get("CustomEvent").New("wasmReady")
	js.Global().Get("document").Call("dispatchEvent", eventInit)

	select {}
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
