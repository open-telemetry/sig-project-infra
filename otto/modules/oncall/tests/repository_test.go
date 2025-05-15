// SPDX-License-Identifier: Apache-2.0

package tests

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

	t.Run("User CRUD operations", func(t *testing.T) {
		// Create user
		user := &oncall.OnCallUser{
			ID:             uuid.New().String(),
			GitHubUsername: "testuser",
			Name:           "Test User",
			Email:          "test@example.com",
			IsActive:       true,
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		// Find by ID
		found, err := repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.GitHubUsername, found.GitHubUsername)

		// Find by GitHub username
		found, err = repo.FindUserByGitHubUsername(ctx, user.GitHubUsername)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)

		// Find all
		users, err := repo.FindAllUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 1)

		// Update
		user.Email = "updated@example.com"
		err = repo.UpdateUser(ctx, user)
		require.NoError(t, err)

		found, err = repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", found.Email)

		// Delete
		err = repo.DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		found, err = repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Rotation CRUD operations", func(t *testing.T) {
		// Create rotation
		rotation := &oncall.OnCallRotation{
			ID:          uuid.New().String(),
			Name:        "Test Rotation",
			Description: "Test rotation description",
			Repository:  "org/repo",
			IsActive:    true,
		}
		err := repo.CreateRotation(ctx, rotation)
		require.NoError(t, err)

		// Find by ID
		found, err := repo.FindRotationByID(ctx, rotation.ID)
		require.NoError(t, err)
		assert.Equal(t, rotation.Name, found.Name)

		// Find by repository
		rotations, err := repo.FindRotationsByRepository(ctx, rotation.Repository)
		require.NoError(t, err)
		assert.Len(t, rotations, 1)

		// Find all
		rotations, err = repo.FindAllRotations(ctx)
		require.NoError(t, err)
		assert.Len(t, rotations, 1)

		// Update
		rotation.Description = "Updated description"
		err = repo.UpdateRotation(ctx, rotation)
		require.NoError(t, err)

		found, err = repo.FindRotationByID(ctx, rotation.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", found.Description)

		// Delete
		err = repo.DeleteRotation(ctx, rotation.ID)
		require.NoError(t, err)

		found, err = repo.FindRotationByID(ctx, rotation.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Assignment CRUD operations", func(t *testing.T) {
		// Create user and rotation first
		user := &oncall.OnCallUser{
			ID:             uuid.New().String(),
			GitHubUsername: "testuser",
			Name:           "Test User",
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		rotation := &oncall.OnCallRotation{
			ID:         uuid.New().String(),
			Name:       "Test Rotation",
			Repository: "org/repo",
		}
		err = repo.CreateRotation(ctx, rotation)
		require.NoError(t, err)

		// Create assignment
		assignment := &oncall.OnCallAssignment{
			ID:         uuid.New().String(),
			RotationID: rotation.ID,
			UserID:     user.ID,
			StartTime:  time.Now(),
			IsCurrent:  true,
		}
		err = repo.CreateAssignment(ctx, assignment)
		require.NoError(t, err)

		// Find by ID
		found, err := repo.FindAssignmentByID(ctx, assignment.ID)
		require.NoError(t, err)
		assert.Equal(t, assignment.RotationID, found.RotationID)
		assert.Equal(t, assignment.UserID, found.UserID)

		// Find current assignment
		current, err := repo.FindCurrentAssignmentByRotation(ctx, rotation.ID)
		require.NoError(t, err)
		assert.Equal(t, assignment.ID, current.ID)

		// Find by user
		userAssignments, err := repo.FindAssignmentsByUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, userAssignments, 1)

		// Find by rotation
		rotationAssignments, err := repo.FindAssignmentsByRotation(ctx, rotation.ID)
		require.NoError(t, err)
		assert.Len(t, rotationAssignments, 1)

		// Create another current assignment to test replacement
		newAssignment := &oncall.OnCallAssignment{
			ID:         uuid.New().String(),
			RotationID: rotation.ID,
			UserID:     user.ID,
			StartTime:  time.Now(),
			IsCurrent:  true,
		}
		err = repo.CreateAssignment(ctx, newAssignment)
		require.NoError(t, err)

		// Check that the first one is no longer current
		found, err = repo.FindAssignmentByID(ctx, assignment.ID)
		require.NoError(t, err)
		assert.False(t, found.IsCurrent)

		// Delete
		err = repo.DeleteAssignment(ctx, newAssignment.ID)
		require.NoError(t, err)

		found, err = repo.FindAssignmentByID(ctx, newAssignment.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Escalation CRUD operations", func(t *testing.T) {
		// Create user, rotation, and assignment first
		user := &oncall.OnCallUser{
			ID:             uuid.New().String(),
			GitHubUsername: "testuser",
			Name:           "Test User",
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		rotation := &oncall.OnCallRotation{
			ID:         uuid.New().String(),
			Name:       "Test Rotation",
			Repository: "org/repo",
		}
		err = repo.CreateRotation(ctx, rotation)
		require.NoError(t, err)

		assignment := &oncall.OnCallAssignment{
			ID:         uuid.New().String(),
			RotationID: rotation.ID,
			UserID:     user.ID,
			StartTime:  time.Now(),
			IsCurrent:  true,
		}
		err = repo.CreateAssignment(ctx, assignment)
		require.NoError(t, err)

		// Create escalation
		escalation := &oncall.OnCallEscalation{
			ID:             uuid.New().String(),
			AssignmentID:   assignment.ID,
			IssueNumber:    123,
			Repository:     "org/repo",
			Status:         oncall.StatusPending,
			EscalationTime: time.Now(),
		}
		err = repo.CreateEscalation(ctx, escalation)
		require.NoError(t, err)

		// Find by ID
		found, err := repo.FindEscalationByID(ctx, escalation.ID)
		require.NoError(t, err)
		assert.Equal(t, escalation.AssignmentID, found.AssignmentID)
		assert.Equal(t, escalation.IssueNumber, found.IssueNumber)

		// Find by status
		statusEscalations, err := repo.FindEscalationsByStatus(ctx, oncall.StatusPending)
		require.NoError(t, err)
		assert.Len(t, statusEscalations, 1)

		// Find by assignment
		assignmentEscalations, err := repo.FindEscalationsByAssignment(ctx, assignment.ID)
		require.NoError(t, err)
		assert.Len(t, assignmentEscalations, 1)

		// Find by issue
		issueEscalation, err := repo.FindEscalationByIssue(ctx, "org/repo", 123)
		require.NoError(t, err)
		assert.Equal(t, escalation.ID, issueEscalation.ID)

		// Update
		escalation.Status = oncall.StatusAcknowledged
		err = repo.UpdateEscalation(ctx, escalation)
		require.NoError(t, err)

		found, err = repo.FindEscalationByID(ctx, escalation.ID)
		require.NoError(t, err)
		assert.Equal(t, oncall.StatusAcknowledged, found.Status)

		// Delete
		err = repo.DeleteEscalation(ctx, escalation.ID)
		require.NoError(t, err)

		found, err = repo.FindEscalationByID(ctx, escalation.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("Transaction operations", func(t *testing.T) {
		err := repo.Transaction(ctx, func(tx oncall.OnCallTransaction) error {
			user := &oncall.OnCallUser{
				ID:             uuid.New().String(),
				GitHubUsername: "txuser",
				Name:           "Transaction User",
			}
			return tx.CreateUser(ctx, user)
		})
		require.NoError(t, err)

		// Verify user was created
		found, err := repo.FindUserByGitHubUsername(ctx, "txuser")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})
}
