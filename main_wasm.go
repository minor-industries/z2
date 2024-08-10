package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database/inmem"
	"github.com/minor-industries/rtgraph/messages"
	"github.com/minor-industries/rtgraph/subscription"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"math/rand"
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

	go func() {
		ticker := time.NewTicker(time.Second)

		for t := range ticker.C {
			err := graph.CreateValue("heartrate", t, 70+2*rand.Float64())
			if err != nil {
				panic(err)
			}
		}
	}()

	select {}
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
