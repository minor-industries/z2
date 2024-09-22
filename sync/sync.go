//go:build !wasm

package sync

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/jinzhu/now"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"time"
)

func BucketAll(
	src *sqlite.Backend,
	lookbackDays int,
	callback func(day time.Time, ns *NamedSeries) error,
) error {
	orm := src.GetORM()
	var seriesNames []string
	tx := orm.Model(&sqlite.Series{}).Distinct("name").Pluck("name", &seriesNames)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "get distinct series names")
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

func Sync(src *sqlite.Backend, client *Client, info func(string)) error {
	err := sendSeries(src, client, info)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	err = sendMarkers(src, client, info)
	if err != nil {
		return errors.Wrap(err, "send series")
	}

	return nil
}

func sendMarkers(src *sqlite.Backend, client *Client, info func(string)) error {
	//TODO: perhaps sendMarkers should also respect lookback days
	var markers []sqlite.Marker

	tx := src.GetORM().Find(&markers)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "load markers")
	}

	mapped := lo.Map(markers, func(m sqlite.Marker, index int) Marker {
		return Marker{
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

	info("markers")
	info(fmt.Sprintf("  %d/%d", resp.NewItems, len(mapped)))

	return nil
}

func sendSeries(src *sqlite.Backend, client *Client, info func(string)) error {
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
