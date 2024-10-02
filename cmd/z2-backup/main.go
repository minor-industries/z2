package main

import (
	"fmt"
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

	processor, err := backup.NewProcessor(opts)
	if err != nil {
		return errors.Wrap(err, "check config")
	}

	for _, backupCfg := range opts.Backup.Targets {
		err := processor.BackupOne(backupCfg, backup.QuantizeFilter(func(msg any) error {
			switch msg := msg.(type) {
			case backup.ResticStatus:
				fmt.Printf("Progress: %.1f%%\n", msg.PercentDone*100)
			case backup.ResticSummary:
				fmt.Println("Backup done!")
			default:
				fmt.Println("Unknown message type")
			}
			return nil
		}))

		if err != nil {
			return errors.Wrap(err, "backup one")
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
