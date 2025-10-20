// SPDX-License-Identifier: Apache-2.0

// db.go sets up otto's shared SQLite connection.

package internal

import (
	"database/sql"
	"fmt"

	// Import sqlite driver for database/sql.
	_ "modernc.org/sqlite"
)

// Database encapsulates database connection management.
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection with the provided path.
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// DB returns the underlying database connection.
func (d *Database) DB() *sql.DB {
	return d.db
}

// OpenDB opens a new database connection with the given path.
// Use this for tests or when you need a separate connection.
// Deprecated: Use NewDatabase instead.
func OpenDB(dbPath string) (*sql.DB, error) {
	database, err := NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	return database.DB(), nil
}
