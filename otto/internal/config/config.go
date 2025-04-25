// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig contains non-secret application configuration
type AppConfig struct {
	Port    string         `yaml:"port"`
	DBPath  string         `yaml:"db_path"`
	Log     map[string]any `yaml:"log"`
	Modules map[string]any `yaml:"modules"`
}

// Global stores the main application configuration
var Global AppConfig

// Load reads YAML config from path into Global.
func Load(path string) error {
	config, err := LoadFromFile(path)
	if err != nil {
		return err
	}

	// Update global config
	Global = *config
	return nil
}

// LoadFromFile reads YAML config from path into an AppConfig struct
func LoadFromFile(path string) (*AppConfig, error) {
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

	// Apply defaults
	ApplyDefaults(config)

	// Log configuration summary
	LogSummary(config)

	return config, nil
}

// Validate checks that all required config fields are present and valid
func Validate(config *AppConfig) error {
	// No required fields in non-secret config
	return nil
}

// ApplyDefaults sets default values for optional config fields
func ApplyDefaults(config *AppConfig) {
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

// LogSummary logs a sanitized summary of the loaded configuration
func LogSummary(config *AppConfig) {
	slog.Info("configuration loaded",
		"port", config.Port,
		"db_path", config.DBPath,
		"log_level", config.Log["level"],
		"modules_configured", len(config.Modules))
}

// GetEnvOrDefault returns the value of the environment variable with the given key,
// or the default value if the environment variable is not set.
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}