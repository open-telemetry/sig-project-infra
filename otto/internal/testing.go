// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	// Import sqlite driver for database/sql.
	_ "modernc.org/sqlite"
)

// TestApp creates a test application with mock dependencies.
func TestApp(t *testing.T) *App {
	// Create temp config file
	configFile, err := os.CreateTemp(t.TempDir(), "otto-test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	configPath := configFile.Name()
	defer os.Remove(configPath)

	// Write test config
	config := `port: "8081"
db_path: ":memory:"
log:
  level: "info"
  format: "text"
modules:
  name: "test"
  config:
    key: "value"
`
	if _, err := configFile.Write([]byte(config)); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatalf("Failed to close test config file: %v", err)
	}

	// Create temp secrets file
	secretsFile, err := os.CreateTemp(t.TempDir(), "otto-test-secrets-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create test secrets file: %v", err)
	}
	secretsPath := secretsFile.Name()
	defer os.Remove(secretsPath)

	// Write test secrets
	secrets := `webhook_secret: "test-secret"
github_app_id: 12345
github_installation_id: 67890
github_private_key_path: ""
`
	if _, err := secretsFile.Write([]byte(secrets)); err != nil {
		t.Fatalf("Failed to write test secrets: %v", err)
	}
	if err := secretsFile.Close(); err != nil {
		t.Fatalf("Failed to close test secrets file: %v", err)
	}

	// Initialize test app
	app, err := NewApp(t.Context(), configPath, secretsPath)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	return app
}

// TestDB creates an in-memory SQLite database for testing.
func TestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Ensure the connection works
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// TestRepository creates a repository with an in-memory database for testing.
func TestRepository(t *testing.T) Repository {
	db := TestDB(t)
	return NewSQLiteRepository(db)
}

// Note: MockCommandHandler has been removed since commands are now
// processed directly by modules in their HandleEvent implementation.

// MockEventHandler is a function that can be used to mock an event handler.
type MockEventHandler struct {
	HandleEventFunc func(eventType string, event any, raw []byte) error
}

// HandleEvent implements the Module interface.
func (m *MockEventHandler) HandleEvent(eventType string, event any, raw []byte) error {
	if m.HandleEventFunc == nil {
		return nil
	}
	return m.HandleEventFunc(eventType, event, raw)
}

// MockModule is a mock implementation of the Module interface for testing.
type MockModule struct {
	MockEventHandler
	name string
}

// Name implements the Module interface.
func (m *MockModule) Name() string {
	return m.name
}

// NewMockModule creates a new mock module with the given name.
func NewMockModule(name string) *MockModule {
	return &MockModule{name: name}
}

// GitHubWebhookPayload represents a structured GitHub webhook event for testing.
type GitHubWebhookPayload struct {
	Action       string                 `json:"action"`
	Issue        map[string]interface{} `json:"issue,omitempty"`
	PullRequest  map[string]interface{} `json:"pull_request,omitempty"`
	Comment      map[string]interface{} `json:"comment,omitempty"`
	Review       map[string]interface{} `json:"review,omitempty"`
	Sender       map[string]interface{} `json:"sender"`
	Repository   map[string]interface{} `json:"repository"`
	Organization map[string]interface{} `json:"organization,omitempty"`
}

// CreateTestWebhookPayload generates a simulated GitHub webhook payload.
func CreateTestWebhookPayload(eventType string, options map[string]interface{}) ([]byte, error) {
	payload := GitHubWebhookPayload{
		Action: options["action"].(string),
		Sender: map[string]interface{}{
			"login": options["sender"].(string),
		},
		Repository: map[string]interface{}{
			"full_name": options["repo"].(string),
			"name":      strings.Split(options["repo"].(string), "/")[1],
			"owner": map[string]interface{}{
				"login": strings.Split(options["repo"].(string), "/")[0],
			},
		},
	}

	switch eventType {
	case "issues":
		payload.Issue = map[string]interface{}{
			"number": options["issue_number"].(int),
			"title":  options["title"].(string),
			"body":   options["body"].(string),
		}
	case "issue_comment":
		payload.Issue = map[string]interface{}{
			"number": options["issue_number"].(int),
		}
		payload.Comment = map[string]interface{}{
			"body": options["comment"].(string),
		}
	case "pull_request":
		payload.PullRequest = map[string]interface{}{
			"number": options["pr_number"].(int),
			"title":  options["title"].(string),
			"body":   options["body"].(string),
		}
	case "pull_request_review":
		payload.PullRequest = map[string]interface{}{
			"number": options["pr_number"].(int),
		}
		payload.Review = map[string]interface{}{
			"state": options["review_state"].(string),
		}
	}

	return json.Marshal(payload)
}

// SimulateWebhookEvent simulates sending a GitHub webhook event to the application.
func (a *App) SimulateWebhookEvent(eventType string, options map[string]interface{}) error {
	payload, err := CreateTestWebhookPayload(eventType, options)
	if err != nil {
		return err
	}

	// Simulate event dispatch
	a.DispatchEvent(eventType, payload, payload)
	return nil
}

// Note: Command simulation has been removed since commands are now
// processed directly by modules in their HandleEvent implementation.
