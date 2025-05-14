// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	// Import file source driver for migrations.
	// This is required for the migrate.NewWithDatabaseInstance to work with the file:// scheme.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigratorProvider defines the interface for database migration operations.
type MigratorProvider interface {
	// MigrateUp applies all pending migrations.
	MigrateUp(ctx context.Context) error
	// MigrateDown rolls back all migrations.
	MigrateDown(ctx context.Context) error
	// MigrateTo migrates to a specific version.
	MigrateTo(ctx context.Context, version uint) error
	// GetVersion returns the current migration version.
	GetVersion(ctx context.Context) (uint, bool, error)
}

// SQLiteMigrator implements MigratorProvider for SQLite.
type SQLiteMigrator struct {
	db         Provider
	migrateDir string
	logger     *slog.Logger
}

// Ensure SQLiteMigrator implements MigratorProvider.
var _ MigratorProvider = (*SQLiteMigrator)(nil)

// NewSQLiteMigrator creates a new migrator for SQLite.
func NewSQLiteMigrator(db Provider, migrateDir string, logger *slog.Logger) MigratorProvider {
	if logger == nil {
		logger = slog.Default()
	}

	return &SQLiteMigrator{
		db:         db,
		migrateDir: migrateDir,
		logger:     logger,
	}
}

// createMigrate creates a new migrate instance.
func (m *SQLiteMigrator) createMigrate() (*migrate.Migrate, error) {
	driver, err := sqlite3.WithInstance(m.db.DB(), &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite driver: %w", err)
	}

	// Use file URL scheme for migrations directory
	sourceURL := "file://" + m.migrateDir

	// Create migrator using the file URL and database driver
	migrate, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return migrate, nil
}

// MigrateUp applies all pending migrations.
func (m *SQLiteMigrator) MigrateUp(ctx context.Context) error {
	// Create a new context with timeout that inherits from the original
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Log migration action
	m.logger.Info("Migrating database up", "dir", m.migrateDir)

	migrator, err := m.createMigrate()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to migrate up: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	m.logger.Info("Database migrated up successfully", "version", version, "dirty", dirty)
	return nil
}

// MigrateDown rolls back all migrations.
func (m *SQLiteMigrator) MigrateDown(ctx context.Context) error {
	// Create a new context with timeout that inherits from the original
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Log migration action
	m.logger.Info("Migrating database down", "dir", m.migrateDir)

	migrator, err := m.createMigrate()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to migrate down: %w", err)
	}

	m.logger.Info("Database migrated down successfully")
	return nil
}

// MigrateTo migrates to a specific version.
func (m *SQLiteMigrator) MigrateTo(ctx context.Context, version uint) error {
	// Create a new context with timeout that inherits from the original
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Log migration action
	m.logger.Info("Migrating database to version", "version", version, "dir", m.migrateDir)

	migrator, err := m.createMigrate()
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Migrate(version); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migration needed", "version", version)
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	m.logger.Info("Database migrated to version successfully", "version", version)
	return nil
}

// GetVersion returns the current migration version.
func (m *SQLiteMigrator) GetVersion(ctx context.Context) (uint, bool, error) {
	// Create a new context with timeout that inherits from the original
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	migrator, err := m.createMigrate()
	if err != nil {
		return 0, false, err
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// MockMigrator implements MigratorProvider for testing.
type MockMigrator struct {
	version uint
	dirty   bool
}

// Ensure MockMigrator implements MigratorProvider.
var _ MigratorProvider = (*MockMigrator)(nil)

// NewMockMigrator creates a new mock migrator.
func NewMockMigrator() *MockMigrator {
	return &MockMigrator{
		version: 0,
		dirty:   false,
	}
}

// MigrateUp applies all pending migrations.
func (m *MockMigrator) MigrateUp(ctx context.Context) error {
	m.version = 999 // Arbitrary high number to simulate all migrations applied
	m.dirty = false
	return nil
}

// MigrateDown rolls back all migrations.
func (m *MockMigrator) MigrateDown(ctx context.Context) error {
	m.version = 0
	m.dirty = false
	return nil
}

// MigrateTo migrates to a specific version.
func (m *MockMigrator) MigrateTo(ctx context.Context, version uint) error {
	m.version = version
	m.dirty = false
	return nil
}

// GetVersion returns the current migration version.
func (m *MockMigrator) GetVersion(ctx context.Context) (uint, bool, error) {
	return m.version, m.dirty, nil
}
