package config

import (
    "encoding/json"
    "os"
    "path/filepath"

)

type AppConfig struct {
    LastPath string `json:"dir_path"`
}

func getConfigFilePath() string {
    if exePath, err := os.Executable(); err == nil {
        exeDir := filepath.Dir(exePath)
        projectConfigPath := filepath.Join(exeDir, "..", "..", "internal", "config", "config.json")
        if _, err := os.Stat(projectConfigPath); err == nil {
            return projectConfigPath
        }
        userConfigDir := filepath.Join(userHomeDir(), ".scaneval")
        _ = os.MkdirAll(userConfigDir, 0755)
        return filepath.Join(userConfigDir, "config.json")
    }

    dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
    return filepath.Join(dir, "config.json")
}

func userHomeDir() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return "."
    }
    return home
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










