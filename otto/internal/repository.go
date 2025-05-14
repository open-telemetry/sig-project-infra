// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"database/sql"
	"time"
)

// Repository provides an interface for database access.
type Repository interface {
	// Common repository methods
	Ping(ctx context.Context) error
	Close() error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction.
type Transaction interface {
	Commit() error
	Rollback() error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLiteRepository implements Repository for SQLite databases.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository.
func NewSQLiteRepository(db *sql.DB) Repository {
	return &SQLiteRepository{db: db}
}

// Ping checks database connectivity.
func (r *SQLiteRepository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.PingContext(ctx)
	if err != nil {
		return LogAndWrapError(err, ErrorTypeDatabase, "ping", nil)
	}
	return nil
}

// Close closes the database connection.
func (r *SQLiteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return LogAndWrapError(err, ErrorTypeDatabase, "close", nil)
	}
	return nil
}

// Exec executes a query without returning rows.
func (r *SQLiteRepository) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, LogAndWrapError(err, ErrorTypeDatabase, "exec", map[string]any{
			"query": truncateQuery(query),
		})
	}
	return resp, nil
}

// Query executes a query that returns rows.
func (r *SQLiteRepository) Query(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, LogAndWrapError(err, ErrorTypeDatabase, "query", map[string]any{
			"query": truncateQuery(query),
		})
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (r *SQLiteRepository) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return r.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a new transaction.
func (r *SQLiteRepository) BeginTx(ctx context.Context) (Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, LogAndWrapError(err, ErrorTypeDatabase, "begin_transaction", nil)
	}
	return &SQLiteTransaction{tx: tx}, nil
}

// SQLiteTransaction implements Transaction for SQLite.
type SQLiteTransaction struct {
	tx *sql.Tx
}

// Commit commits the transaction.
func (t *SQLiteTransaction) Commit() error {
	err := t.tx.Commit()
	if err != nil {
		return LogAndWrapError(err, ErrorTypeDatabase, "commit_transaction", nil)
	}
	return nil
}

// Rollback rolls back the transaction.
func (t *SQLiteTransaction) Rollback() error {
	err := t.tx.Rollback()
	if err != nil {
		return LogAndWrapError(err, ErrorTypeDatabase, "rollback_transaction", nil)
	}
	return nil
}

// Exec executes a query within the transaction.
func (t *SQLiteTransaction) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, LogAndWrapError(err, ErrorTypeDatabase, "tx_exec", map[string]any{
			"query": truncateQuery(query),
		})
	}
	return resp, nil
}

// Query executes a query that returns rows within the transaction.
func (t *SQLiteTransaction) Query(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, LogAndWrapError(err, ErrorTypeDatabase, "tx_query", map[string]any{
			"query": truncateQuery(query),
		})
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row within the transaction.
func (t *SQLiteTransaction) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return t.tx.QueryRowContext(ctx, query, args...)
}

// truncateQuery trims a query string for logging purposes.
func truncateQuery(query string) string {
	const maxLen = 100
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "...(truncated)"
}
