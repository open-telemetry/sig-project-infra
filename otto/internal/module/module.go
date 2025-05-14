// SPDX-License-Identifier: Apache-2.0

// Package module defines interfaces and registry for Otto feature modules.
package module

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
	"github.com/open-telemetry/sig-project-infra/otto/internal/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
	"github.com/open-telemetry/sig-project-infra/otto/internal/telemetry"
)

// CommandContext represents a slash command invocation.
type CommandContext struct {
	Context  context.Context
	Command  string   // e.g. "oncall"
	Args     []string // parsed args
	Issuer   string   // user who issued command
	Repo     string
	IssueNum int
	RawBody  string // raw comment body, if needed
}

// Interface is the basic interface that all modules must implement.
type Interface interface {
	// Name returns the module's name.
	Name() string
	// HandleEvent processes events from GitHub webhooks.
	HandleEvent(eventType string, event any, raw json.RawMessage) error
}

// Initializer is an optional interface that modules can implement
// for initialization logic.
type Initializer interface {
	// Initialize initializes the module with the provided dependencies.
	Initialize(ctx context.Context) error
}

// Shutdowner is an optional interface that modules can implement
// for graceful shutdown.
type Shutdowner interface {
	// Shutdown performs cleanup when the application is shutting down.
	Shutdown(ctx context.Context) error
}

// Module represents a complete Otto module with all dependencies.
type Module interface {
	Interface

	// Optional interfaces
	// A module may implement these, but they're not required
	// Initializer
	// Shutdowner
}

// Dependencies contains all the dependencies a module might need.
type Dependencies struct {
	Logger     logging.Logger
	Telemetry  telemetry.Provider
	Database   database.Provider
	Repository database.Repository[database.AnyEntity]
	GitHub     github.Provider
}

// Registry manages the registration and retrieval of modules.
type Registry interface {
	// RegisterModule adds a module to the registry.
	RegisterModule(m Module)
	// GetModules returns a copy of the registered modules map.
	GetModules() map[string]Module
}

// DefaultRegistry is the standard implementation of Registry.
type DefaultRegistry struct {
	modulesMu sync.RWMutex
	modules   map[string]Module
	logger    logging.Logger
}

// Ensure DefaultRegistry implements Registry.
var _ Registry = (*DefaultRegistry)(nil)

// NewRegistry creates a new module registry.
func NewRegistry(logger logging.Logger) Registry {
	return &DefaultRegistry{
		modules: make(map[string]Module),
		logger:  logger,
	}
}

// RegisterModule adds a module to the registry.
func (r *DefaultRegistry) RegisterModule(m Module) {
	r.modulesMu.Lock()
	defer r.modulesMu.Unlock()

	if _, exists := r.modules[m.Name()]; exists {
		r.logger.Error("module registered twice", "name", m.Name())
		return
	}

	r.modules[m.Name()] = m
	r.logger.Info("module registered", "name", m.Name())
}

// GetModules returns a copy of the registered modules map.
func (r *DefaultRegistry) GetModules() map[string]Module {
	r.modulesMu.RLock()
	defer r.modulesMu.RUnlock()

	modulesCopy := make(map[string]Module, len(r.modules))
	for name, mod := range r.modules {
		modulesCopy[name] = mod
	}

	return modulesCopy
}

// MockRegistry is a mock implementation of Registry for testing.
type MockRegistry struct {
	modules map[string]Module
	logger  logging.Logger
}

// Ensure MockRegistry implements Registry.
var _ Registry = (*MockRegistry)(nil)

// NewMockRegistry creates a new mock module registry.
func NewMockRegistry(logger logging.Logger) *MockRegistry {
	return &MockRegistry{
		modules: make(map[string]Module),
		logger:  logger,
	}
}

// RegisterModule adds a module to the registry.
func (r *MockRegistry) RegisterModule(m Module) {
	r.modules[m.Name()] = m
	r.logger.Info("mock module registered", "name", m.Name())
}

// GetModules returns the registered modules map.
func (r *MockRegistry) GetModules() map[string]Module {
	return r.modules
}
