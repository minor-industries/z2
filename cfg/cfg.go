package cfg

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"os"
)

type Backup struct {
	AwsAccessKeyId     string `toml:"aws_access_key_id"`
	AwsSecretAccessKey string `toml:"aws_secret_access_key"`
	ResticRepository   string `toml:"restic_repository"`
	ResticPassword     string `toml:"restic_password"`
	CACertPath         string `toml:"ca_cert_path"`
}

type Config struct {
	Source            string   `toml:"source"`
	ReplayDB          string   `toml:"replay_db"`
	Port              int      `toml:"port"`
	HeartrateMonitors []string `toml:"heartrate_monitors"`
	StaticPath        string   `toml:"static_path"`
	RemoveDB          bool     `toml:"remove_db"`
	Webview           bool     `toml:"webview"`
	XRes              int      `toml:"xres"`
	YRes              int      `toml:"yres"`
	Scan              bool     `toml:"scan"`
	Audio             string   `toml:"audio"`
	WriteRawValues    bool     `toml:"write_raw_values"`

	BackupPrefix string `toml:"backup_prefix"`

	Backups []Backup `toml:"backups"`

	Devices map[string]string `toml:"devices"`
}

var Default = Config{
	Port:    8077,
	Webview: true,
	XRes:    1132,
	YRes:    700,
	Audio:   "browser",
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
