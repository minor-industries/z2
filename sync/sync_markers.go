//go:build !wasm

package sync

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/rtgraph/storage"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func sendMarkers(src storage.StorageBackend, client *Client, info func(string)) error {
	//TODO: perhaps sendMarkers should also respect lookback days
	var markers []sqlite.Marker

	db := src.(*sqlite.Backend)

	tx := db.GetORM().Find(&markers)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "load markers")
	}

	mapped := lo.Map(markers, func(m sqlite.Marker, index int) Marker {
		return Marker{
			ID:        m.ID,
			Type:      m.Type,
			Ref:       m.Ref,
			Timestamp: m.Timestamp,
		}
	})

	resp, err := client.SendMarkers(mapped)
	if err != nil {
		return errors.Wrap(err, "send markers")
	}

	info("markers")
	info(fmt.Sprintf("  %d/%d", resp.NewItems, len(mapped)))

	return nil
}
