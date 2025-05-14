// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"fmt"
	"sync"
)

// MockAnyRepository is an in-memory implementation of Repository for AnyEntity.
type MockAnyRepository struct {
	entities map[string]*AnyEntity
	mu       sync.RWMutex
}

// NewMockAnyRepository creates a new in-memory repository for AnyEntity.
func NewMockAnyRepository() *MockAnyRepository {
	return &MockAnyRepository{
		entities: make(map[string]*AnyEntity),
	}
}

// FindByID retrieves an entity by its ID.
func (r *MockAnyRepository) FindByID(ctx context.Context, id string) (*AnyEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.entities[id]
	if !ok {
		return nil, fmt.Errorf("entity not found with id %s", id)
	}
	return entity, nil
}

// FindAll retrieves all entities.
func (r *MockAnyRepository) FindAll(ctx context.Context) ([]AnyEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := make([]AnyEntity, 0, len(r.entities))
	for _, entity := range r.entities {
		entities = append(entities, *entity)
	}
	return entities, nil
}

// Create inserts a new entity.
func (r *MockAnyRepository) Create(ctx context.Context, entity *AnyEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entities[entity.GetID()]; exists {
		return fmt.Errorf("entity with id %s already exists", entity.GetID())
	}
	r.entities[entity.GetID()] = entity
	return nil
}

// Update modifies an existing entity.
func (r *MockAnyRepository) Update(ctx context.Context, entity *AnyEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entities[entity.GetID()]; !exists {
		return fmt.Errorf("entity with id %s not found", entity.GetID())
	}
	r.entities[entity.GetID()] = entity
	return nil
}

// Delete removes an entity by its ID.
func (r *MockAnyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entities[id]; !exists {
		return fmt.Errorf("entity with id %s not found", id)
	}
	delete(r.entities, id)
	return nil
}

// Transaction executes a function within a transaction.
func (r *MockAnyRepository) Transaction(ctx context.Context, fn func(tx Transaction[AnyEntity]) error) error {
	tx := &MockAnyTransaction{
		repo: r,
	}
	return fn(tx)
}

// MockAnyTransaction is an in-memory implementation of Transaction for AnyEntity.
type MockAnyTransaction struct {
	repo *MockAnyRepository
}

// FindByID retrieves an entity by its ID within the transaction.
func (t *MockAnyTransaction) FindByID(ctx context.Context, id string) (*AnyEntity, error) {
	return t.repo.FindByID(ctx, id)
}

// Create inserts a new entity within the transaction.
func (t *MockAnyTransaction) Create(ctx context.Context, entity *AnyEntity) error {
	return t.repo.Create(ctx, entity)
}

// Update modifies an existing entity within the transaction.
func (t *MockAnyTransaction) Update(ctx context.Context, entity *AnyEntity) error {
	return t.repo.Update(ctx, entity)
}

// Delete removes an entity by its ID within the transaction.
func (t *MockAnyTransaction) Delete(ctx context.Context, id string) error {
	return t.repo.Delete(ctx, id)
}
