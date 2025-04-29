// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file for testing
	tempFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test config to the temp file
	testConfig := `
port: "9090"
db_path: "test.db"
log:
  level: "debug"
  format: "json"
modules:
  test: true
`
	if _, err := tempFile.Write([]byte(testConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load the config
	config, err := LoadFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Check that the config was loaded correctly
	if config.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Port)
	}
	if config.DBPath != "test.db" {
		t.Errorf("Expected db_path test.db, got %s", config.DBPath)
	}
	if config.Log["level"] != "debug" {
		t.Errorf("Expected log level debug, got %s", config.Log["level"])
	}
	if config.Log["format"] != "json" {
		t.Errorf("Expected log format json, got %s", config.Log["format"])
	}
	if _, ok := config.Modules["test"]; !ok {
		t.Errorf("Expected modules to contain test")
	}
}

func TestApplyDefaults(t *testing.T) {
	// Create a config with no values
	config := &AppConfig{}

	// Apply defaults
	ApplyDefaults(config)

	// Check that defaults were applied
	if config.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Port)
	}
	if config.DBPath != "data.db" {
		t.Errorf("Expected default db_path data.db, got %s", config.DBPath)
	}
	if config.Log["level"] != "info" {
		t.Errorf("Expected default log level info, got %s", config.Log["level"])
	}
	if config.Log["format"] != "json" {
		t.Errorf("Expected default log format json, got %s", config.Log["format"])
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Set a test environment variable
	t.Setenv("TEST_ENV_VAR", "test-value")

	// Test with existing environment variable
	if got := GetEnvOrDefault("TEST_ENV_VAR", "default"); got != "test-value" {
		t.Errorf("GetEnvOrDefault() = %v, want %v", got, "test-value")
	}

	// Test with non-existent environment variable
	if got := GetEnvOrDefault("NON_EXISTENT_VAR", "default"); got != "default" {
		t.Errorf("GetEnvOrDefault() = %v, want %v", got, "default")
	}
}
