// SPDX-License-Identifier: Apache-2.0

package database

// BaseEntity provides a basic implementation of the Entity interface.
// It can be embedded in other structs to fulfill the Entity interface.
type BaseEntity struct {
	ID string `db:"id"`
}

// GetID returns the entity's ID.
func (e BaseEntity) GetID() string {
	return e.ID
}

// AnyEntity wraps any value to satisfy the Entity interface.
// This is primarily used for type compatibility in generic contexts.
type AnyEntity struct {
	BaseEntity
	Value any
}

// NewAnyEntity creates a new AnyEntity with the given ID and value.
func NewAnyEntity(id string, value any) *AnyEntity {
	return &AnyEntity{
		BaseEntity: BaseEntity{ID: id},
		Value:      value,
	}
}