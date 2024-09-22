package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/sync"
	"github.com/pkg/errors"
	"os"
)

func run() error {
	src, err := sqlite.Get(os.ExpandEnv("$HOME/.z2/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get src db")
	}

	client := sync.NewClient("localhost:8080", "z2-jeremy")

	return sync.Sync(src, client, func(s string) {
		fmt.Println(s)
	})
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
