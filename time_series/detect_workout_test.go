package time_series

import (
	"fmt"
	"github.com/minor-industries/rtgraph/schema"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

type Detector struct {
	active bool
}

func (d Detector) Detect(values []schema.Value) {
	for _, v := range values {
		_ = v
	}
}

func Test_DetectWorkout(t *testing.T) {
	dbPath := os.ExpandEnv("$HOME/.z2/z2.db")
	errCh := make(chan error)

	db, err := database.Get(dbPath, errCh)
	require.NoError(t, err)

	window, err := db.LoadDataWindow("bike_instant_speed", time.Time{})
	require.NoError(t, err)

	d := &Detector{}

	d.Detect(window.Values)

	for i, value := range window.Values {
		fmt.Println(i, value.Timestamp, value.Value)
	}
}
