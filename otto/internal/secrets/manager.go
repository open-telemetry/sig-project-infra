// SPDX-License-Identifier: Apache-2.0

// Package secrets provides interfaces and implementations for managing secrets.
package secrets

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
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
type EnvManager struct {
	webhookSecret  string
	gitHubAppID    int64
	installationID int64
	privateKey     []byte
}

// NewEnvManager creates a new EnvManager that reads from environment variables once.
func NewEnvManager() *EnvManager {
	e := &EnvManager{
		webhookSecret: os.Getenv("OTTO_WEBHOOK_SECRET"),
	}

	if appIDStr := os.Getenv("OTTO_GITHUB_APP_ID"); appIDStr != "" {
		appID, err := strconv.ParseInt(appIDStr, 10, 64)
		if err == nil && appID > 0 {
			e.gitHubAppID = appID
		}
	}

	if installIDStr := os.Getenv("OTTO_GITHUB_INSTALLATION_ID"); installIDStr != "" {
		installID, err := strconv.ParseInt(installIDStr, 10, 64)
		if err == nil && installID > 0 {
			e.installationID = installID
		}
	}

	if privateKey := os.Getenv("OTTO_GITHUB_PRIVATE_KEY"); privateKey != "" {
		e.privateKey = []byte(privateKey)
	}

	return e
}

// GetWebhookSecret returns the GitHub webhook secret from environment variable.
func (e *EnvManager) GetWebhookSecret() string {
	return e.webhookSecret
}

// GetGitHubAppID returns the GitHub App ID from environment variable.
func (e *EnvManager) GetGitHubAppID() int64 {
	return e.gitHubAppID
}

// GetGitHubInstallationID returns the GitHub App Installation ID from environment variable.
func (e *EnvManager) GetGitHubInstallationID() int64 {
	return e.installationID
}

// GetGitHubPrivateKey returns the GitHub App private key from environment variable.
func (e *EnvManager) GetGitHubPrivateKey() []byte {
	return e.privateKey
}

// FileManager implements the Manager interface using a local file.
type FileManager struct {
	WebhookSecret        string
	GitHubAppID          int64
	GitHubInstallationID int64
	GitHubPrivateKeyPath string
	privateKey           []byte

	// Environment values take precedence and are cached during initialization
	envWebhookSecret  string
	envGitHubAppID    int64
	envInstallationID int64
	envPrivateKey     []byte
	hasEnvWebhook     bool
	hasEnvAppID       bool
	hasEnvInstallID   bool
	hasEnvPrivateKey  bool
}

// NewFileManager creates a new FileManager with the given values.
func NewFileManager(
	webhook string,
	appID, installID int64,
	keyPath string,
	keyData []byte,
) *FileManager {
	fm := &FileManager{
		WebhookSecret:        webhook,
		GitHubAppID:          appID,
		GitHubInstallationID: installID,
		GitHubPrivateKeyPath: keyPath,
		privateKey:           keyData,
	}

	// Check for environment variables once during initialization
	if envVal := os.Getenv("OTTO_WEBHOOK_SECRET"); envVal != "" {
		fm.envWebhookSecret = envVal
		fm.hasEnvWebhook = true
	}

	if envVal := os.Getenv("OTTO_GITHUB_APP_ID"); envVal != "" {
		id, err := strconv.ParseInt(envVal, 10, 64)
		if err == nil && id > 0 {
			fm.envGitHubAppID = id
			fm.hasEnvAppID = true
		}
	}

	if envVal := os.Getenv("OTTO_GITHUB_INSTALLATION_ID"); envVal != "" {
		id, err := strconv.ParseInt(envVal, 10, 64)
		if err == nil && id > 0 {
			fm.envInstallationID = id
			fm.hasEnvInstallID = true
		}
	}

	if envVal := os.Getenv("OTTO_GITHUB_PRIVATE_KEY"); envVal != "" {
		fm.envPrivateKey = []byte(envVal)
		fm.hasEnvPrivateKey = true
	}

	return fm
}

// GetWebhookSecret returns the GitHub webhook secret, with environment variable fallback.
func (f *FileManager) GetWebhookSecret() string {
	if f.hasEnvWebhook {
		return f.envWebhookSecret
	}
	return f.WebhookSecret
}

// GetGitHubAppID returns the GitHub App ID, with environment variable fallback.
func (f *FileManager) GetGitHubAppID() int64 {
	if f.hasEnvAppID {
		return f.envGitHubAppID
	}
	return f.GitHubAppID
}

// GetGitHubInstallationID returns the GitHub App Installation ID, with environment variable fallback.
func (f *FileManager) GetGitHubInstallationID() int64 {
	if f.hasEnvInstallID {
		return f.envInstallationID
	}
	return f.GitHubInstallationID
}

// GetGitHubPrivateKey returns the GitHub App private key, with environment variable fallback.
func (f *FileManager) GetGitHubPrivateKey() []byte {
	if f.hasEnvPrivateKey {
		return f.envPrivateKey
	}
	return f.privateKey
}

// ValidateFileManager checks that all required fields are present and valid.
func ValidateFileManager(secrets *FileManager) error {
	// Skip validation if we have webhook secret from environment
	if secrets.hasEnvWebhook {
		return nil
	}

	// Validate required fields
	if secrets.WebhookSecret == "" {
		return errors.New("webhook_secret must be set")
	}

	// For GitHub App authentication, we need all three fields or none
	hasAppID := secrets.GitHubAppID > 0 || secrets.hasEnvAppID
	hasInstallID := secrets.GitHubInstallationID > 0 || secrets.hasEnvInstallID
	hasKeyPath := secrets.GitHubPrivateKeyPath != "" || secrets.hasEnvPrivateKey

	if (hasAppID || hasInstallID || hasKeyPath) &&
		(!hasAppID || !hasInstallID || !hasKeyPath) {
		return errors.New(
			"github_app_id, github_installation_id, and github_private_key_path must all be set for GitHub App authentication",
		)
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
	return nil, errors.New("not implemented")
}

// LoadFromEnv loads secret configuration from environment variables.
func LoadFromEnv() (*EnvManager, error) {
	// Create a new EnvManager
	envManager := NewEnvManager()

	// Check if required environment variables are present
	if envManager.GetWebhookSecret() == "" {
		return nil, errors.New("OTTO_WEBHOOK_SECRET environment variable is required")
	}

	slog.Info("Loading secrets from environment variables")
	return envManager, nil
}

// Validate checks if the manager has all required secrets.
func Validate(m Manager) error {
	if m.GetWebhookSecret() == "" {
		return errors.New("webhook secret is required")
	}

	// For GitHub App authentication, we need all three fields or none
	hasAppID := m.GetGitHubAppID() > 0
	hasInstallID := m.GetGitHubInstallationID() > 0
	hasKeyData := len(m.GetGitHubPrivateKey()) > 0

	if (hasAppID || hasInstallID || hasKeyData) &&
		(!hasAppID || !hasInstallID || !hasKeyData) {
		return errors.New(
			"github_app_id, github_installation_id, and github_private_key must all be set for GitHub App authentication",
		)
	}

	return nil
}
