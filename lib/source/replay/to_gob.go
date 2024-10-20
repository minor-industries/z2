package replay

import (
	"encoding/gob"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/lib/data"
	"github.com/pkg/errors"
	"log"
	"os"
)

func toGob() error {
	db, err := sqlite.Get(os.ExpandEnv("$HOME/z2-replay.db"))
	if err != nil {
		return errors.Wrap(err, "get replay database")
	}

	var rows []data.RawValue
	tx := db.GetORM().Find(&rows)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "load rows")
	}

	// Create the gob file
	file, err := os.Create("rawValues.gob")
	if err != nil {
		log.Fatal("failed to create gob file:", err)
	}
	defer file.Close()

	// Serialize the data to the gob file
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(rows)
	if err != nil {
		log.Fatal("failed to serialize data:", err)
	}

	return nil
}
