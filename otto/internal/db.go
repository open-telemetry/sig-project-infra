// SPDX-License-Identifier: Apache-2.0

// db.go sets up otto's shared SQLite connection.

package internal

import (
	"database/sql"
	"fmt"
	"sync"

	// Import sqlite driver for database/sql.
	_ "github.com/mattn/go-sqlite3"

	"github.com/open-telemetry/sig-project-infra/otto/internal/config"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// InitDB opens the global DB connection using the configured database path.
func InitDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		dbPath := config.Global.DBPath
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			err = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// Verify connection
		if pingErr := db.Ping(); pingErr != nil {
			db.Close()
			err = fmt.Errorf("failed to connect to database: %w", pingErr)
			return
		}
	})
	return db, err
}

// GetDB returns the shared *sql.DB.
func GetDB() *sql.DB {
	return db
}

// OpenDB opens a new database connection with the given path.
// Use this for tests or when you need a separate connection.
func OpenDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
