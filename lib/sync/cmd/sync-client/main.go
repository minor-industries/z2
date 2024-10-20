package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	sync2 "github.com/minor-industries/z2/lib/sync"
	"github.com/pkg/errors"
	"os"
)

func run() error {
	src, err := sqlite.Get(os.ExpandEnv("$HOME/.z2/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get src db")
	}

	client := sync2.NewClient("localhost:8080", "z2-jeremy")

	return sync2.Sync(src, client, 0, func(s string) {
		fmt.Println(s)
	})
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
