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
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source file")
	}
	defer srcFile.Close()

	// Create the destination gzip file for writing
	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "create destination file")
	}
	defer destFile.Close()

	// Create a new gzip writer
	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	// Copy the source file into the gzip writer
	_, err = io.Copy(gzipWriter, srcFile)
	if err != nil {
		return errors.Wrap(err, "copy to gzip")
	}

	return nil
}

func copyFile(src, dest string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open source file")
	}
	defer srcFile.Close()

	// Create the destination file for writing
	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "create destination file")
	}
	defer destFile.Close()

	// Copy the content from the source file to the destination file
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrap(err, "copy file")
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

	// Remove the existing backup file if it exists
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

	// Gzip the backup file and place it one folder up from the backups directory
	gzippedBackupFile := filepath.Join(z2Path, "z2-backup.gz")
	if err := gzipFile(backupFile, gzippedBackupFile); err != nil {
		return errors.Wrap(err, "gzip backup file")
	}

	// Get the current week number and day of the week number
	now := time.Now()
	_, week := now.ISOWeek()
	dayOfWeek := strings.ToLower(now.Format("Mon"))

	// Define the paths for the rolling backup files with zero-padded week and lowercase abbreviated day names
	weekBackupFile := filepath.Join(backupPath, fmt.Sprintf("z2-backup-week-%02d.gz", week))
	dayBackupFile := filepath.Join(backupPath, fmt.Sprintf("z2-backup-%s.gz", dayOfWeek))

	// Copy the gzipped backup file to the rolling backup files
	if err := copyFile(gzippedBackupFile, weekBackupFile); err != nil {
		return errors.Wrap(err, "copy week backup")
	}

	if err := copyFile(gzippedBackupFile, dayBackupFile); err != nil {
		return errors.Wrap(err, "copy day backup")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
