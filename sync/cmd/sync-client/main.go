package main

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/sync"
	"github.com/pkg/errors"
	"os"
	"time"
)

func run() error {
	src, err := sqlite.Get(os.ExpandEnv("$HOME/.z2/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get src db")
	}

	client := sync.NewClient("http://localhost:8080")

	seen := set.Set[time.Time]{}
	err = sync.BucketAll(src, 365, func(day time.Time, ns sync.NamedSeries) error {
		resp, err := client.SendSeries(ns)
		if err != nil {
			return errors.Wrap(err, "send series")
		}

		if !seen.Has(day) {
			fmt.Println(day.Format("2006-01-02"))
		}

		seen.Add(day)
		fmt.Printf("  %s: %d/%d\n", ns.Name, resp.NewItems, len(ns.Timestamps))

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "bucket all")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
