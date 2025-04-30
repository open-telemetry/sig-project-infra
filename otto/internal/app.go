// SPDX-License-Identifier: Apache-2.0

// Package internal provides shared infrastructure for Otto.
package internal

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/go-github/v71/github"
	"github.com/jferrl/go-githubauth"
	"github.com/open-telemetry/sig-project-infra/otto/internal/config"
	"github.com/open-telemetry/sig-project-infra/otto/internal/secrets"
	"golang.org/x/oauth2"
)

// App encapsulates all application dependencies.
type App struct {
	Config         *config.AppConfig
	Secrets        secrets.Manager
	Database       *Database
	Telemetry      *TelemetryManager
	Logger         *slog.Logger
	Addr           string
	GitHubClient   *github.Client // GitHub API client for interacting with GitHub
	ModuleRegistry *ModuleRegistry
	server         *Server
	shutdownSignal chan struct{}
}

// NewApp creates and initializes a new application instance.
//
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - configPath: The file path to the application's configuration file.
//   - secretsPath: The file path to the secrets file used for managing sensitive data.
func NewApp(ctx context.Context, configPath, secretsPath string) (*App, error) {
	// Load configuration
	appConfig, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	// Load secrets
	secretsManager, err := secrets.LoadSecrets(secretsPath)
	if err != nil {
		return nil, err
	}

	// Initialize app with config and empty module registry
	app := &App{
		Config:         appConfig,
		Secrets:        secretsManager,
		Addr:           appConfig.Port,
		ModuleRegistry: NewModuleRegistry(),
		shutdownSignal: make(chan struct{}),
	}

	// Initialize GitHub client
	if err := app.initializeGitHubClient(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub client: %w", err)
	}

	// Initialize telemetry
	app.Telemetry, err = NewTelemetryManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Get logger from telemetry
	app.Logger = app.Telemetry.Logger

	// Initialize database
	app.Database, err = NewDatabase(app.Config.DBPath)
	if err != nil {
		return nil, err
	}

	// Create HTTP server with app reference
	app.server = NewServerWithApp(app.Addr, app.Secrets, app)

	return app, nil
}

// Start begins all application services.
func (a *App) Start(ctx context.Context) error {
	// Initialize and start all modules
	if err := a.initializeModules(ctx); err != nil {
		return err
	}

	// Start HTTP server (non-blocking)
	go func() {
		if err := a.server.Start(); err != nil {
			a.Logger.Error("Server error", "err", err)
		}
	}()

	a.Logger.Info("otto started", "addr", a.Addr)
	return nil
}

// Shutdown gracefully stops all application services.
func (a *App) Shutdown(ctx context.Context) error {
	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		a.Logger.Error("Error during server shutdown", "err", err)
	}

	// Shutdown modules
	if err := a.shutdownModules(ctx); err != nil {
		a.Logger.Error("Error during module shutdown", "err", err)
	}

	// Shutdown telemetry
	if a.Telemetry != nil {
		if err := a.Telemetry.Shutdown(ctx); err != nil {
			a.Logger.Error("Error during telemetry shutdown", "err", err)
		}
	}

	// Close database
	if a.Database != nil {
		if err := a.Database.Close(); err != nil {
			a.Logger.Error("Error closing database", "err", err)
		}
	}

	return nil
}

// WaitForShutdown blocks until the application is signaled to shut down.
func (a *App) WaitForShutdown() {
	<-a.shutdownSignal
}

// SignalShutdown triggers the application to begin shutting down.
func (a *App) SignalShutdown() {
	close(a.shutdownSignal)
}

// RegisterModule registers a module with this app instance.
func (a *App) RegisterModule(m Module) {
	a.ModuleRegistry.RegisterModule(m)
}

// GetModules returns all registered modules for this app instance.
func (a *App) GetModules() map[string]Module {
	return a.ModuleRegistry.GetModules()
}

// initializeModules initializes all registered modules.
func (a *App) initializeModules(ctx context.Context) error {
	// Get all registered modules
	modules := a.ModuleRegistry.GetModules()

	for name, mod := range modules {
		if initializer, ok := mod.(ModuleInitializer); ok {
			if err := initializer.Initialize(ctx, a); err != nil {
				a.Logger.Error("Failed to initialize module", "name", name, "err", err)
				return err
			}
		}
	}
	return nil
}

// shutdownModules gracefully shuts down all modules.
func (a *App) shutdownModules(ctx context.Context) error {
	// Get all registered modules
	modules := a.ModuleRegistry.GetModules()

	var wg sync.WaitGroup
	errors := make(chan error, len(modules))

	for name, mod := range modules {
		if shutdowner, ok := mod.(ModuleShutdowner); ok {
			wg.Add(1)
			go func(n string, m ModuleShutdowner) {
				defer wg.Done()
				if err := m.Shutdown(ctx); err != nil {
					a.Logger.Error("Module shutdown error", "name", n, "err", err)
					errors <- err
				}
			}(name, shutdowner)
		}
	}

	wg.Wait()
	close(errors)

	// Return the first error if any
	for err := range errors {
		return err
	}

	return nil
}

// Command handling has been removed since commands are processed through events

// DispatchEvent hands an event to all modules.
func (a *App) DispatchEvent(eventType string, event any, raw []byte) {
	// Get all registered modules
	modules := a.ModuleRegistry.GetModules()

	for name, mod := range modules {
		go func(n string, m Module) {
			if err := m.HandleEvent(eventType, event, raw); err != nil {
				a.Logger.Error("Event handling error", "module", n, "event", eventType, "err", err)
			}
		}(name, mod)
	}
}

// initializeGitHubClient sets up the GitHub API client with proper authentication.
func (a *App) initializeGitHubClient(ctx context.Context) error {
	// Check if GitHub App authentication is configured
	appID := a.Secrets.GetGitHubAppID()
	installID := a.Secrets.GetGitHubInstallationID()
	privateKey := a.Secrets.GetGitHubPrivateKey()

	if appID > 0 || installID > 0 || len(privateKey) > 0 {
		// Check for partially provided credentials
		if appID <= 0 || installID <= 0 || len(privateKey) == 0 {
			return fmt.Errorf("incomplete GitHub App credentials provided: appID=%d, installID=%d, privateKeyLength=%d",
				appID, installID, len(privateKey))
		}
		// Use GitHub App authentication
		appTokenSource, err := githubauth.NewApplicationTokenSource(appID, privateKey)
		if err != nil {
			return fmt.Errorf("failed to create GitHub app token source: %w", err)
		}

		installationTokenSource := githubauth.NewInstallationTokenSource(installID, appTokenSource)

		// Create an HTTP client that uses the installation token
		httpClient := oauth2.NewClient(ctx, installationTokenSource)

		// Create a new GitHub client with the custom HTTP client
		a.GitHubClient = github.NewClient(httpClient)
		slog.Info("GitHub client initialized with GitHub App authentication",
			"app_id", appID,
			"installation_id", installID)
	} else {
		// If no authentication configured, use unauthenticated client
		a.GitHubClient = github.NewClient(nil)
		slog.Info("GitHub client initialized (no auth)")
	}

	return nil
}
