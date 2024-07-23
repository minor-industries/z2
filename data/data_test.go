package data

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

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

		interval := 15 * time.Minute
		for i := 0; ; i++ {
			t := t0.Add(interval * time.Duration(i))
			tn := t0.Add(interval * time.Duration(i+1))

			if tn.After(t1) {
				tn = t1
			}

			fmt.Println(" interval", i, t, tn.Sub(t))

			if tn.Equal(t1) {
				break
			}
		}
	}
}
