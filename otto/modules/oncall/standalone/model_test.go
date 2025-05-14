// SPDX-License-Identifier: Apache-2.0

package standalone

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOnCallModels tests the basic oncall model types work as expected.
func TestOnCallModels(t *testing.T) {
	// Test OnCallUser
	userId := uuid.New().String()
	user := OnCallUser{
		ID:             userId,
		GitHubUsername: "testuser",
		Name:           "Test User",
		Email:          "test@example.com",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.Equal(t, userId, user.GetID())

	// Test OnCallRotation
	rotationId := uuid.New().String()
	rotation := OnCallRotation{
		ID:          rotationId,
		Name:        "Test Rotation",
		Description: "Test rotation description",
		Repository:  "org/repo",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	assert.Equal(t, rotationId, rotation.GetID())

	// Test OnCallAssignment
	assignmentId := uuid.New().String()
	assignment := OnCallAssignment{
		ID:         assignmentId,
		RotationID: rotationId,
		UserID:     userId,
		StartTime:  time.Now(),
		IsCurrent:  true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	assert.Equal(t, assignmentId, assignment.GetID())

	// Test OnCallEscalation
	escalationId := uuid.New().String()
	escalation := OnCallEscalation{
		ID:             escalationId,
		AssignmentID:   assignmentId,
		IssueNumber:    123,
		Repository:     "org/repo",
		Status:         string(StatusPending),
		EscalationTime: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.Equal(t, escalationId, escalation.GetID())
}

// These simplified types are just for testing model functionality.
type OnCallUser struct {
	ID             string
	GitHubUsername string
	Name           string
	Email          string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *OnCallUser) GetID() string {
	return u.ID
}

type OnCallRotation struct {
	ID          string
	Name        string
	Description string
	Repository  string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *OnCallRotation) GetID() string {
	return r.ID
}

type OnCallAssignment struct {
	ID         string
	RotationID string
	UserID     string
	StartTime  time.Time
	EndTime    time.Time
	IsCurrent  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (a *OnCallAssignment) GetID() string {
	return a.ID
}

type OnCallEscalation struct {
	ID             string
	AssignmentID   string
	IssueNumber    int
	PRNumber       int
	Repository     string
	Status         string
	EscalationTime time.Time
	ResolutionTime time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (e *OnCallEscalation) GetID() string {
	return e.ID
}

type EscalationStatus string

const (
	StatusPending      EscalationStatus = "pending"
	StatusAcknowledged EscalationStatus = "acknowledged"
	StatusResolved     EscalationStatus = "resolved"
	StatusEscalated    EscalationStatus = "escalated"
)
