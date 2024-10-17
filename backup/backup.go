package backup

import (
	"fmt"
	cfg2 "github.com/minor-industries/backup/cfg"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

// PrepareForBackup dumps the database to a directory and returns this path
func PrepareForBackup(opts *cfg2.BackupConfig) (string, error) {
	if opts.SourceHost == "" {
		return "", errors.New("backup_host unset in config file")
	}

	if len(opts.Targets)+len(opts.KeychainProfiles) == 0 {
		return "", errors.New("no backup configs found")
	}

	z2Path := os.ExpandEnv("$HOME/.z2")
	dbFile := filepath.Join(z2Path, "z2.db")
	backupPath := filepath.Join(z2Path, "backup")

	prefixed := fmt.Sprintf("z2-backup-%s", opts.SourceHost)
	backupFile := filepath.Join(backupPath, prefixed+".db")

	err := os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return "", errors.Wrap(err, "mkdir backup path")
	}

	db, err := sqlite.Get(dbFile)
	if err != nil {
		return "", errors.Wrap(err, "get database")
	}

	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(backupFile); err != nil {
			return "", errors.Wrap(err, "remove existing backup file")
		}
	}

	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return "", errors.Wrap(tx.Error, "vacuum into")
	}

	return backupPath, nil
}
