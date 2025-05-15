// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"time"
)

// OnCallUser represents a user who can be assigned to on-call rotations.
type OnCallUser struct {
	ID             string    `db:"id" json:"id"`
	GitHubUsername string    `db:"github_username" json:"github_username"`
	Name           string    `db:"name" json:"name"`
	Email          string    `db:"email" json:"email"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (u OnCallUser) GetID() string {
	return u.ID
}

// OnCallRotation represents an on-call rotation for a specific repository.
type OnCallRotation struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Repository  string    `db:"repository" json:"repository"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (r OnCallRotation) GetID() string {
	return r.ID
}

// OnCallAssignment represents a user's assignment to an on-call rotation.
type OnCallAssignment struct {
	ID         string    `db:"id" json:"id"`
	RotationID string    `db:"rotation_id" json:"rotation_id"`
	UserID     string    `db:"user_id" json:"user_id"`
	StartTime  time.Time `db:"start_time" json:"start_time"`
	EndTime    time.Time `db:"end_time" json:"end_time"`
	IsCurrent  bool      `db:"is_current" json:"is_current"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (a OnCallAssignment) GetID() string {
	return a.ID
}

// OnCallEscalation represents an escalation of an issue or PR to the on-call user.
type OnCallEscalation struct {
	ID             string         `db:"id" json:"id"`
	AssignmentID   string         `db:"assignment_id" json:"assignment_id"`
	IssueNumber    int            `db:"issue_number" json:"issue_number,omitempty"`
	PRNumber       int            `db:"pr_number" json:"pr_number,omitempty"`
	Repository     string         `db:"repository" json:"repository"`
	Status         EscalationStatus `db:"status" json:"status"`
	EscalationTime time.Time      `db:"escalation_time" json:"escalation_time,omitempty"`
	ResolutionTime time.Time      `db:"resolution_time" json:"resolution_time,omitempty"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (e OnCallEscalation) GetID() string {
	return e.ID
}

// EscalationStatus defines the possible states of an escalation.
type EscalationStatus string

const (
	// StatusPending indicates the escalation is pending attention.
	StatusPending EscalationStatus = "pending"
	// StatusAcknowledged indicates the escalation has been acknowledged.
	StatusAcknowledged EscalationStatus = "acknowledged"
	// StatusResolved indicates the escalation has been resolved.
	StatusResolved EscalationStatus = "resolved"
	// StatusEscalated indicates the escalation has been escalated to another user.
	StatusEscalated EscalationStatus = "escalated"
)