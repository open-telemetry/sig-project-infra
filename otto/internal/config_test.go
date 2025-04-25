// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"testing"
)

func TestFileSecretConfig(t *testing.T) {
	// Create a test FileSecretConfig
	config := &FileSecretConfig{
		WebhookSecret:       "test-webhook-secret",
		GitHubAppID:         12345,
		GitHubInstallationID: 67890,
		GitHubPrivateKeyPath: "test-key-path",
		privateKey:          []byte("test-private-key"),
	}
	
	// Test GetWebhookSecret
	if config.GetWebhookSecret() != "test-webhook-secret" {
		t.Errorf("Expected webhook secret 'test-webhook-secret', got '%s'", config.GetWebhookSecret())
	}
	
	// Test GetGitHubAppID
	if config.GetGitHubAppID() != 12345 {
		t.Errorf("Expected GitHub App ID 12345, got %d", config.GetGitHubAppID())
	}
	
	// Test GetGitHubInstallationID
	if config.GetGitHubInstallationID() != 67890 {
		t.Errorf("Expected GitHub Installation ID 67890, got %d", config.GetGitHubInstallationID())
	}
	
	// Test GetGitHubPrivateKey
	if string(config.GetGitHubPrivateKey()) != "test-private-key" {
		t.Errorf("Expected private key 'test-private-key', got '%s'", string(config.GetGitHubPrivateKey()))
	}
}

func TestFileSecretConfigWithEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("OTTO_WEBHOOK_SECRET", "env-webhook-secret")
	os.Setenv("OTTO_GITHUB_APP_ID", "54321")
	os.Setenv("OTTO_GITHUB_INSTALLATION_ID", "98765")
	os.Setenv("OTTO_GITHUB_PRIVATE_KEY", "env-private-key")
	defer func() {
		os.Unsetenv("OTTO_WEBHOOK_SECRET")
		os.Unsetenv("OTTO_GITHUB_APP_ID")
		os.Unsetenv("OTTO_GITHUB_INSTALLATION_ID")
		os.Unsetenv("OTTO_GITHUB_PRIVATE_KEY")
	}()
	
	// Create an empty FileSecretConfig that will use env vars
	config := &FileSecretConfig{}
	
	// Test GetWebhookSecret with env var
	if config.GetWebhookSecret() != "env-webhook-secret" {
		t.Errorf("Expected webhook secret 'env-webhook-secret', got '%s'", config.GetWebhookSecret())
	}
	
	// Test GetGitHubAppID with env var
	if config.GetGitHubAppID() != 54321 {
		t.Errorf("Expected GitHub App ID 54321, got %d", config.GetGitHubAppID())
	}
	
	// Test GetGitHubInstallationID with env var
	if config.GetGitHubInstallationID() != 98765 {
		t.Errorf("Expected GitHub Installation ID 98765, got %d", config.GetGitHubInstallationID())
	}
	
	// Test GetGitHubPrivateKey with env var
	if string(config.GetGitHubPrivateKey()) != "env-private-key" {
		t.Errorf("Expected private key 'env-private-key', got '%s'", string(config.GetGitHubPrivateKey()))
	}
}

func TestValidateSecrets(t *testing.T) {
	// Test complete config
	complete := &FileSecretConfig{
		WebhookSecret:        "webhook-secret",
		GitHubAppID:          12345,
		GitHubInstallationID: 67890,
		GitHubPrivateKeyPath: "key-path",
	}
	if err := ValidateSecrets(complete); err != nil {
		t.Errorf("ValidateSecrets failed for complete config: %v", err)
	}
	
	// Test config with webhook secret only
	webhookOnly := &FileSecretConfig{
		WebhookSecret: "webhook-secret",
	}
	if err := ValidateSecrets(webhookOnly); err != nil {
		t.Errorf("ValidateSecrets failed for webhook-only config: %v", err)
	}
	
	// Test config with missing webhook secret
	missingWebhook := &FileSecretConfig{}
	if err := ValidateSecrets(missingWebhook); err == nil {
		t.Error("ValidateSecrets should fail for config with missing webhook secret")
	}
	
	// Test config with incomplete GitHub App config
	incompleteApp := &FileSecretConfig{
		WebhookSecret: "webhook-secret",
		GitHubAppID:  12345,
		// Missing installation ID and key path
	}
	if err := ValidateSecrets(incompleteApp); err == nil {
		t.Error("ValidateSecrets should fail for config with incomplete GitHub App config")
	}
}
