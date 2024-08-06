package main

import (
	"fmt"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"os"
	"os/exec"
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

	db, err := database.Get(dbFile)
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
		cmd := exec.Command(os.ExpandEnv(opts.ResticPath), "backup", "--host", opts.BackupHost, ".")
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
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "run restic")
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
