package sync

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/jinzhu/now"
	"github.com/minor-industries/rtgraph/storage"
	"github.com/pkg/errors"
	"time"
)

func BucketAll(
	src storage.StorageBackend,
	lookbackDays int,
	callback func(day time.Time, ns *NamedSeries) error,
) error {
	seriesNames, err := src.AllSeriesNames()
	if err != nil {
		return errors.Wrap(err, "get series names")
	}

	t0 := now.With(time.Now()).BeginningOfDay()

	for i := 0; i < lookbackDays; i++ {
		t1 := t0.AddDate(0, 0, 1)

		for _, series := range seriesNames {
			samples, err := src.LoadDataBetween(series, t0, t1)
			if err != nil {
				return errors.Wrap(err, "load data between")
			}
			if len(samples.Values) == 0 {
				continue
			}

			ns := &NamedSeries{
				Name:       series,
				Timestamps: make([]int64, len(samples.Values)),
				Values:     make([]float64, len(samples.Values)),
			}

			for i, s := range samples.Values {
				ns.Timestamps[i] = s.Timestamp.UnixMilli()
				ns.Values[i] = s.Value
			}

			if err := callback(t0, ns); err != nil {
				return errors.Wrap(err, "callback")
			}
		}
		t0 = t0.AddDate(0, 0, -1)
	}

	return nil
}

func Sync(src storage.StorageBackend, client *Client, info func(string)) error {
	info("staring sync")

	err := sendSeries(src, client, info)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	err = sendMarkers(src, client, info)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	info("sync complete")

	return nil
}

func sendSeries(src storage.StorageBackend, client *Client, info func(string)) error {
	seen := set.Set[time.Time]{}
	err := BucketAll(src, 365, func(day time.Time, ns *NamedSeries) error {
		resp, err := client.SendSeries(ns)
		if err != nil {
			return errors.Wrap(err, "send series")
		}

		if !seen.Has(day) {
			info(day.Format("2006-01-02"))
		}

		seen.Add(day)
		info(fmt.Sprintf("  %s: %d/%d", ns.Name, resp.NewItems, len(ns.Timestamps)))

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "bucket all")
	}
	return nil
}
