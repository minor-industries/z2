package main

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/jinzhu/now"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"time"
)

func insertSeriesBatchWithTransaction(db *gorm.DB, series NamedSeries) (int, error) {
	count := 0

	err := db.Transaction(func(tx *gorm.DB) error {
		for i := range series.Timestamps {
			id := sqlite.HashedID(series.Name)
			row := sqlite.Sample{
				SeriesID:  id,
				Timestamp: series.Timestamps[i],
				Value:     series.Values[i],
			}

			res := tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&row)

			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected > 0 {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}
func run() error {
	dst, err := sqlite.Get("synced.db")
	if err != nil {
		return errors.Wrap(err, "get dst db")
	}

	src, err := sqlite.Get(os.ExpandEnv("$HOME/.z2/z2.db"))
	if err != nil {
		return errors.Wrap(err, "get src db")
	}

	seen := set.Set[time.Time]{}
	err = bucketAll(src, 365, func(day time.Time, ns NamedSeries) error {
		count, err := insertSeriesBatchWithTransaction(dst.GetORM(), ns)
		if err != nil {
			return errors.Wrap(err, "insert batch")
		}

		if !seen.Has(day) {
			fmt.Println(day.Format("2006-01-02"))
		}

		seen.Add(day)
		fmt.Printf("  %s: %d/%d\n", ns.Name, count, len(ns.Timestamps))

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "bucket all")
	}

	return nil
}

func bucketAll(
	src *sqlite.Backend,
	lookbackDays int,
	callback func(day time.Time, ns NamedSeries) error,
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

			ns := NamedSeries{
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

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
