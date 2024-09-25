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

	for _, backupCfg := range opts.Backups {
		lastQuantum := -1.0
		processor.BackupOne(backupCfg, func(msg any) error {
			switch msg := msg.(type) {
			case backup.ResticStatus:
				currentQuantum := float64(int(msg.PercentDone*10)) / 10.0

				if currentQuantum > lastQuantum {
					lastQuantum = currentQuantum
					fmt.Printf("Progress: %.1f%%\n", msg.PercentDone*100)
				}
			case backup.ResticSummary:
				fmt.Println("Backup done!")
			default:
				fmt.Println("Unknown message type")
			}
			return nil
		})
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
