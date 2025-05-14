// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/sig-project-infra/otto/modules/oncall"
)

func TestOnCallModule(t *testing.T) {
	ctx := t.Context()
	logger := NewMockLogger()

	t.Run("Repository operations", func(t *testing.T) {
		repo := oncall.NewMockOnCallRepository()

		// Test user CRUD
		user := &oncall.OnCallUser{
			GitHubUsername: "testuser",
			Name:           "Test User",
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindUserByGitHubUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("Issue comment handling", func(t *testing.T) {
		repo := oncall.NewMockOnCallRepository()
		githubClient := NewMockGitHubProvider()

		config := oncall.Config{
			EnabledRepositories: []string{"org/repo"},
		}

		module := oncall.New(config, oncall.Dependencies{
			Logger: logger,
			GitHub: githubClient,
		})

		// Set repository directly
		module.Repo = repo

		// Create a mock issue comment event with /ack command
		commentEvent := &github.IssueCommentEvent{
			Action: github.Ptr("created"),
			Issue: &github.Issue{
				Number: github.Ptr(123),
			},
			Comment: &github.IssueComment{
				Body: github.Ptr("/ack"),
				User: &github.User{
					Login: github.Ptr("testuser"),
				},
			},
			Repo: &github.Repository{
				FullName: github.Ptr("org/repo"),
			},
		}

		// Create a raw message to pass to HandleEvent
		rawMessage := []byte(`{}`)

		// Handle the event
		err := module.HandleEvent("issue_comment", commentEvent, rawMessage)
		require.NoError(t, err)

		// Check that a user was created
		user, err := repo.FindUserByGitHubUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.NotNil(t, user)

		// Check that an escalation was created
		escalation, err := repo.FindEscalationByIssue(ctx, "org/repo", 123)
		require.NoError(t, err)
		assert.NotNil(t, escalation)
		assert.Equal(t, string(oncall.StatusAcknowledged), escalation.Status)

		// Check that a comment was posted
		key := "org/repo/123"
		require.NotEmpty(t, githubClient.Comments[key])
		assert.Contains(t, githubClient.Comments[key][0], "@testuser has acknowledged this issue")
	})
}
