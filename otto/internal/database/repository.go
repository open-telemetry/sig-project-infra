// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Entity represents a database entity with an ID.
type Entity interface {
	GetID() string
}

// Repository provides a generic interface for CRUD operations on entities.
type Repository[T Entity] interface {
	// FindByID retrieves an entity by its ID.
	FindByID(ctx context.Context, id string) (*T, error)
	// FindAll retrieves all entities.
	FindAll(ctx context.Context) ([]T, error)
	// Create inserts a new entity.
	Create(ctx context.Context, entity *T) error
	// Update modifies an existing entity.
	Update(ctx context.Context, entity *T) error
	// Delete removes an entity by its ID.
	Delete(ctx context.Context, id string) error
	// Transaction executes a function within a database transaction.
	Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error
}

// Transaction represents a database transaction for entity operations.
type Transaction[T Entity] interface {
	// FindByID retrieves an entity by its ID within the transaction.
	FindByID(ctx context.Context, id string) (*T, error)
	// Create inserts a new entity within the transaction.
	Create(ctx context.Context, entity *T) error
	// Update modifies an existing entity within the transaction.
	Update(ctx context.Context, entity *T) error
	// Delete removes an entity by its ID within the transaction.
	Delete(ctx context.Context, id string) error
}

// SQLMapper provides functions to map between entities and database rows.
type SQLMapper[T Entity] interface {
	// ToRow maps an entity to database column values.
	ToRow(entity *T) []interface{}
	// FromRow creates an entity from a database row.
	FromRow(row *sql.Row) (*T, error)
	// FromRows creates entities from database rows.
	FromRows(rows *sql.Rows) ([]T, error)
	// Columns returns the database column names.
	Columns() []string
	// Table returns the database table name.
	Table() string
}

// SQLRepository implements Repository using SQL operations.
type SQLRepository[T Entity] struct {
	db     Provider
	mapper SQLMapper[T]
}

// Ensure SQLRepository implements Repository.
var _ Repository[Entity] = (*SQLRepository[Entity])(nil)

// validateTableName ensures table names are safe for use in SQL queries.
func validateTableName(tableName string) error {
	// Restrict table names to alphanumeric characters, underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validPattern.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	return nil
}

// NewSQLRepository creates a new SQL repository.
func NewSQLRepository[T Entity](db Provider, mapper SQLMapper[T]) (Repository[T], error) {
	// Validate table name at initialization time
	tableName := mapper.Table()
	if err := validateTableName(tableName); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &SQLRepository[T]{
		db:     db,
		mapper: mapper,
	}, nil
}

// FindByID retrieves an entity by its ID.
func (r *SQLRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	columns := r.mapper.Columns()
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns(columns) + " FROM " + r.mapper.Table() + " WHERE id = ?"

	row := r.db.DB().QueryRowContext(ctx, query, id)
	entity, err := r.mapper.FromRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding entity with id %s: %w", id, err)
	}

	return entity, nil
}

// FindAll retrieves all entities.
func (r *SQLRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	columns := r.mapper.Columns()
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns(columns) + " FROM " + r.mapper.Table()

	rows, err := r.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all entities: %w", err)
	}
	defer rows.Close()

	return r.mapper.FromRows(rows)
}

// Create inserts a new entity.
func (r *SQLRepository[T]) Create(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	columns := r.mapper.Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + r.mapper.Table() + " (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := r.db.DB().ExecContext(ctx, query, r.mapper.ToRow(entity)...)
	if err != nil {
		return fmt.Errorf("error creating entity: %w", err)
	}

	return nil
}

// Update modifies an existing entity.
func (r *SQLRepository[T]) Update(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	columns := r.mapper.Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + r.mapper.Table() + " SET " + setClause + " WHERE id = ?"

	args := r.mapper.ToRow(entity)
	// Add ID as the last parameter for the WHERE clause
	args = append(args, (*entity).GetID())

	_, err := r.db.DB().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating entity: %w", err)
	}

	return nil
}

// Delete removes an entity by its ID.
func (r *SQLRepository[T]) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + r.mapper.Table() + " WHERE id = ?"

	_, err := r.db.DB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting entity with id %s: %w", id, err)
	}

	return nil
}

// Transaction executes a function within a database transaction.
func (r *SQLRepository[T]) Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sqlTx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	txRepo := &SQLTransaction[T]{
		tx:     sqlTx,
		mapper: r.mapper,
	}

	if err := fn(txRepo); err != nil {
		// Attempt rollback, but don't override original error
		rollbackErr := sqlTx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback failed: %w)", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// SQLTransaction implements Transaction for SQL.
type SQLTransaction[T Entity] struct {
	tx     *sql.Tx
	mapper SQLMapper[T]
}

// Ensure SQLTransaction implements Transaction.
var _ Transaction[Entity] = (*SQLTransaction[Entity])(nil)

// FindByID retrieves an entity by its ID within the transaction.
func (t *SQLTransaction[T]) FindByID(ctx context.Context, id string) (*T, error) {
	columns := t.mapper.Columns()
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns(columns) + " FROM " + t.mapper.Table() + " WHERE id = ?"

	row := t.tx.QueryRowContext(ctx, query, id)
	entity, err := t.mapper.FromRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding entity with id %s: %w", id, err)
	}

	return entity, nil
}

// Create inserts a new entity within the transaction.
func (t *SQLTransaction[T]) Create(ctx context.Context, entity *T) error {
	columns := t.mapper.Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + t.mapper.Table() + " (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := t.tx.ExecContext(ctx, query, t.mapper.ToRow(entity)...)
	if err != nil {
		return fmt.Errorf("error creating entity in transaction: %w", err)
	}

	return nil
}

// Update modifies an existing entity within the transaction.
func (t *SQLTransaction[T]) Update(ctx context.Context, entity *T) error {
	columns := t.mapper.Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + t.mapper.Table() + " SET " + setClause + " WHERE id = ?"

	args := t.mapper.ToRow(entity)
	// Add ID as the last parameter for the WHERE clause
	args = append(args, (*entity).GetID())

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating entity in transaction: %w", err)
	}

	return nil
}

// Delete removes an entity by its ID within the transaction.
func (t *SQLTransaction[T]) Delete(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + t.mapper.Table() + " WHERE id = ?"

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting entity with id %s in transaction: %w", id, err)
	}

	return nil
}

// Helper functions for SQL generation

func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

func createPlaceholders(count int) string {
	result := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ", "
		}
		result += "?"
	}
	return result
}

func createSetClause(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col + " = ?"
	}
	return result
}
