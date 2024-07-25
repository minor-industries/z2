package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func run() error {
	z2Path := os.ExpandEnv("$HOME/.z2")
	dbPath := filepath.Join(z2Path, "z2.db")

	db, err := database.Get(dbPath)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	backupFile := filepath.Join(z2Path, "z2-backup.db")
	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "vacuum into")
	}

	backupPath := filepath.Join(z2Path, "backup")
	err = os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return errors.Wrap(err, "mkdir backup path")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
