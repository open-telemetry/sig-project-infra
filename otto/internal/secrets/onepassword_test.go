// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"os"
	"testing"
)

func TestOnePasswordManagerValidation(t *testing.T) {
	// Skip test if not running in an environment with actual 1Password access
	if os.Getenv("OTTO_RUN_1PASSWORD_TESTS") == "" {
		t.Skip("Skipping 1Password tests - set OTTO_RUN_1PASSWORD_TESTS=1 to run")
	}

	// Skip test if token is not available
	token := os.Getenv("OTTO_1PASSWORD_TOKEN")
	if token == "" {
		t.Skip("Skipping 1Password test as OTTO_1PASSWORD_TOKEN is not set")
	}

	// Only run these tests when explicitly requested with proper token set up

	// Test missing webhook reference
	_, err := NewOnePasswordManager("", "", "", "")
	if err == nil {
		t.Error("NewOnePasswordManager should fail for empty webhook reference")
	}

	// Test incomplete GitHub App references
	_, err = NewOnePasswordManager(
		"op://vault/item/webhook",
		"op://vault/item/app_id",
		"", // Missing installation ID
		"",
	)
	if err == nil {
		t.Error("NewOnePasswordManager should fail for incomplete GitHub App references")
	}
}

func TestOnePasswordManagerEnvironmentFallback(t *testing.T) {
	// Skip test if not running in an environment with actual 1Password access
	if os.Getenv("OTTO_RUN_1PASSWORD_TESTS") == "" {
		t.Skip("Skipping 1Password tests - set OTTO_RUN_1PASSWORD_TESTS=1 to run")
	}

	// Set environment variables
	t.Setenv("OTTO_WEBHOOK_SECRET", "env-webhook-secret")
	t.Setenv("OTTO_GITHUB_APP_ID", "54321")
	t.Setenv("OTTO_GITHUB_INSTALLATION_ID", "98765")
	t.Setenv("OTTO_GITHUB_PRIVATE_KEY", "env-private-key")

	// Skip test if token is not available
	token := os.Getenv("OTTO_1PASSWORD_TOKEN")
	if token == "" {
		t.Skip("Skipping 1Password test as OTTO_1PASSWORD_TOKEN is not set")
	}
}
