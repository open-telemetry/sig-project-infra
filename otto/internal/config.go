// SPDX-License-Identifier: Apache-2.0

// config.go handles loading YAML config for Otto and its modules.

package internal

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

// GlobalConfig stores the main application configuration
var GlobalConfig AppConfig

// SecretConfig is an interface for accessing secrets
type SecretConfig interface {
	// GetWebhookSecret returns the GitHub webhook secret
	GetWebhookSecret() string
	
	// GetGitHubAppID returns the GitHub App ID
	GetGitHubAppID() int64
	
	// GetGitHubInstallationID returns the GitHub App Installation ID
	GetGitHubInstallationID() int64
	
	// GetGitHubPrivateKey returns the GitHub App private key
	GetGitHubPrivateKey() []byte
}

// FileSecretConfig implements SecretConfig by loading from a YAML file
type FileSecretConfig struct {
	WebhookSecret       string `yaml:"webhook_secret"`
	GitHubAppID         int64  `yaml:"github_app_id"`
	GitHubInstallationID int64  `yaml:"github_installation_id"`
	GitHubPrivateKeyPath string `yaml:"github_private_key_path"`
	privateKey          []byte // loaded from file, not from YAML
}

// GetWebhookSecret returns the webhook secret
func (f *FileSecretConfig) GetWebhookSecret() string {
	if envVal := os.Getenv("OTTO_WEBHOOK_SECRET"); envVal != "" {
		return envVal
	}
	return f.WebhookSecret
}

// GetGitHubAppID returns the GitHub App ID
func (f *FileSecretConfig) GetGitHubAppID() int64 {
	if envVal := os.Getenv("OTTO_GITHUB_APP_ID"); envVal != "" {
		var id int64
		fmt.Sscanf(envVal, "%d", &id)
		if id > 0 {
			return id
		}
	}
	return f.GitHubAppID
}

// GetGitHubInstallationID returns the GitHub App Installation ID
func (f *FileSecretConfig) GetGitHubInstallationID() int64 {
	if envVal := os.Getenv("OTTO_GITHUB_INSTALLATION_ID"); envVal != "" {
		var id int64
		fmt.Sscanf(envVal, "%d", &id)
		if id > 0 {
			return id
		}
	}
	return f.GitHubInstallationID
}

// GetGitHubPrivateKey returns the GitHub App private key
func (f *FileSecretConfig) GetGitHubPrivateKey() []byte {
	// First check if private key is in environment variable
	if envVal := os.Getenv("OTTO_GITHUB_PRIVATE_KEY"); envVal != "" {
		return []byte(envVal)
	}
	
	// Otherwise, return the loaded private key
	return f.privateKey
}

// GlobalSecrets stores the application secrets
var GlobalSecrets SecretConfig

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

// LoadSecrets loads secrets from the given path or environment variables
func LoadSecrets(path string) (SecretConfig, error) {
	// Try to load from file first
	secrets, err := LoadSecretsFromFile(path)
	if err != nil {
		// If file doesn't exist, try environment variables
		if os.IsNotExist(err) {
			slog.Info("secrets file not found, checking environment variables")
			
			// Check if required environment variables are set
			if os.Getenv("OTTO_WEBHOOK_SECRET") == "" {
				return nil, fmt.Errorf("OTTO_WEBHOOK_SECRET environment variable is required when secrets file is not present")
			}
			
			// Create a minimal FileSecretConfig that will use environment variables
			return &FileSecretConfig{}, nil
		}
		return nil, err
	}
	
	// Load private key from file if path is specified
	if secrets.GitHubPrivateKeyPath != "" {
		keyData, err := os.ReadFile(secrets.GitHubPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GitHub private key: %w", err)
		}
		secrets.privateKey = keyData
	}
	
	return secrets, nil
}

// LoadSecretsFromFile loads secrets from a YAML file
func LoadSecretsFromFile(path string) (*FileSecretConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	secrets := &FileSecretConfig{}
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(secrets); err != nil {
		return nil, fmt.Errorf("failed to decode secrets: %w", err)
	}

	// Validate required fields
	if err := ValidateSecrets(secrets); err != nil {
		return nil, err
	}

	slog.Info("secrets loaded successfully")
	return secrets, nil
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

	// Apply defaults
	ApplyConfigDefaults(config)

	// Log configuration summary
	LogConfigSummary(config)

	return config, nil
}

// ValidateConfig checks that all required config fields are present and valid
func ValidateConfig(config *AppConfig) error {
	// No required fields in non-secret config
	return nil
}

// ValidateSecrets checks that all required secret fields are present and valid
func ValidateSecrets(secrets *FileSecretConfig) error {
	// Skip validation if we expect environment variables
	if os.Getenv("OTTO_WEBHOOK_SECRET") != "" {
		return nil
	}
	
	// Validate required fields
	if secrets.WebhookSecret == "" {
		return fmt.Errorf("webhook_secret must be set")
	}
	
	// For GitHub App authentication, we need all three fields or none
	hasAppID := secrets.GitHubAppID > 0
	hasInstallID := secrets.GitHubInstallationID > 0
	hasKeyPath := secrets.GitHubPrivateKeyPath != ""
	
	if (hasAppID || hasInstallID || hasKeyPath) && 
	   !(hasAppID && hasInstallID && hasKeyPath) {
		return fmt.Errorf("github_app_id, github_installation_id, and github_private_key_path must all be set for GitHub App authentication")
	}
	
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
