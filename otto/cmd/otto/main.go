// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the Otto application.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-telemetry/sig-project-infra/otto/internal"
	"github.com/open-telemetry/sig-project-infra/otto/internal/config"
	"github.com/open-telemetry/sig-project-infra/otto/modules" // Importing modules for explicit registration
)

func main() {
	// Create root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration paths from environment
	configPath := config.GetEnvOrDefault("OTTO_CONFIG", "config.yaml")
	secretsPath := config.GetEnvOrDefault("OTTO_SECRETS", "secrets.yaml")

	// App will load the configuration internally

	// Create and initialize application
	app, err := internal.NewApp(ctx, configPath, secretsPath)
	if err != nil {
		slog.Error("Failed to initialize application", "err", err)
		os.Exit(1)
	}

	// Register modules explicitly
	app.RegisterModule(&modules.OnCallModule{})

	// Start the application
	if err := app.Start(ctx); err != nil {
		slog.Error("Failed to start application", "err", err)
		os.Exit(1)
	}

	// Set up graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	slog.Info("otto is running, press Ctrl+C to stop")
	<-done
	slog.Info("shutdown signal received")

	// Create shutdown context with timeout
	ctxShutdown, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	// Gracefully shut down the application
	if err := app.Shutdown(ctxShutdown); err != nil {
		slog.Error("Error during application shutdown", "err", err)
	}

	slog.Info("otto has been gracefully shut down")
}
