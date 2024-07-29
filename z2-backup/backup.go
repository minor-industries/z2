package main

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func removeIfExist(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "remove existing file")
		}
	}
	return nil
}

func hardLinkFile(src, dest string) error {
	if err := removeIfExist(dest); err != nil {
		return err
	}
	if err := os.Link(src, dest); err != nil {
		return errors.Wrap(err, "create hard link")
	}
	return nil
}

func run() error {
	opts, err := cfg.Load(cfg.DefaultConfigPath)
	if err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if opts.BackupPrefix == "" {
		return errors.New("backup_prefix unset in config file")
	}

	z2Path := os.ExpandEnv("$HOME/.z2")
	dbFile := filepath.Join(z2Path, "z2.db")
	backupPath := filepath.Join(z2Path, "backup")

	prefixed := fmt.Sprintf("z2-backup-%s", opts.BackupPrefix)
	backupFile := filepath.Join(backupPath, prefixed+".db")

	err = os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return errors.Wrap(err, "mkdir backup path")
	}

	db, err := database.Get(dbFile)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "vacuum into")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
