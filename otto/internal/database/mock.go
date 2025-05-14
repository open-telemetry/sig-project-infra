// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

// MockRepository provides an in-memory implementation of Repository for testing.
type MockRepository[T Entity] struct {
	entities map[string]T
	mu       sync.RWMutex
}

// Ensure MockRepository implements Repository.
var _ Repository[Entity] = (*MockRepository[Entity])(nil)

// NewMockRepository creates a new in-memory repository for testing.
func NewMockRepository[T Entity]() Repository[T] {
	return &MockRepository[T]{
		entities: make(map[string]T),
	}
}

// FindByID retrieves an entity by its ID.
func (r *MockRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found with id %s: %w", id, sql.ErrNoRows)
	}

	return &entity, nil
}

// FindAll retrieves all entities.
func (r *MockRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, 0, len(r.entities))
	for _, entity := range r.entities {
		result = append(result, entity)
	}

	return result, nil
}

// Create inserts a new entity.
func (r *MockRepository[T]) Create(ctx context.Context, entity *T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := (*entity).GetID()
	if _, exists := r.entities[id]; exists {
		return fmt.Errorf("entity with id %s already exists", id)
	}

	r.entities[id] = *entity
	return nil
}

// Update modifies an existing entity.
func (r *MockRepository[T]) Update(ctx context.Context, entity *T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := (*entity).GetID()
	if _, exists := r.entities[id]; !exists {
		return fmt.Errorf("entity with id %s does not exist", id)
	}

	r.entities[id] = *entity
	return nil
}

// Delete removes an entity by its ID.
func (r *MockRepository[T]) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entities[id]; !exists {
		return fmt.Errorf("entity with id %s does not exist", id)
	}

	delete(r.entities, id)
	return nil
}

// Transaction executes a function within a mock transaction.
func (r *MockRepository[T]) Transaction(ctx context.Context, fn func(tx Transaction[T]) error) error {
	// Create a temporary copy of the repository to simulate a transaction
	tempRepo := &MockTransaction[T]{
		entities:     make(map[string]T),
		parentRepo:   r,
		shouldCommit: false,
	}

	// Copy current entities
	r.mu.RLock()
	for id, entity := range r.entities {
		tempRepo.entities[id] = entity
	}
	r.mu.RUnlock()

	// Execute the transaction function
	if err := fn(tempRepo); err != nil {
		return err
	}

	// Commit changes if no error
	tempRepo.shouldCommit = true
	return tempRepo.commit()
}

// MockTransaction simulates a database transaction in memory.
type MockTransaction[T Entity] struct {
	entities     map[string]T
	parentRepo   *MockRepository[T]
	shouldCommit bool
	mu           sync.RWMutex
}

// Ensure MockTransaction implements Transaction.
var _ Transaction[Entity] = (*MockTransaction[Entity])(nil)

// FindByID retrieves an entity by its ID within the transaction.
func (t *MockTransaction[T]) FindByID(ctx context.Context, id string) (*T, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entity, ok := t.entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found with id %s: %w", id, sql.ErrNoRows)
	}

	return &entity, nil
}

// Create inserts a new entity within the transaction.
func (t *MockTransaction[T]) Create(ctx context.Context, entity *T) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := (*entity).GetID()
	if _, exists := t.entities[id]; exists {
		return fmt.Errorf("entity with id %s already exists", id)
	}

	t.entities[id] = *entity
	return nil
}

// Update modifies an existing entity within the transaction.
func (t *MockTransaction[T]) Update(ctx context.Context, entity *T) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := (*entity).GetID()
	if _, exists := t.entities[id]; !exists {
		return fmt.Errorf("entity with id %s does not exist", id)
	}

	t.entities[id] = *entity
	return nil
}

// Delete removes an entity by its ID within the transaction.
func (t *MockTransaction[T]) Delete(ctx context.Context, id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.entities[id]; !exists {
		return fmt.Errorf("entity with id %s does not exist", id)
	}

	delete(t.entities, id)
	return nil
}

// commit applies all changes from the transaction to the parent repository.
func (t *MockTransaction[T]) commit() error {
	if !t.shouldCommit {
		return errors.New("transaction not committed")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	t.parentRepo.mu.Lock()
	defer t.parentRepo.mu.Unlock()

	// Replace entire entity map
	t.parentRepo.entities = make(map[string]T)
	for id, entity := range t.entities {
		t.parentRepo.entities[id] = entity
	}

	return nil
}
