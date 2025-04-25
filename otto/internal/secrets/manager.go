// SPDX-License-Identifier: Apache-2.0

// Package secrets provides interfaces and implementations for managing secrets.
package secrets

import (
	"fmt"
	"log/slog"
	"os"
)

// Manager is an interface for accessing secrets.
type Manager interface {
	// GetWebhookSecret returns the GitHub webhook secret.
	GetWebhookSecret() string

	// GetGitHubAppID returns the GitHub App ID.
	GetGitHubAppID() int64

	// GetGitHubInstallationID returns the GitHub App Installation ID.
	GetGitHubInstallationID() int64

	// GetGitHubPrivateKey returns the GitHub App private key.
	GetGitHubPrivateKey() []byte
}

// EnvManager implements the Manager interface using environment variables.
type EnvManager struct{}

// GetWebhookSecret returns the GitHub webhook secret from environment variable.
func (e *EnvManager) GetWebhookSecret() string {
	return os.Getenv("OTTO_WEBHOOK_SECRET")
}

// GetGitHubAppID returns the GitHub App ID from environment variable.
func (e *EnvManager) GetGitHubAppID() int64 {
	envVal := os.Getenv("OTTO_GITHUB_APP_ID")
	if envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}
	return 0
}

// GetGitHubInstallationID returns the GitHub App Installation ID from environment variable.
func (e *EnvManager) GetGitHubInstallationID() int64 {
	envVal := os.Getenv("OTTO_GITHUB_INSTALLATION_ID")
	if envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}
	return 0
}

// GetGitHubPrivateKey returns the GitHub App private key from environment variable.
func (e *EnvManager) GetGitHubPrivateKey() []byte {
	return []byte(os.Getenv("OTTO_GITHUB_PRIVATE_KEY"))
}

// FileManager implements the Manager interface using a local file.
type FileManager struct {
	WebhookSecret       string
	GitHubAppID         int64
	GitHubInstallationID int64
	GitHubPrivateKeyPath string
	privateKey          []byte
}

// NewFileManager creates a new FileManager with the given values.
func NewFileManager(webhook string, appID, installID int64, keyPath string, keyData []byte) *FileManager {
	return &FileManager{
		WebhookSecret:       webhook,
		GitHubAppID:         appID,
		GitHubInstallationID: installID,
		GitHubPrivateKeyPath: keyPath,
		privateKey:          keyData,
	}
}

// GetWebhookSecret returns the GitHub webhook secret, with environment variable fallback.
func (f *FileManager) GetWebhookSecret() string {
	if envVal := os.Getenv("OTTO_WEBHOOK_SECRET"); envVal != "" {
		return envVal
	}
	return f.WebhookSecret
}

// GetGitHubAppID returns the GitHub App ID, with environment variable fallback.
func (f *FileManager) GetGitHubAppID() int64 {
	if envVal := os.Getenv("OTTO_GITHUB_APP_ID"); envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}
	return f.GitHubAppID
}

// GetGitHubInstallationID returns the GitHub App Installation ID, with environment variable fallback.
func (f *FileManager) GetGitHubInstallationID() int64 {
	if envVal := os.Getenv("OTTO_GITHUB_INSTALLATION_ID"); envVal != "" {
		var id int64
		_, err := fmt.Sscanf(envVal, "%d", &id)
		if err == nil && id > 0 {
			return id
		}
	}
	return f.GitHubInstallationID
}

// GetGitHubPrivateKey returns the GitHub App private key, with environment variable fallback.
func (f *FileManager) GetGitHubPrivateKey() []byte {
	if envVal := os.Getenv("OTTO_GITHUB_PRIVATE_KEY"); envVal != "" {
		return []byte(envVal)
	}
	return f.privateKey
}

// ValidateFileManager checks that all required fields are present and valid.
func ValidateFileManager(secrets *FileManager) error {
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

// Chain implements the Manager interface by trying multiple managers in order.
type Chain struct {
	managers []Manager
}

// NewChain creates a new Chain with the given managers.
func NewChain(managers ...Manager) *Chain {
	return &Chain{managers: managers}
}

// GetWebhookSecret returns the GitHub webhook secret from the first manager that returns a non-empty value.
func (c *Chain) GetWebhookSecret() string {
	for _, m := range c.managers {
		if m == nil {
			continue
		}
		if v := m.GetWebhookSecret(); v != "" {
			return v
		}
	}
	return ""
}

// GetGitHubAppID returns the GitHub App ID from the first manager that returns a non-zero value.
func (c *Chain) GetGitHubAppID() int64 {
	for _, m := range c.managers {
		if m == nil {
			continue
		}
		if v := m.GetGitHubAppID(); v != 0 {
			return v
		}
	}
	return 0
}

// GetGitHubInstallationID returns the GitHub App Installation ID from the first manager that returns a non-zero value.
func (c *Chain) GetGitHubInstallationID() int64 {
	for _, m := range c.managers {
		if m == nil {
			continue
		}
		if v := m.GetGitHubInstallationID(); v != 0 {
			return v
		}
	}
	return 0
}

// GetGitHubPrivateKey returns the GitHub App private key from the first manager that returns a non-empty value.
func (c *Chain) GetGitHubPrivateKey() []byte {
	for _, m := range c.managers {
		if m == nil {
			continue
		}
		if v := m.GetGitHubPrivateKey(); len(v) > 0 {
			return v
		}
	}
	return nil
}

// LoadFileConfig loads secret configuration from a file.
func LoadFileConfig(path string) (*FileManager, error) {
	// Function implementation will be moved from config.go
	// This is a placeholder
	slog.Info("Loading secrets from file", "path", path)
	return nil, fmt.Errorf("not implemented")
}

// LoadFromEnv loads secret configuration from environment variables.
func LoadFromEnv() (*EnvManager, error) {
	// Check if required environment variables are set
	if os.Getenv("OTTO_WEBHOOK_SECRET") == "" {
		return nil, fmt.Errorf("OTTO_WEBHOOK_SECRET environment variable is required")
	}
	
	slog.Info("Loading secrets from environment variables")
	return &EnvManager{}, nil
}

// Validate checks if the manager has all required secrets.
func Validate(m Manager) error {
	if m.GetWebhookSecret() == "" {
		return fmt.Errorf("webhook secret is required")
	}
	
	// For GitHub App authentication, we need all three fields or none
	hasAppID := m.GetGitHubAppID() > 0
	hasInstallID := m.GetGitHubInstallationID() > 0
	hasKeyData := len(m.GetGitHubPrivateKey()) > 0
	
	if (hasAppID || hasInstallID || hasKeyData) && 
	   !(hasAppID && hasInstallID && hasKeyData) {
		return fmt.Errorf("github_app_id, github_installation_id, and github_private_key must all be set for GitHub App authentication")
	}
	
	return nil
}