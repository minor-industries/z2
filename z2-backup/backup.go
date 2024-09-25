package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
)

func run() error {
	opts, err := cfg.Load(cfg.DefaultConfigPath)
	if err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if opts.BackupHost == "" {
		return errors.New("backup_host unset in config file")
	}

	z2Path := os.ExpandEnv("$HOME/.z2")
	dbFile := filepath.Join(z2Path, "z2.db")
	backupPath := filepath.Join(z2Path, "backup")

	prefixed := fmt.Sprintf("z2-backup-%s", opts.BackupHost)
	backupFile := filepath.Join(backupPath, prefixed+".db")

	err = os.MkdirAll(backupPath, 0o750)
	if err != nil {
		return errors.Wrap(err, "mkdir backup path")
	}

	db, err := sqlite.Get(dbFile)
	if err != nil {
		return errors.Wrap(err, "get database")
	}

	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Remove(backupFile); err != nil {
			return errors.Wrap(err, "remove existing backup file")
		}
	}

	tx := db.GetORM().Exec("VACUUM INTO ?", backupFile)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "vacuum into")
	}

	if len(opts.Backups) == 0 {
		fmt.Println("no backup configs found!")
		return nil
	}

	for _, backupCfg := range opts.Backups {
		//cmd := exec.Command(opts.ResticPath, "init")
		cmd := exec.Command(os.ExpandEnv(opts.ResticPath), "backup", "--json", "--host", opts.BackupHost, ".")
		cmd.Env = append(os.Environ(),
			"AWS_ACCESS_KEY_ID="+backupCfg.AwsAccessKeyId,
			"AWS_SECRET_ACCESS_KEY="+backupCfg.AwsSecretAccessKey,
			"RESTIC_REPOSITORY="+backupCfg.ResticRepository,
			"RESTIC_PASSWORD="+backupCfg.ResticPassword,
		)
		if backupCfg.CACertPath != "" {
			cmd.Env = append(cmd.Env, "RESTIC_CACERT="+os.ExpandEnv(backupCfg.CACertPath))
		}
		cmd.Dir = backupPath

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return errors.Wrap(err, "get stdout pipe")
		}

		if err := cmd.Start(); err != nil {
			return errors.Wrap(err, "start restic")
		}

		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Bytes()
			fmt.Println(string(line))
			var result map[string]interface{}
			if err := json.Unmarshal(line, &result); err != nil {
				fmt.Printf("Failed to parse JSON: %v\n", err)
				continue
			}

			if percentDone, ok := result["percent_done"].(float64); ok {
				fmt.Printf("Progress: %.2f%%\n", percentDone*100)
			} else {
				fmt.Printf("Restic output: %s\n", line)
			}
		}

		if err := scanner.Err(); err != nil {
			return errors.Wrap(err, "read from restic stdout")
		}

		if err := cmd.Wait(); err != nil {
			return errors.Wrap(err, "wait for restic command")
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
