// SPDX-License-Identifier: Apache-2.0

package module

import (
	"context"
	"encoding/json"

	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
	"github.com/open-telemetry/sig-project-infra/otto/internal/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
	"github.com/open-telemetry/sig-project-infra/otto/internal/telemetry"
)

// BaseModule provides common functionality for Otto modules.
// Modules can embed this struct to get access to shared functionality.
type BaseModule struct {
	name       string
	logger     logging.Logger
	telemetry  telemetry.Provider
	db         database.Provider
	repository database.Repository[database.AnyEntity]
	github     github.Provider
}

// Ensure BaseModule implements Interface.
var _ Interface = (*BaseModule)(nil)

// NewBaseModule creates a new base module with the given name and dependencies.
func NewBaseModule(name string, deps Dependencies) *BaseModule {
	logger := deps.Logger
	if logger == nil {
		// If no logger is provided, create a no-op logger
		logger = &noopLogger{}
	}

	// Create a module-specific logger with the module name
	moduleLogger := logger.With("module", name)

	return &BaseModule{
		name:       name,
		logger:     moduleLogger,
		telemetry:  deps.Telemetry,
		db:         deps.Database,
		repository: deps.Repository,
		github:     deps.GitHub,
	}
}

// Name returns the module's name.
func (m *BaseModule) Name() string {
	return m.name
}

// HandleEvent processes events from GitHub webhooks.
// This is a placeholder implementation that should be overridden by modules.
func (m *BaseModule) HandleEvent(eventType string, event any, raw json.RawMessage) error {
	return nil
}

// Logger returns the module's logger.
func (m *BaseModule) Logger() logging.Logger {
	return m.logger
}

// Telemetry returns the module's telemetry provider.
func (m *BaseModule) Telemetry() telemetry.Provider {
	return m.telemetry
}

// Database returns the module's database provider.
func (m *BaseModule) Database() database.Provider {
	return m.db
}

// Repository returns the module's repository.
func (m *BaseModule) Repository() database.Repository[database.AnyEntity] {
	return m.repository
}

// GitHub returns the module's GitHub client.
func (m *BaseModule) GitHub() github.Provider {
	return m.github
}

// MockModule is a mock implementation of Module for testing.
type MockModule struct {
	*BaseModule
	InitializeCalled bool
	ShutdownCalled   bool
	EventsHandled    []EventHandled
	ReturnError      error
}

// EventHandled represents a handled event for testing.
type EventHandled struct {
	Type  string
	Event any
	Raw   json.RawMessage
}

// NewMockModule creates a new mock module with the given name.
func NewMockModule(name string, deps Dependencies) *MockModule {
	return &MockModule{
		BaseModule:    NewBaseModule(name, deps),
		EventsHandled: make([]EventHandled, 0),
	}
}

// HandleEvent records event handling for testing.
func (m *MockModule) HandleEvent(eventType string, event any, raw json.RawMessage) error {
	m.EventsHandled = append(m.EventsHandled, EventHandled{
		Type:  eventType,
		Event: event,
		Raw:   raw,
	})
	return m.ReturnError
}

// Initialize implements Initializer for testing.
func (m *MockModule) Initialize(ctx context.Context) error {
	m.InitializeCalled = true
	return m.ReturnError
}

// Shutdown implements Shutdowner for testing.
func (m *MockModule) Shutdown(ctx context.Context) error {
	m.ShutdownCalled = true
	return m.ReturnError
}

// WasEventHandled checks if a specific event type was handled.
func (m *MockModule) WasEventHandled(eventType string) bool {
	for _, e := range m.EventsHandled {
		if e.Type == eventType {
			return true
		}
	}
	return false
}

// noopLogger is a logger that does nothing, used as a fallback.
type noopLogger struct{}

func (l *noopLogger) Debug(msg string, args ...any) {}
func (l *noopLogger) Info(msg string, args ...any)  {}
func (l *noopLogger) Warn(msg string, args ...any)  {}
func (l *noopLogger) Error(msg string, args ...any) {}
func (l *noopLogger) With(args ...any) logging.Logger {
	return l
}
