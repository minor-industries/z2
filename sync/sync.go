package main

import (
	"fmt"
	"github.com/jinzhu/now"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"time"
)

type NamedSeries struct {
	Name       string
	Timestamps []int64
	Values     []float64
}

var namedSampleData = []NamedSeries{
	{
		Name:       "Series A",
		Timestamps: []int64{1627689945, 1627689955, 1627689965},
		Values:     []float64{1.23, 2.34, 3.45},
	},
	{
		Name:       "Series B",
		Timestamps: []int64{1627689975, 1627689985, 1627689995},
		Values:     []float64{4.56, 5.67, 6.78},
	},
	{
		Name:       "Series A",
		Timestamps: []int64{1627689965, 1627690005, 1627690015}, // Overlap with 1627689965, new unique timestamps
		Values:     []float64{7.89, 8.90, 9.01},
	},
}

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

	err = bucketAll(src, 60, func(ns NamedSeries) error {
		fmt.Println(" ", ns.Name, len(ns.Timestamps))
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, s := range namedSampleData {
		count, err := insertSeriesBatchWithTransaction(dst.GetORM(), s)
		if err != nil {
			return errors.Wrap(err, "insert")
		}
		fmt.Println(s.Name, s, count)
	}

	return nil
}

func bucketAll(
	src *sqlite.Backend,
	lookbackDays int,
	callback func(ns NamedSeries) error,
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
		fmt.Println(t0)

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

			if err := callback(ns); err != nil {
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
