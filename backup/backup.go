package backup

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

type ResticMessage struct {
	MessageType string `json:"message_type"`
}

type ResticStatus struct {
	MessageType  string   `json:"message_type"`
	PercentDone  float64  `json:"percent_done"`
	TotalFiles   int      `json:"total_files"`
	FilesDone    int      `json:"files_done"`
	TotalBytes   int64    `json:"total_bytes"`
	BytesDone    int64    `json:"bytes_done"`
	CurrentFiles []string `json:"current_files"`
}

type ResticSummary struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           int64   `json:"data_added"`
	DataAddedPacked     int64   `json:"data_added_packed"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed int64   `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
}

func decodeResticMessage(data []byte) (any, error) {
	var shim ResticMessage
	if err := json.Unmarshal(data, &shim); err != nil {
		return nil, err
	}

	switch shim.MessageType {
	case "status":
		var status ResticStatus
		if err := json.Unmarshal(data, &status); err != nil {
			return nil, err
		}
		return status, nil
	case "summary":
		var summary ResticSummary
		if err := json.Unmarshal(data, &summary); err != nil {
			return nil, err
		}
		return summary, nil
	default:
		return nil, errors.New("unknown message type")
	}
}

func Run(callback func(any) error) error {
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

			msg, err := decodeResticMessage(line)
			if err != nil {
				return errors.Wrap(err, "decode restic message")
			}

			if err := callback(msg); err != nil {
				return errors.Wrap(err, "callback returned error")
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
