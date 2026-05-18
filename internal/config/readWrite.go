package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Db_url string `json:"db_url"`
	Name   string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filepath := filepath.Join(homeDir, ".gatorconfig.json")

	return filepath, err
}

func Read() (Config, error) {

	filename, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	result := Config{}

	if err := json.Unmarshal(data, &result); err != nil {
		return Config{}, err
	}

	return result, nil
}

func SetUser(user string) error {

	configStruct, err := Read()
	if err != nil {
		return err
	}

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configStruct.Name = user

	rawBytes, err := json.Marshal(configStruct)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, rawBytes, 0777); err != nil {
		return err
	}
	return nil
}
