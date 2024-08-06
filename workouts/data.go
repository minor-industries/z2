package workouts

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math"
	"time"
)

func avg(values []sqlite.Sample) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	sum := 0.0
	count := 0
	for _, value := range values {
		sum += value.Value
		count++
	}

	return sum / float64(count)
}

const (
	maxCols = 6
)

type Row struct {
	Date string
	Info string
	Data [][maxCols]string
}

func GenerateData(
	orm *gorm.DB,
	ref string,
	paceMetric string,
) ([]Row, error) {
	var result []Row

	err := orm.AutoMigrate(&sqlite.Marker{})
	if err != nil {
		return nil, errors.Wrap(err, "migrate")
	}

	var markers []sqlite.Marker
	tx := orm.Where("ref = ?", ref).Order("timestamp asc").Find(&markers)
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "find")
	}

	if len(markers)%2 != 0 {
		return nil, errors.New("uneven number of markers")
	}

	for i, marker := range markers {
		switch i % 2 {
		case 0:
			if marker.Type != "b" {
				return nil, errors.New("expected b marker")
			}
		case 1:
			if marker.Type != "e" {
				return nil, errors.New("expected e marker")
			}
		}
	}

	for i := 0; i < len(markers); i += 2 {
		b := markers[i]
		e := markers[i+1]

		t0 := time.UnixMilli(b.Timestamp)
		t1 := time.UnixMilli(e.Timestamp)

		dt := t1.Sub(t0)

		row := Row{
			Info: fmt.Sprint(t0, dt),
			Date: t0.Format("2006-01-02"),
		}

		if err := computeIntervals(orm, &row, t0, t1, 2, paceMetric); err != nil {
			return nil, errors.Wrap(err, "compute interval")
		}
		if err := computeIntervals(orm, &row, t0, t1, 1, "heartrate"); err != nil {
			return nil, errors.Wrap(err, "compute interval")
		}

		result = append(result, row)
	}

	return result, nil
}

func computeIntervals(
	orm *gorm.DB,
	row *Row,
	t0 time.Time,
	t1 time.Time,
	decimalPlaces int,
	metricName string,
) error {
	var cols [6]string

	interval := 15 * time.Minute
	for i := 0; i < maxCols; i++ {
		ti := t0.Add(interval * time.Duration(i))
		tn := t0.Add(interval * time.Duration(i+1))

		if tn.After(t1) {
			tn = t1
		}

		sID := sqlite.HashedID(metricName)

		var values []sqlite.Sample
		tx := orm.Where(
			"series_id = ? and timestamp >= ? and timestamp < ?",
			sID,
			ti.UnixMilli(),
			tn.UnixMilli(),
		).Find(&values)
		if tx.Error != nil {
			return errors.Wrap(tx.Error, "find")
		}

		cols[i] = fmt.Sprintf("%.*f", decimalPlaces, avg(values))

		if tn.Equal(t1) {
			break
		}
	}

	row.Data = append(row.Data, cols)

	return nil
}
