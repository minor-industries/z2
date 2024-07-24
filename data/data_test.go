package data

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

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

func TestData(t *testing.T) {
	errCh := make(chan error)

	db, err := database.Get(os.ExpandEnv("$HOME/.z2/z2.db"), errCh)
	require.NoError(t, err)

	orm := db.GetORM()

	err = orm.AutoMigrate(&Marker{})
	require.NoError(t, err)

	var markers []Marker
	tx := orm.Where("ref = ?", "bike").Order("timestamp asc").Find(&markers)
	require.NoError(t, tx.Error)

	require.True(t, len(markers)%2 == 0)
	for i, marker := range markers {
		switch i % 2 {
		case 0:
			require.Equal(t, "b", marker.Type)
		case 1:
			require.Equal(t, "e", marker.Type)
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

		computeIntervals(t, orm, t0, t1, 2, "bike_instant_speed")
		computeIntervals(t, orm, t0, t1, 1, "heartrate")
	}
}

func computeIntervals(
	t *testing.T,
	orm *gorm.DB,
	t0 time.Time,
	t1 time.Time,
	decimalPlaces int,
	metricName string,
) {
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
		require.NoError(t, tx.Error)

		intervals = append(intervals, fmt.Sprintf("%.*f", decimalPlaces, avg(values)))

		if tn.Equal(t1) {
			break
		}
	}

	fmt.Println("\t" + strings.Join(intervals, "   "))
}
