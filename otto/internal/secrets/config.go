// SPDX-License-Identifier: Apache-2.0

// Package secrets provides secure storage and retrieval of sensitive configuration.
package secrets

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the secrets configuration in a YAML file.
type FileConfig struct {
	WebhookSecret        string `yaml:"webhook_secret"`
	GitHubAppID          int64  `yaml:"github_app_id"`
	GitHubInstallationID int64  `yaml:"github_installation_id"`
	GitHubPrivateKeyPath string `yaml:"github_private_key_path"`
}

// OnePasswordConfig represents the 1Password secrets configuration in a YAML file.
type OnePasswordConfig struct {
	WebhookSecretRef string `yaml:"webhook_secret_ref"`
	AppIDRef         string `yaml:"github_app_id_ref"`
	InstallIDRef     string `yaml:"github_installation_id_ref"`
	PrivateKeyRef    string `yaml:"github_private_key_ref"`
}

// Manager implementations provide access to sensitive configuration.

// LoadSecrets loads secrets from the given path, environment variables, or 1Password.
func LoadSecrets(path string) (Manager, error) {
	// Check for 1Password configuration
	opPath := os.Getenv("OTTO_1PASSWORD_CONFIG")
	if opPath != "" {
		// Try to load 1Password configuration
		secrets, err := loadOnePasswordConfig(opPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load 1Password secrets: %w", err)
		}
		return secrets, nil
	}

	// Try to load from file
	secrets, err := loadFileConfig(path)
	if err != nil {
		// If file doesn't exist, try environment variables
		if os.IsNotExist(err) {
			slog.Info("secrets file not found, checking environment variables")

			// Check if required environment variables are set
			if os.Getenv("OTTO_WEBHOOK_SECRET") == "" {
				return nil, errors.New(
					"OTTO_WEBHOOK_SECRET environment variable is required when secrets file is not present",
				)
			}

			// Create an environment manager
			return &EnvManager{}, nil
		}
		return nil, err
	}

	return secrets, nil
}

// loadFileConfig loads secrets from a YAML file.
func loadFileConfig(path string) (*FileManager, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config FileConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode secrets: %w", err)
	}

	// Create a file manager
	manager := NewFileManager(
		config.WebhookSecret,
		config.GitHubAppID,
		config.GitHubInstallationID,
		config.GitHubPrivateKeyPath,
		nil, // Private key will be loaded below
	)

	// Load private key from file if path is specified
	if config.GitHubPrivateKeyPath != "" {
		keyData, err := os.ReadFile(config.GitHubPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GitHub private key: %w", err)
		}
		manager.privateKey = keyData
	}

	// Validate the configuration
	if err := ValidateFileManager(manager); err != nil {
		return nil, err
	}

	slog.Info("secrets loaded successfully")
	return manager, nil
}

// loadOnePasswordConfig loads secrets from a 1Password configuration file.
func loadOnePasswordConfig(path string) (*OnePasswordManager, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config OnePasswordConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode 1Password config: %w", err)
	}

	// Create a 1Password manager
	manager, err := NewOnePasswordManager(
		config.WebhookSecretRef,
		config.AppIDRef,
		config.InstallIDRef,
		config.PrivateKeyRef,
	)
	if err != nil {
		return nil, err
	}

	slog.Info("1Password secrets configured successfully")
	return manager, nil
}
