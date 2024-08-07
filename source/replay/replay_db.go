//go:build !wasm

package replay

import (
	"context"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/source"
	"github.com/pkg/errors"
	"time"
	"tinygo.org/x/bluetooth"
)

func mustParseUUID(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

func RunDB(
	ctx context.Context,
	errCh chan error,
	filename string,
	callback source.MessageCallback,
) error {
	db, err := sqlite.Get(filename)
	if err != nil {
		return errors.Wrap(err, "get replay database")
	}

	var rows []data.RawValue
	tx := db.GetORM().Find(&rows)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "load rows")
	}

	if len(rows) == 0 {
		return errors.New("no rows")
	}

	t0 := time.Now()
	ticker := time.NewTicker(10 * time.Millisecond)
	//
	tFirst := rows[0].Timestamp

	for {
		select {
		case now := <-ticker.C:
			dt := now.Sub(t0)
			t := tFirst.Add(dt)
			row := rows[0]
			if t.Before(row.Timestamp) {
				continue
			}
			if err := callback(
				now,
				source.UUID(row.ServiceID),
				source.UUID(row.CharacteristicID),
				row.Message,
			); err != nil {
				// something seems "off" here
				errCh <- errors.Wrap(err, "callback")
				return nil
			}
			rows = rows[1:]
			if len(rows) == 0 {
				ticker.Stop()
				break
			}
		case <-ctx.Done():
			return nil
		}
	}
}
