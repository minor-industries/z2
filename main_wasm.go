package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/database/inmem"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
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

	fmt.Println(opts)

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

	fmt.Println(graph)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}
