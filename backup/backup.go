package backup

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type Processor struct {
	BackupPath string
	opts       *cfg.Config
}

func NewProcessor(opts *cfg.Config) (*Processor, error) {
	if opts.Backup.SourceHost == "" {
		return nil, errors.New("backup_host unset in config file")
	}

	if len(opts.Backup.Targets) == 0 {
		return nil, errors.New("no backup configs found!")
	}

	z2Path := os.ExpandEnv("$HOME/.z2")
	dbFile := filepath.Join(z2Path, "z2.db")
	backupPath := filepath.Join(z2Path, "backup")

	prefixed := fmt.Sprintf("z2-backup-%s", opts.Backup.SourceHost)
	backupFile := filepath.Join(backupPath, prefixed+".db")

	err := os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return nil, errors.Wrap(err, "mkdir backup path")
	}

	db, err := sqlite.Get(dbFile)
	if err != nil {
		return nil, errors.Wrap(err, "get database")
	}

	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(backupFile); err != nil {
			return nil, errors.Wrap(err, "remove existing backup file")
		}
	}

	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "vacuum into")
	}

	return &Processor{
		opts:       opts,
		BackupPath: backupPath,
	}, nil
}
