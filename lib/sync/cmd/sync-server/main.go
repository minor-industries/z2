package main

import (
	"github.com/minor-industries/z2/lib/sync"
)

var dbNames = []string{
	"z2-jeremy",
	"z2-risa",
	"z2-jeremy-iphone",
}

func run() error {
	panic(sync.RunServer(dbNames))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
