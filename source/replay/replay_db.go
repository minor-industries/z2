//go:build !wasm

package replay

import (
	"context"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/source"
	"github.com/pkg/errors"
	"tinygo.org/x/bluetooth"
)

func mustParseUUID(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

func FromDatabase(
	ctx context.Context,
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

	return replay(ctx, rows, callback)
}
