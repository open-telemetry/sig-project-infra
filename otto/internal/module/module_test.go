// SPDX-License-Identifier: Apache-2.0

package module

import (
	"testing"

	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
)

func TestRegistryFunctions(t *testing.T) {
	// Create a mock logger for testing
	logger := logging.NewNoopLogger()

	// Create a registry
	registry := NewRegistry(logger)

	// Create a mock module
	mockRepo := database.NewMockAnyRepository()
	deps := Dependencies{
		Logger:     logger,
		Repository: mockRepo,
	}
	module := NewMockModule("test-module", deps)

	// Register the module
	registry.RegisterModule(module)

	// Test GetModules
	modules := registry.GetModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}

	if _, ok := modules["test-module"]; !ok {
		t.Errorf("Expected to find module with name 'test-module'")
	}

	// Try registering the same module again (should log a warning and not add it)
	registry.RegisterModule(module)
	modules = registry.GetModules()
	if len(modules) != 1 {
		t.Errorf("Expected still 1 module after duplicate registration, got %d", len(modules))
	}
}

func TestMockModuleFunctions(t *testing.T) {
	// Create a mock logger for testing
	logger := logging.NewNoopLogger()

	// Create mock dependencies
	mockRepo := database.NewMockAnyRepository()
	deps := Dependencies{
		Logger:     logger,
		Repository: mockRepo,
	}

	// Create a mock module
	module := NewMockModule("test-module", deps)

	// Verify name
	if module.Name() != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", module.Name())
	}

	// Test HandleEvent
	err := module.HandleEvent("test-event", nil, nil)
	if err != nil {
		t.Errorf("HandleEvent returned unexpected error: %v", err)
	}

	if !module.WasEventHandled("test-event") {
		t.Error("WasEventHandled should return true for 'test-event'")
	}

	if module.WasEventHandled("unknown-event") {
		t.Error("WasEventHandled should return false for 'unknown-event'")
	}

	// Test Initialize with the test context
	err = module.Initialize(t.Context())
	if err != nil {
		t.Errorf("Initialize returned unexpected error: %v", err)
	}
	if !module.InitializeCalled {
		t.Error("InitializeCalled should be true after Initialize")
	}

	// Test Shutdown with the test context
	err = module.Shutdown(t.Context())
	if err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
	if !module.ShutdownCalled {
		t.Error("ShutdownCalled should be true after Shutdown")
	}
}
