package main

import (
	"fmt"
	"github.com/minor-industries/backup/restic"
	"github.com/minor-industries/z2/backup"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"os"
)

func run() error {
	opts, err := cfg.Load(cfg.DefaultConfigPath)
	if err != nil {
		return errors.Wrap(err, "load configuration")
	}

	backupPath, err := backup.PrepareForBackup(&opts.Backup)
	if err != nil {
		return errors.Wrap(err, "check config")
	}

	err = restic.Run(&opts.Backup, backupPath, restic.QuantizeFilter(func(msg any) error {
		switch msg := msg.(type) {
		case restic.StartBackup:
			if msg.KeychainProfile != "" {
				fmt.Println("loading keychain profile:", msg.KeychainProfile)
			}
			if msg.Repository != "" {
				fmt.Println("starting backup:", msg.Repository)
			}
		case restic.ResticStatus:
			fmt.Printf("Progress: %.1f%%\n", msg.PercentDone*100)
		case restic.ResticSummary:
			fmt.Println("Backup done!")
		default:
			fmt.Println("Unknown message type")
		}
		return nil
	}))

	return errors.Wrap(err, "backup one")
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
