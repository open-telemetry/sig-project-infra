// SPDX-License-Identifier: Apache-2.0

package oncall_test

import (
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

	// Test user creation and retrieval
	user := &oncall.OnCallUser{
		ID:             uuid.New().String(),
		GitHubUsername: "testuser",
		Name:           "Test User",
		Email:          "test@example.com",
		IsActive:       true,
	}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Test rotation creation and retrieval
	rotation := &oncall.OnCallRotation{
		ID:          uuid.New().String(),
		Name:        "Test Rotation",
		Description: "Test rotation description",
		Repository:  "org/repo",
		IsActive:    true,
	}
	err = repo.CreateRotation(ctx, rotation)
	require.NoError(t, err)

	// Test assignment creation and retrieval
	assignment := &oncall.OnCallAssignment{
		ID:         uuid.New().String(),
		RotationID: rotation.ID,
		UserID:     user.ID,
		StartTime:  time.Now(),
		IsCurrent:  true,
	}
	err = repo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)

	// Test escalation creation and retrieval
	escalation := &oncall.OnCallEscalation{
		ID:             uuid.New().String(),
		AssignmentID:   assignment.ID,
		IssueNumber:    123,
		Repository:     "org/repo",
		Status:         string(oncall.StatusPending),
		EscalationTime: time.Now(),
	}
	err = repo.CreateEscalation(ctx, escalation)
	require.NoError(t, err)

	// Verify created entities
	foundUser, err := repo.FindUserByGitHubUsername(ctx, "testuser")
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)

	foundRotation, err := repo.FindRotationByID(ctx, rotation.ID)
	require.NoError(t, err)
	assert.Equal(t, rotation.Name, foundRotation.Name)

	foundAssignment, err := repo.FindAssignmentByID(ctx, assignment.ID)
	require.NoError(t, err)
	assert.Equal(t, assignment.UserID, foundAssignment.UserID)

	foundEscalation, err := repo.FindEscalationByIssue(ctx, "org/repo", 123)
	require.NoError(t, err)
	assert.Equal(t, escalation.ID, foundEscalation.ID)
}
