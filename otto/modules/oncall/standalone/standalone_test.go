// SPDX-License-Identifier: Apache-2.0

package standalone

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/sig-project-infra/otto/modules/oncall"
)

func TestMockRepository(t *testing.T) {
	ctx := t.Context()
	repo := oncall.NewMockOnCallRepository()

	// Test entity creation
	testUser := createTestUser(t, ctx, repo)
	testRotation := createTestRotation(t, ctx, repo)
	testAssignment := createTestAssignment(t, ctx, repo, testRotation.ID, testUser.ID)
	testEscalation := createTestEscalation(t, ctx, repo, testAssignment.ID)

	// Test entity retrieval
	foundUser, err := repo.FindUserByID(ctx, testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, testUser.GitHubUsername, foundUser.GitHubUsername)

	foundRotation, err := repo.FindRotationByID(ctx, testRotation.ID)
	require.NoError(t, err)
	assert.Equal(t, testRotation.Name, foundRotation.Name)

	foundAssignment, err := repo.FindAssignmentByID(ctx, testAssignment.ID)
	require.NoError(t, err)
	assert.Equal(t, testAssignment.UserID, foundAssignment.UserID)

	foundEscalation, err := repo.FindEscalationByID(ctx, testEscalation.ID)
	require.NoError(t, err)
	assert.Equal(t, testEscalation.Status, foundEscalation.Status)
}

func createTestUser(t *testing.T, ctx context.Context, repo oncall.OnCallRepository) *oncall.OnCallUser {
	user := &oncall.OnCallUser{
		ID:             uuid.New().String(),
		GitHubUsername: "testuser",
		Name:           "Test User",
		Email:          "test@example.com",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	return user
}

func createTestRotation(t *testing.T, ctx context.Context, repo oncall.OnCallRepository) *oncall.OnCallRotation {
	rotation := &oncall.OnCallRotation{
		ID:          uuid.New().String(),
		Name:        "Test Rotation",
		Description: "Test rotation description",
		Repository:  "org/repo",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := repo.CreateRotation(ctx, rotation)
	require.NoError(t, err)
	return rotation
}

func createTestAssignment(
	t *testing.T,
	ctx context.Context,
	repo oncall.OnCallRepository,
	rotationID, userID string,
) *oncall.OnCallAssignment {
	assignment := &oncall.OnCallAssignment{
		ID:         uuid.New().String(),
		RotationID: rotationID,
		UserID:     userID,
		StartTime:  time.Now(),
		IsCurrent:  true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := repo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)
	return assignment
}

func createTestEscalation(
	t *testing.T,
	ctx context.Context,
	repo oncall.OnCallRepository,
	assignmentID string,
) *oncall.OnCallEscalation {
	escalation := &oncall.OnCallEscalation{
		ID:             uuid.New().String(),
		AssignmentID:   assignmentID,
		IssueNumber:    123,
		Repository:     "org/repo",
		Status:         string(oncall.StatusPending),
		EscalationTime: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err := repo.CreateEscalation(ctx, escalation)
	require.NoError(t, err)
	return escalation
}
