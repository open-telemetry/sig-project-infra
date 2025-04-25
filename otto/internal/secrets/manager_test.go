// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"os"
	"testing"
)

func TestEnvManager(t *testing.T) {
	// Set environment variables
	os.Setenv("OTTO_WEBHOOK_SECRET", "test-webhook-secret")
	os.Setenv("OTTO_GITHUB_APP_ID", "54321")
	os.Setenv("OTTO_GITHUB_INSTALLATION_ID", "98765")
	os.Setenv("OTTO_GITHUB_PRIVATE_KEY", "test-private-key")
	defer func() {
		os.Unsetenv("OTTO_WEBHOOK_SECRET")
		os.Unsetenv("OTTO_GITHUB_APP_ID")
		os.Unsetenv("OTTO_GITHUB_INSTALLATION_ID")
		os.Unsetenv("OTTO_GITHUB_PRIVATE_KEY")
	}()

	// Create env manager
	envManager := &EnvManager{}

	// Test webhook secret
	if got := envManager.GetWebhookSecret(); got != "test-webhook-secret" {
		t.Errorf("GetWebhookSecret() = %v, want %v", got, "test-webhook-secret")
	}

	// Test GitHub App ID
	if got := envManager.GetGitHubAppID(); got != 54321 {
		t.Errorf("GetGitHubAppID() = %v, want %v", got, 54321)
	}

	// Test GitHub Installation ID
	if got := envManager.GetGitHubInstallationID(); got != 98765 {
		t.Errorf("GetGitHubInstallationID() = %v, want %v", got, 98765)
	}

	// Test GitHub Private Key
	if got := string(envManager.GetGitHubPrivateKey()); got != "test-private-key" {
		t.Errorf("GetGitHubPrivateKey() = %v, want %v", got, "test-private-key")
	}
}

func TestFileManager(t *testing.T) {
	// Create file manager
	fileManager := NewFileManager(
		"test-webhook-secret",
		12345,
		67890,
		"test-key-path",
		[]byte("test-private-key"),
	)

	// Test webhook secret
	if got := fileManager.GetWebhookSecret(); got != "test-webhook-secret" {
		t.Errorf("GetWebhookSecret() = %v, want %v", got, "test-webhook-secret")
	}

	// Test GitHub App ID
	if got := fileManager.GetGitHubAppID(); got != 12345 {
		t.Errorf("GetGitHubAppID() = %v, want %v", got, 12345)
	}

	// Test GitHub Installation ID
	if got := fileManager.GetGitHubInstallationID(); got != 67890 {
		t.Errorf("GetGitHubInstallationID() = %v, want %v", got, 67890)
	}

	// Test GitHub Private Key
	if got := string(fileManager.GetGitHubPrivateKey()); got != "test-private-key" {
		t.Errorf("GetGitHubPrivateKey() = %v, want %v", got, "test-private-key")
	}
}

func TestFileManagerWithEnv(t *testing.T) {
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

	// Create file manager
	fileManager := NewFileManager(
		"test-webhook-secret",
		12345,
		67890,
		"test-key-path",
		[]byte("test-private-key"),
	)

	// Test webhook secret (should get from env)
	if got := fileManager.GetWebhookSecret(); got != "env-webhook-secret" {
		t.Errorf("GetWebhookSecret() = %v, want %v", got, "env-webhook-secret")
	}

	// Test GitHub App ID (should get from env)
	if got := fileManager.GetGitHubAppID(); got != 54321 {
		t.Errorf("GetGitHubAppID() = %v, want %v", got, 54321)
	}

	// Test GitHub Installation ID (should get from env)
	if got := fileManager.GetGitHubInstallationID(); got != 98765 {
		t.Errorf("GetGitHubInstallationID() = %v, want %v", got, 98765)
	}

	// Test GitHub Private Key (should get from env)
	if got := string(fileManager.GetGitHubPrivateKey()); got != "env-private-key" {
		t.Errorf("GetGitHubPrivateKey() = %v, want %v", got, "env-private-key")
	}
}

func TestValidateFileManager(t *testing.T) {
	// Test complete config
	complete := NewFileManager(
		"webhook-secret",
		12345,
		67890,
		"key-path",
		[]byte("test-private-key"),
	)
	if err := ValidateFileManager(complete); err != nil {
		t.Errorf("ValidateFileManager failed for complete config: %v", err)
	}

	// Test config with webhook secret only
	webhookOnly := NewFileManager(
		"webhook-secret",
		0,
		0,
		"",
		nil,
	)
	if err := ValidateFileManager(webhookOnly); err != nil {
		t.Errorf("ValidateFileManager failed for webhook-only config: %v", err)
	}

	// Test config with missing webhook secret
	missingWebhook := NewFileManager(
		"",
		0,
		0,
		"",
		nil,
	)
	if err := ValidateFileManager(missingWebhook); err == nil {
		t.Error("ValidateFileManager should fail for config with missing webhook secret")
	}

	// Test config with incomplete GitHub App config
	incompleteApp := NewFileManager(
		"webhook-secret",
		12345,
		0, // Missing installation ID and key path
		"",
		nil,
	)
	if err := ValidateFileManager(incompleteApp); err == nil {
		t.Error("ValidateFileManager should fail for config with incomplete GitHub App config")
	}
}

func TestChain(t *testing.T) {
	// Create managers
	fileManager := NewFileManager(
		"file-webhook-secret",
		12345,
		67890,
		"file-key-path",
		[]byte("file-private-key"),
	)
	
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
	
	envManager := &EnvManager{}
	
	// Create chain with env first, file second
	chain1 := NewChain(envManager, fileManager)
	
	// Test chain with env first
	if got := chain1.GetWebhookSecret(); got != "env-webhook-secret" {
		t.Errorf("GetWebhookSecret() for chain1 = %v, want %v", got, "env-webhook-secret")
	}
	
	if got := chain1.GetGitHubAppID(); got != 54321 {
		t.Errorf("GetGitHubAppID() for chain1 = %v, want %v", got, 54321)
	}
	
	// Testing order is important - env vars take precedence over file in our current test
	chain2 := NewChain(fileManager, envManager)
	
	// Test chain2, but expect env vars to still be used since they exist
	if got := chain2.GetWebhookSecret(); got != "env-webhook-secret" {
		t.Errorf("GetWebhookSecret() for chain2 = %v, want %v", got, "env-webhook-secret")
	}
	
	if got := chain2.GetGitHubAppID(); got != 54321 {
		t.Errorf("GetGitHubAppID() for chain2 = %v, want %v", got, 54321)
	}
	
	// Create chain with nil managers (should skip them)
	chain3 := NewChain(nil, fileManager)
	
	// But still expect env vars to be used via fileManager's implementation
	if got := chain3.GetWebhookSecret(); got != "env-webhook-secret" {
		t.Errorf("GetWebhookSecret() for chain3 = %v, want %v", got, "env-webhook-secret")
	}
}