package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
    DBUrl  string `json:"db_url"`
    CurrentUserName   string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("Error getting home directory: %w", err)
    }
    filePath := filepath.Join(homeDir, ".gatorconfig.json")
    return filePath, nil
}

func write(cfg Config) error {
    fullPath, err := getConfigFilePath()
    if err != nil {
        return  fmt.Errorf("Error getting home directory: %w", err)
    }
    jsonData, err := json.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("unable to turn struct into Json")
    }
    err = os.WriteFile(fullPath, jsonData, 0644)
    if err != nil {
        return fmt.Errorf("failed to write file")
    }
    return nil
}

func Read() (Config, error) {
    fullPath, err := getConfigFilePath()
    if err != nil {
        return Config{}, fmt.Errorf("Error getting home directory: %w", err)
    }
    file, err := os.ReadFile(fullPath)
    if err != nil {
        return Config{}, fmt.Errorf("Unable to open file: %w", err)
    }
    var jsonData Config
    err = json.Unmarshal(file, &jsonData)
    if err != nil {
        return Config{}, fmt.Errorf("Unable to decode data: %w", err)
    }
    return jsonData, nil
}

func (cfg *Config) SetUser(username string) error {
    cfg.CurrentUserName = username
    err := write(*cfg)
    if err != nil {
        return fmt.Errorf("Unable to write to file")
    }
    return nil
}

