package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// Construct the full path to the config file
	configPath := filepath.Join(homeDir, ".gatorconfig.json")

	// Convert to JSON
	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}
	// Write to file
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetDbURL(url string) error {
	c.DbURL = url
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// Construct the full path to the config file
	configPath := filepath.Join(homeDir, ".gatorconfig.json")
	// Convert to JSON
	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}
	// Write to file
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil

}

func Read() (Config, error) {
	var cfg Config
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, err
	}
	// Construct the full path to the config file
	configPath := filepath.Join(homeDir, ".gatorconfig.json")
	// Open the file
	file, err := os.Open(configPath)
	if err != nil {
		return cfg, err
	}
	defer file.Close()
	// Decode the JSON into the Config struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
