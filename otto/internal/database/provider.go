// SPDX-License-Identifier: Apache-2.0

// Package database provides database access interfaces and implementations.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Import sqlite driver for database/sql.
	_ "github.com/mattn/go-sqlite3"
)

// Provider defines the interface for database operations.
type Provider interface {
	// DB returns the underlying database connection.
	DB() *sql.DB
	// Sqlx returns the sqlx database connection.
	Sqlx() *sqlx.DB
	// Close closes the database connection.
	Close() error
	// Ping checks the database connection.
	Ping(ctx context.Context) error
}

// SQLiteProvider implements the Provider interface for SQLite databases.
type SQLiteProvider struct {
	db   *sql.DB
	sqlx *sqlx.DB
}

// Ensure SQLiteProvider implements Provider.
var _ Provider = (*SQLiteProvider)(nil)

// NewSQLiteProvider creates a new SQLite database provider.
func NewSQLiteProvider(dbPath string) (Provider, error) {
	// Open the database with sqlx
	sqlxDB, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := sqlxDB.Ping(); err != nil {
		sqlxDB.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure database
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(25)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)

	return &SQLiteProvider{
		db:   sqlxDB.DB,
		sqlx: sqlxDB,
	}, nil
}

// DB returns the underlying database connection.
func (p *SQLiteProvider) DB() *sql.DB {
	return p.db
}

// Sqlx returns the sqlx database connection.
func (p *SQLiteProvider) Sqlx() *sqlx.DB {
	return p.sqlx
}

// Close closes the database connection.
func (p *SQLiteProvider) Close() error {
	if p.sqlx != nil {
		return p.sqlx.Close()
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
	db   *sql.DB
	sqlx *sqlx.DB
}

// Ensure MockProvider implements Provider.
var _ Provider = (*MockProvider)(nil)

// NewMockProvider creates a new in-memory SQLite database for testing.
func NewMockProvider() (Provider, error) {
	// Open the database with sqlx
	sqlxDB, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory database: %w", err)
	}

	// Verify connection
	if err := sqlxDB.Ping(); err != nil {
		sqlxDB.Close()
		return nil, fmt.Errorf("failed to connect to in-memory database: %w", err)
	}

	return &MockProvider{
		db:   sqlxDB.DB,
		sqlx: sqlxDB,
	}, nil
}

// DB returns the underlying database connection.
func (p *MockProvider) DB() *sql.DB {
	return p.db
}

// Sqlx returns the sqlx database connection.
func (p *MockProvider) Sqlx() *sqlx.DB {
	return p.sqlx
}

// Close closes the database connection.
func (p *MockProvider) Close() error {
	if p.sqlx != nil {
		return p.sqlx.Close()
	}
	return nil
}

// Ping checks the database connection.
func (p *MockProvider) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}