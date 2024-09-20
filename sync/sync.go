package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	z2db, err := sqlite.Get("synced.db")
	if err != nil {
		return errors.Wrap(err, "get db")
	}

	for _, s := range namedSampleData {
		count, err := insertSeriesBatchWithTransaction(z2db.GetORM(), s)
		if err != nil {
			return errors.Wrap(err, "insert")
		}
		fmt.Println(s.Name, s, count)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
