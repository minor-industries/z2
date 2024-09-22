package sync

import (
	"github.com/jinzhu/now"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"time"
)

func BucketAll(
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
