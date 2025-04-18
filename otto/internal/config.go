// SPDX-License-Identifier: Apache-2.0

// config.go handles loading YAML config for Otto and its modules.

package internal

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	WebHookSecret string         `yaml:"web_hook_secret"`
	Port          string         `yaml:"port"`
	DBPath        string         `yaml:"db_path"`
	GitHubToken   string         `yaml:"github_token"`
	Log           map[string]any `yaml:"log"`
	Modules       map[string]any `yaml:"modules"`
}

var GlobalConfig AppConfig

// LoadConfig reads YAML config from path into GlobalConfig.
func LoadConfig(path string) error {
	config, err := LoadConfigFromFile(path)
	if err != nil {
		return err
	}
	
	// Update global config
	GlobalConfig = *config
	return nil
}

// LoadConfigFromFile reads YAML config from path into an AppConfig struct
func LoadConfigFromFile(path string) (*AppConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	config := &AppConfig{}
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Validate required fields
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	// Apply defaults
	ApplyConfigDefaults(config)

	// Log configuration summary
	LogConfigSummary(config)

	return config, nil
}

// ValidateConfig checks that all required config fields are present and valid
func ValidateConfig(config *AppConfig) error {
	// Validate required fields
	if config.WebHookSecret == "" {
		return fmt.Errorf("webhook secret must be set")
	}
	
	// Additional validation can be added here
	
	return nil
}

// ApplyConfigDefaults sets default values for optional config fields
func ApplyConfigDefaults(config *AppConfig) {
	if config.Port == "" {
		config.Port = "8080"
	}

	if config.DBPath == "" {
		config.DBPath = "data.db"
	}

	if config.Log == nil {
		config.Log = map[string]any{
			"level":  "info",
			"format": "json",
		}
	}
}

// LogConfigSummary logs a sanitized summary of the loaded configuration
func LogConfigSummary(config *AppConfig) {
	slog.Info("configuration loaded", 
		"port", config.Port,
		"db_path", config.DBPath,
		"log_level", config.Log["level"],
		"modules_configured", len(config.Modules))
}
