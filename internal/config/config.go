package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type AppConfig struct {
	LastPath string `json:"last_path"`
}

func getConfigFilePath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "config.json")
}

func SaveLastPath(path string) error {
	cfg := AppConfig{LastPath: path}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigFilePath(), data, 0644)
}

func LoadLastPath() (string, error) {
	data, err := os.ReadFile(getConfigFilePath())
	if err != nil {
		return "", err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.LastPath, nil
}
