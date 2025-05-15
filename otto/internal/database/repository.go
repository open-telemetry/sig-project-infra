// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
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

// SQLXRepository implements Repository using sqlx for database operations.
type SQLXRepository[T Entity] struct {
	db        Provider
	tableName string
}

// Ensure SQLXRepository implements Repository.
var _ Repository[Entity] = (*SQLXRepository[Entity])(nil)

// validateTableName ensures table names are safe for use in SQL queries.
func validateTableName(tableName string) error {
	// Restrict table names to alphanumeric characters, underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validPattern.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	return nil
}

// NewSQLXRepository creates a new repository using sqlx.
func NewSQLXRepository[T Entity](db Provider, tableName string) (Repository[T], error) {
	// Validate table name at initialization time
	if err := validateTableName(tableName); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &SQLXRepository[T]{
		db:        db,
		tableName: tableName,
	}, nil
}

// FindByID retrieves an entity by its ID.
func (r *SQLXRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", r.tableName)

	var entity T
	err := r.db.Sqlx().GetContext(ctx, &entity, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding entity with id %s: %w", id, err)
	}

	return &entity, nil
}

// FindAll retrieves all entities.
func (r *SQLXRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("SELECT * FROM %s", r.tableName)

	var entities []T
	err := r.db.Sqlx().SelectContext(ctx, &entities, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all entities: %w", err)
	}

	return entities, nil
}

// Create inserts a new entity.
func (r *SQLXRepository[T]) Create(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use sqlx's NamedExec to handle mapping struct fields to columns
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("INSERT INTO %s VALUES (:id, :github_username, :name, :email, :is_active, :created_at, :updated_at)", r.tableName)

	_, err := r.db.Sqlx().NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("error creating entity: %w", err)
	}

	return nil
}

// Update modifies an existing entity.
func (r *SQLXRepository[T]) Update(ctx context.Context, entity *T) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use sqlx's NamedExec to handle mapping struct fields to columns
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("UPDATE %s SET github_username=:github_username, name=:name, email=:email, is_active=:is_active, updated_at=:updated_at WHERE id=:id", r.tableName)

	_, err := r.db.Sqlx().NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("error updating entity: %w", err)
	}

	return nil
}

// Delete removes an entity by its ID.
func (r *SQLXRepository[T]) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.tableName)

	_, err := r.db.Sqlx().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting entity with id %s: %w", id, err)
	}

	return nil
}

// Transaction executes a function within a database transaction.
func (r *SQLXRepository[T]) Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sqlxTx, err := r.db.Sqlx().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	txRepo := &SQLXTransaction[T]{
		tx:        sqlxTx,
		tableName: r.tableName,
	}

	if err := fn(txRepo); err != nil {
		// Attempt rollback, but don't override original error
		rollbackErr := sqlxTx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback failed: %w)", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := sqlxTx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// SQLXTransaction implements Transaction for sqlx.
type SQLXTransaction[T Entity] struct {
	tx        *sqlx.Tx
	tableName string
}

// Ensure SQLXTransaction implements Transaction.
var _ Transaction[Entity] = (*SQLXTransaction[Entity])(nil)

// FindByID retrieves an entity by its ID within the transaction.
func (t *SQLXTransaction[T]) FindByID(ctx context.Context, id string) (*T, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", t.tableName)

	var entity T
	err := t.tx.GetContext(ctx, &entity, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entity not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding entity with id %s in transaction: %w", id, err)
	}

	return &entity, nil
}

// Create inserts a new entity within the transaction.
func (t *SQLXTransaction[T]) Create(ctx context.Context, entity *T) error {
	// Use sqlx's NamedExec to handle mapping struct fields to columns
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("INSERT INTO %s VALUES (:id, :github_username, :name, :email, :is_active, :created_at, :updated_at)", t.tableName)

	_, err := t.tx.NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("error creating entity in transaction: %w", err)
	}

	return nil
}

// Update modifies an existing entity within the transaction.
func (t *SQLXTransaction[T]) Update(ctx context.Context, entity *T) error {
	// Use sqlx's NamedExec to handle mapping struct fields to columns
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("UPDATE %s SET github_username=:github_username, name=:name, email=:email, is_active=:is_active, updated_at=:updated_at WHERE id=:id", t.tableName)

	_, err := t.tx.NamedExecContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("error updating entity in transaction: %w", err)
	}

	return nil
}

// Delete removes an entity by its ID within the transaction.
func (t *SQLXTransaction[T]) Delete(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", t.tableName)

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting entity with id %s in transaction: %w", id, err)
	}

	return nil
}