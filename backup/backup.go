package main

import (
	"compress/gzip"
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func gzipFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source file")
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "create destination file")
	}
	defer destFile.Close()

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	if err != nil {
		return errors.Wrap(err, "copy to gzip")
	}

	return nil
}

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
	z2Path := os.ExpandEnv("$HOME/.z2")
	dbPath := filepath.Join(z2Path, "z2.db")

	db, err := database.Get(dbPath)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	backupFile := filepath.Join(z2Path, "z2-backup.db")

	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(backupFile); err != nil {
			return errors.Wrap(err, "remove existing backup file")
		}
	}

	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "vacuum into")
	}

	backupPath := filepath.Join(z2Path, "backup")
	err = os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return errors.Wrap(err, "mkdir backup path")
	}

	gzippedBackupFile := filepath.Join(backupPath, "z2-backup.gz")
	if err := removeIfExist(gzippedBackupFile); err != nil {
		return errors.Wrap(err, "remove existing gzipped backup file")
	}
	if err := gzipFile(backupFile, gzippedBackupFile); err != nil {
		return errors.Wrap(err, "gzip backup file")
	}

	now := time.Now()
	_, week := now.ISOWeek()
	dayOfWeek := strings.ToLower(now.Format("Mon"))

	weekBackupFile := filepath.Join(backupPath, fmt.Sprintf("z2-backup-week-%02d.gz", week))
	dayBackupFile := filepath.Join(backupPath, fmt.Sprintf("z2-backup-%s.gz", dayOfWeek))

	if err := hardLinkFile(gzippedBackupFile, weekBackupFile); err != nil {
		return errors.Wrap(err, "create hard link for week backup")
	}

	if err := hardLinkFile(gzippedBackupFile, dayBackupFile); err != nil {
		return errors.Wrap(err, "create hard link for day backup")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
