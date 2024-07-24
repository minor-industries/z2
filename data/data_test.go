package data

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func TestData(t *testing.T) {
	require.NoError(t, generateData())
}

func avg(values []database.Value) float64 {
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

func generateData() error {
	errCh := make(chan error)

	db, err := database.Get(os.ExpandEnv("$HOME/.z2/z2.db"), errCh)
	if err != nil {
		return errors.Wrap(err, "get db")
	}

	orm := db.GetORM()

	err = orm.AutoMigrate(&Marker{})
	if err != nil {
		return errors.Wrap(err, "migrate")
	}

	var markers []Marker
	tx := orm.Where("ref = ?", "bike").Order("timestamp asc").Find(&markers)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "find")
	}

	if len(markers)%2 != 0 {
		return errors.New("uneven number of markers")
	}

	for i, marker := range markers {
		switch i % 2 {
		case 0:
			if marker.Type != "b" {
				return errors.New("expected b marker")
			}
		case 1:
			if marker.Type != "e" {
				return errors.New("expected e marker")
			}
		}
	}

	for i := 0; i < len(markers); i += 2 {
		b := markers[i]
		e := markers[i+1]

		t0 := time.UnixMilli(b.Timestamp)
		t1 := time.UnixMilli(e.Timestamp)

		dt := t1.Sub(t0)
		fmt.Println(
			t0,
			dt,
		)

		if err := computeIntervals(orm, t0, t1, 2, "bike_instant_speed"); err != nil {
			return errors.Wrap(err, "compute interval")
		}
		if err := computeIntervals(orm, t0, t1, 1, "heartrate"); err != nil {
			return errors.Wrap(err, "compute interval")
		}
	}

	return nil
}

func computeIntervals(
	orm *gorm.DB,
	t0 time.Time,
	t1 time.Time,
	decimalPlaces int,
	metricName string,
) error {
	var intervals []string
	interval := 15 * time.Minute
	for i := 0; ; i++ {
		ti := t0.Add(interval * time.Duration(i))
		tn := t0.Add(interval * time.Duration(i+1))

		if tn.After(t1) {
			tn = t1
		}

		sID := database.HashedID(metricName)

		var values []database.Value
		tx := orm.Where(
			"series_id = ? and timestamp >= ? and timestamp < ?",
			sID,
			ti.UnixMilli(),
			tn.UnixMilli(),
		).Find(&values)
		if tx.Error != nil {
			return errors.Wrap(tx.Error, "find")
		}

		intervals = append(intervals, fmt.Sprintf("%.*f", decimalPlaces, avg(values)))

		if tn.Equal(t1) {
			break
		}
	}

	fmt.Println("\t" + strings.Join(intervals, "   "))
	return nil
}
