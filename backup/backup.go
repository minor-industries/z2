package main

import (
	"github.com/minor-industries/rtgraph/database"
	"github.com/pkg/errors"
	"os"
)

func main() {
	dbPath := os.ExpandEnv("$HOME/.z2/z2.db")

	errCh := make(chan error)

	db, err := database.Get(dbPath, errCh)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

}
