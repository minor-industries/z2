package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database/inmem"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"syscall/js"
	"time"
)

func run() error {
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

	var err error
	graph, err := rtgraph.New(
		inmem.NewBackend(),
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

	eventInit := js.Global().Get("CustomEvent").New("wasmReady")
	js.Global().Get("document").Call("dispatchEvent", eventInit)

	select {}
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
