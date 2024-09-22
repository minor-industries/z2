package main

import (
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/sync"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var dbNames = []string{
	"z2-jeremy",
	"z2-risa",
	"z2-jeremy-iphone",
}

func run() error {

	dbs := lo.Associate(dbNames, func(name string) (string, *sqlite.Backend) {
		dst, _ := sqlite.Get(name + ".db")
		return name, dst
	})

	if lo.Contains(lo.Values(dbs), nil) {
		return errors.New("couldn't open one or more dbs")
	}

	panic(sync.RunServer(dbs))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
