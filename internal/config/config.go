package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
    DBUrl  string `json:"db_url"`
    CurrentUserName   string `json:"current_user_name"`
}


func Read() (Config, error) {

    homeDir, err := os.UserHomeDir()
    if err != nil {
        return Config{}, fmt.Errorf("Error getting home directory: %w", err)
    }

    filePath := filepath.Join(homeDir, ".gatorconfig.json")
    f, err := os.Open(filePath)
    if err != nil {
        return Config{}, fmt.Errorf("Unable to open file: %w", err)
    }

    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return Config{}, fmt.Errorf("Unable to read file into buffer: %w", err)
    }

    config := Config{}
    err = json.Unmarshal(data, &config)
    if err != nil {
        return Config{}, fmt.Errorf("Unable to Unmarshal data into new Config struct: %w", err)
    }

    return config, nil
}

//func (cfg *Config) error {
//
//}

func (cfg *Config) SetUer(username string) error {

    cfg.CurrentUserName = username
    fmt.Println()

    return nil
}
