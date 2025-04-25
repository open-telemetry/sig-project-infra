// SPDX-License-Identifier: Apache-2.0

// modules.go defines the interface and registry for Otto feature modules.

package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
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
	App      *App   // reference to the app instance
}

// Module is the Otto feature/module interface.
type Module interface {
	Name() string
	HandleEvent(eventType string, event any, raw json.RawMessage) error
}

// ModuleInitializer is an optional interface that modules can implement
// for initialization logic.
type ModuleInitializer interface {
	Initialize(ctx context.Context, app *App) error
}

// ModuleShutdowner is an optional interface that modules can implement
// for graceful shutdown.
type ModuleShutdowner interface {
	Shutdown(ctx context.Context) error
}

var (
	modulesMu sync.RWMutex
	modules   = make(map[string]Module)
)

// RegisterModule adds a module to the global registry.
func RegisterModule(m Module) {
	modulesMu.Lock()
	defer modulesMu.Unlock()
	if _, exists := modules[m.Name()]; exists {
		slog.Error("module registered twice", "name", m.Name())
		return
	}
	modules[m.Name()] = m
	slog.Info("module registered", "name", m.Name())
}

// GetModules returns a copy of the registered modules map
func GetModules() map[string]Module {
	modulesMu.RLock()
	defer modulesMu.RUnlock()

	modulesCopy := make(map[string]Module, len(modules))
	for name, mod := range modules {
		modulesCopy[name] = mod
	}
	return modulesCopy
}
