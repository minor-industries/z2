package main

import (
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/sync"
	"github.com/pkg/errors"
)

func run() error {
	dstDBPath := "synced.db"

	dst, err := sqlite.Get(dstDBPath)
	if err != nil {
		return errors.Wrap(err, "get destination database")
	}

	panic(sync.RunServer(dst))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
