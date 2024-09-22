package main

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/sync"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"os"
	"time"
)

func run() error {
	src, err := sqlite.Get(os.ExpandEnv("$HOME/.z2/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get src db")
	}

	client := sync.NewClient("localhost:8080")

	err = sendSeries(src, client)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	err = sendMarkers(src, client)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	return nil
}

func sendMarkers(src *sqlite.Backend, client *sync.Client) error {
	//TODO: perhaps sendMarkers should also respect lookback days
	var markers []sqlite.Marker

	tx := src.GetORM().Find(&markers)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "load markers")
	}

	mapped := lo.Map(markers, func(m sqlite.Marker, index int) sync.Marker {
		return sync.Marker{
			ID:        m.ID,
			Type:      m.Type,
			Ref:       m.Ref,
			Timestamp: m.Timestamp,
		}
	})

	resp, err := client.SendMarkers(mapped)
	if err != nil {
		return errors.Wrap(err, "send markers")
	}

	fmt.Println("markers")
	fmt.Printf("  %d/%d\n", resp.NewItems, len(mapped))

	return nil
}

func sendSeries(src *sqlite.Backend, client *sync.Client) error {
	seen := set.Set[time.Time]{}
	err := sync.BucketAll(src, 365, func(day time.Time, ns *sync.NamedSeries) error {
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
