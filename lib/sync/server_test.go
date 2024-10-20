package sync

import (
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestInsertSeriesWithClient(t *testing.T) {
	srcDBPath := os.ExpandEnv("$HOME/.z2/z2.db")
	dstDBPath := "synced.db"

	src, err := sqlite.Get(srcDBPath)
	if err != nil {
		t.Fatalf("failed to get src database: %v", err)
	}

	dst, err := sqlite.Get(dstDBPath)
	if err != nil {
		t.Fatalf("failed to get dst database: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		if err := RunServer(nil); err != nil {
			t.Fatalf("failed to run server: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)

	client := NewClient("localhost:8080")

	seen := set.Set[time.Time]{}
	err = BucketAll(src, 365, func(day time.Time, ns NamedSeries) error {
		resp, err := client.SendSeries(ns)
		if err != nil {
			return err
		}

		if !seen.Has(day) {
			t.Logf("Processing day: %s", day.Format("2006-01-02"))
		}

		seen.Add(day)
		t.Logf("Series: %s, New Items: %d, Total Timestamps: %d", ns.Name, resp.NewItems, len(ns.Timestamps))

		return nil
	})
	if err != nil {
		t.Fatalf("failed during bucketAll: %v", err)
	}

	var count int64
	dst.GetORM().Model(&sqlite.Sample{}).Count(&count)

	assert.Greater(t, count, int64(0), "Expected at least one row to be inserted")
}
