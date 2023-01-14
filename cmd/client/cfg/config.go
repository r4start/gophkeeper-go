package cfg

import (
	"encoding/json"
	"os"

	"github.com/r4start/goph-keeper/internal/client"
)

type AuthorizationData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type Config struct {
	Server        client.ServerEndpoint `json:"server"`
	StoragePath   string                `json:"storage_path"`
	SyncDirectory string                `json:"sync_dir"`
	filePath      string
}

func NewConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.filePath = filePath
	return cfg, err
}

func NewConfig() *Config {
	return &Config{
		Server: client.ServerEndpoint{
			Addr:   "localhost",
			Port:   "10081",
			UseTLS: false,
		},
	}
}

func (c *Config) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0700)
}
