// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"time"
)

// OnCallUser represents a user who can be assigned to on-call rotations.
type OnCallUser struct {
	ID             string    `json:"id"`
	GitHubUsername string    `json:"github_username"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (u OnCallUser) GetID() string {
	return u.ID
}

// OnCallRotation represents an on-call rotation for a specific repository.
type OnCallRotation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Repository  string    `json:"repository"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (r OnCallRotation) GetID() string {
	return r.ID
}

// OnCallAssignment represents a user's assignment to an on-call rotation.
type OnCallAssignment struct {
	ID         string    `json:"id"`
	RotationID string    `json:"rotation_id"`
	UserID     string    `json:"user_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	IsCurrent  bool      `json:"is_current"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetID implements the database.Entity interface.
func (a OnCallAssignment) GetID() string {
	return a.ID
}

// OnCallEscalation represents an escalation of an issue or PR to the on-call user.
type OnCallEscalation struct {
	ID             string    `json:"id"`
	AssignmentID   string    `json:"assignment_id"`
	IssueNumber    int       `json:"issue_number,omitempty"`
	PRNumber       int       `json:"pr_number,omitempty"`
	Repository     string    `json:"repository"`
	Status         string    `json:"status"`
	EscalationTime time.Time `json:"escalation_time,omitempty"`
	ResolutionTime time.Time `json:"resolution_time,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
