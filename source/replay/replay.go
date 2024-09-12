package replay

import (
	"context"
	"fmt"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/source"
	"github.com/pkg/errors"
	"time"
)

func replay(
	ctx context.Context,
	rows []data.RawValue,
	callback source.MessageCallback,
) error {
	if len(rows) == 0 {
		return errors.New("no rows")
	}

	t0 := time.Now()
	fmt.Println("t0", t0)
	ticker := time.NewTicker(10 * time.Millisecond)
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
				return errors.Wrap(err, "callback")
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
