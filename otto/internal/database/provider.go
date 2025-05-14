// SPDX-License-Identifier: Apache-2.0

// Package database provides database access interfaces and implementations.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Import sqlite driver for database/sql.
	_ "github.com/mattn/go-sqlite3"
)

// Provider defines the interface for database operations.
type Provider interface {
	// DB returns the underlying database connection.
	DB() *sql.DB
	// Close closes the database connection.
	Close() error
	// Ping checks the database connection.
	Ping(ctx context.Context) error
}

// SQLiteProvider implements the Provider interface for SQLite databases.
type SQLiteProvider struct {
	db *sql.DB
}

// Ensure SQLiteProvider implements Provider.
var _ Provider = (*SQLiteProvider)(nil)

// NewSQLiteProvider creates a new SQLite database provider.
func NewSQLiteProvider(dbPath string) (Provider, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure database
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &SQLiteProvider{db: db}, nil
}

// DB returns the underlying database connection.
func (p *SQLiteProvider) DB() *sql.DB {
	return p.db
}

// Close closes the database connection.
func (p *SQLiteProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks the database connection.
func (p *SQLiteProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}

// MockProvider implements Provider for testing.
type MockProvider struct {
	db *sql.DB
}

// Ensure MockProvider implements Provider.
var _ Provider = (*MockProvider)(nil)

// NewMockProvider creates a new in-memory SQLite database for testing.
func NewMockProvider() (Provider, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to in-memory database: %w", err)
	}

	return &MockProvider{db: db}, nil
}

// DB returns the underlying database connection.
func (p *MockProvider) DB() *sql.DB {
	return p.db
}

// Close closes the database connection.
func (p *MockProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks the database connection.
func (p *MockProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}
