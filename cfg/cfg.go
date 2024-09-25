package cfg

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"os"
)

type Device struct {
	Addr    string `toml:"addr"`
	Kind    string `toml:"kind"`
	Name    string `toml:"name"`
	Disable bool   `toml:"disable"`
}

type BackupTarget struct {
	AwsAccessKeyId     string `toml:"aws_access_key_id"`
	AwsSecretAccessKey string `toml:"aws_secret_access_key"`
	ResticRepository   string `toml:"restic_repository"`
	ResticPassword     string `toml:"restic_password"`
	CACertPath         string `toml:"ca_cert_path"`
}

type BackupConfig struct {
	ResticPath string         `toml:"restic_path"`
	SourceHost string         `toml:"source_host"`
	Targets    []BackupTarget `toml:"targets"`
}

type SyncServerConfig struct {
	Enable    bool     `toml:"enable"`
	Databases []string `toml:"databases"`
	DBPath    string   `toml:"db_path"`
}

type SyncConfig struct {
	Host     string `toml:"host"`
	Database string `toml:"database"`
	Days     int    `toml:"days"`
}

type Config struct {
	DBPath         string `toml:"db_path"`
	ReplayDB       string `toml:"replay_db"`
	Port           int    `toml:"port"`
	StaticPath     string `toml:"static_path"`
	RemoveDB       bool   `toml:"remove_db"`
	Webview        bool   `toml:"webview"`
	XRes           int    `toml:"xres"`
	YRes           int    `toml:"yres"`
	Scan           bool   `toml:"scan"`
	Audio          string `toml:"audio"`
	WriteRawValues bool   `toml:"write_raw_values"`

	Devices []Device `toml:"devices"`

	Backup BackupConfig `toml:"backup"`

	Sync       SyncConfig       `toml:"sync"`
	SyncServer SyncServerConfig `toml:"sync_server"`
}

var Default = Config{
	DBPath:  "$HOME/.z2/z2.db",
	Port:    8077,
	Webview: true,
	XRes:    1132,
	YRes:    700,
	Audio:   "browser",

	Backup: BackupConfig{
		ResticPath: "restic",
		SourceHost: "",
	},

	Sync: SyncConfig{
		Host:     "",
		Database: "",
		Days:     60,
	},

	SyncServer: SyncServerConfig{
		Enable: false,
	},
}

const (
	DefaultConfigPath = "$HOME/.z2/config.toml"
)

func Load(path string) (*Config, error) {
	opts := Default
	cfgFile := os.ExpandEnv(path)
	if _, err := toml.DecodeFile(cfgFile, &opts); err != nil {
		return nil, errors.Wrap(err, "decode file")
	}
	return &opts, nil
}
