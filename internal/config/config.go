package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
	Trap   TrapConfig   `toml:"trap"`
	Data   DataConfig   `toml:"data"`
	Sync   SyncConfig   `toml:"sync"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type TrapConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type DataConfig struct {
	Dir string `toml:"dir"`
}

type SyncConfig struct {
	PollIntervalSeconds int      `toml:"poll_interval_seconds"`
	DefaultFolders      []string `toml:"default_folders"`
}

// DefaultDir returns ~/.catchy, the root directory for all catchy data.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".catchy"
	}
	return filepath.Join(home, ".catchy")
}

// Load reads the TOML config from path. If path is empty, it looks for
// ~/.catchy/config.toml. A missing config file is not an error — defaults apply.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path == "" {
		path = filepath.Join(DefaultDir(), "config.toml")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

func defaults() *Config {
	dir := DefaultDir()
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8765,
		},
		Trap: TrapConfig{
			Host: "localhost",
			Port: 1025,
		},
		Data: DataConfig{
			Dir: filepath.Join(dir, "data"),
		},
		Sync: SyncConfig{
			PollIntervalSeconds: 60,
			DefaultFolders:      []string{"INBOX", "Sent"},
		},
	}
}

// LoadOrCreateSecretKey loads the 32-byte secret key from dir/secret.key.
// If the file does not exist, a new key is generated and written with 0600 permissions.
func LoadOrCreateSecretKey(dir string) ([]byte, error) {
	path := filepath.Join(dir, "secret.key")

	data, err := os.ReadFile(path)
	if err == nil && len(data) == 32 {
		return data, nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating secret key: %w", err)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating catchy dir: %w", err)
	}

	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("writing secret key: %w", err)
	}

	return key, nil
}
