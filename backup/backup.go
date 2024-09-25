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

func QuantizeFilter(callback func(msg any) error) func(msg any) error {
	lastQuantum := -1.0

	return func(msg any) error {
		switch msg := msg.(type) {
		case ResticStatus:
			currentQuantum := float64(int(msg.PercentDone*10)) / 10.0
			if currentQuantum > lastQuantum {
				lastQuantum = currentQuantum
				return callback(msg)
			}
			return nil
		default:
			return callback(msg)
		}
	}
}

type Processor struct {
	backupPath string
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
		backupPath: backupPath,
	}, nil
}

func (p *Processor) BackupOne(backupCfg cfg.BackupTarget, callback func(any) error) error {
	//cmd := exec.Command(opts.ResticPath, "init")
	cmd := exec.Command(
		os.ExpandEnv(p.opts.Backup.ResticPath),
		"backup",
		"--json",
		"--host",
		p.opts.Backup.SourceHost,
		".",
	)
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID="+backupCfg.AwsAccessKeyId,
		"AWS_SECRET_ACCESS_KEY="+backupCfg.AwsSecretAccessKey,
		"RESTIC_REPOSITORY="+backupCfg.ResticRepository,
		"RESTIC_PASSWORD="+backupCfg.ResticPassword,
	)
	if backupCfg.CACertPath != "" {
		cmd.Env = append(cmd.Env, "RESTIC_CACERT="+os.ExpandEnv(backupCfg.CACertPath))
	}
	cmd.Dir = p.backupPath

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

	return nil
}
