package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Addons []string `json:"addons"`
	path   string
}

func configDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "cremio")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cremio")
}

func Load() (*Config, error) {
	dir := configDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.Addons = []string{}
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0o644)
}

func (c *Config) AddAddon(url string) {
	for _, a := range c.Addons {
		if a == url {
			return
		}
	}
	c.Addons = append(c.Addons, url)
}

func (c *Config) RemoveAddon(url string) {
	for i, a := range c.Addons {
		if a == url {
			c.Addons = append(c.Addons[:i], c.Addons[i+1:]...)
			return
		}
	}
}
