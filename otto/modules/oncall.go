// SPDX-License-Identifier: Apache-2.0

// Package modules contains implementations of the various Otto modules.
package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal"
)

// Import internal types for error handling.
var (
	LogAndWrapError  = internal.LogAndWrapError
	ErrorTypeCommand = internal.ErrorTypeCommand
)

// OnCallModule handles on-call rotation management.
type OnCallModule struct {
	app      *internal.App
	database *internal.Database
}

func (o *OnCallModule) Name() string { return "oncall" }

// Initialize implements the ModuleInitializer interface.
func (o *OnCallModule) Initialize(ctx context.Context, app *internal.App) error {
	o.app = app
	o.database = app.Database

	// Initialize database tables
	if err := AutoMigrateOnCall(o.database.DB()); err != nil {
		return err
	}

	// Start a ticker to check unacknowledged tasks every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := o.CheckUnacknowledgedTasks(); err != nil {
					slog.Error("Error checking unacknowledged tasks", "error", err)
				}
			}
		}
	}()

	return nil
}

func (o *OnCallModule) AcknowledgeTask(repo string, issueNum int, user string) error {
	// Find the task
	task, err := GetTaskByIssueNumber(o.database.DB(), repo, issueNum)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("no task found for repo %s, issue %d", repo, issueNum)
	}

	// Update task status
	err = UpdateTaskStatus(o.database.DB(), task.ID, "ack")
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (o *OnCallModule) CheckUnacknowledgedTasks() error {
	// Query for unacknowledged tasks older than 24 hours
	rows, err := o.database.DB().Query(`
		SELECT id, repo, issue_num, assigned_to, created_at
		FROM oncall_tasks
		WHERE status != 'ack'
		AND created_at < datetime('now', '-24 hours')
	`)
	if err != nil {
		return fmt.Errorf("failed to query unacknowledged tasks: %w", err)
	}
	defer rows.Close()

	// Process each unacknowledged task
	for rows.Next() {
		var taskID int64
		var repo string
		var issueNum int
		var assignedToID int64
		var createdAt time.Time

		if err := rows.Scan(&taskID, &repo, &issueNum, &assignedToID, &createdAt); err != nil {
			slog.Error("Failed to scan task row", "error", err)
			continue
		}

		// Notify about escalation
		err = o.EscalateTask(taskID, repo, issueNum)
		if err != nil {
			slog.Error("Task escalation failed",
				"task_id", taskID,
				"repo", repo,
				"issue_num", issueNum,
				"error", err)
		}
	}

	return nil
}

func (o *OnCallModule) EscalateTask(taskID int64, repo string, issueNum int) error {
	// Get the task details
	task, err := GetTask(o.database.DB(), taskID)
	if err != nil {
		return fmt.Errorf("failed to get task details: %w", err)
	}

	// Determine escalation group (could be a configuration)
	escalationGroup := []string{"@org/oncall-team", "@org/leadership"}

	// Post escalation comment
	err = o.PostGitHubComment(repo, issueNum,
		fmt.Sprintf("⚠️ ESCALATION: Task has been unacknowledged for over 24 hours.\n"+
			"Assigned to: %d\n"+
			"Escalation Group: %s",
			task.AssignedTo,
			strings.Join(escalationGroup, ", ")))

	return err
}

func (o *OnCallModule) PostGitHubComment(repo string, issueNum int, message string) error {
	// Check if we have GitHub client available
	if o.app == nil || o.app.GitHubClient == nil {
		// Log the action without posting to GitHub
		slog.Info("GitHub comment would be posted (no GitHub client available)",
			"repo", repo,
			"issue_num", issueNum,
			"message", message)
		return nil
	}

	// Parse repo into owner and repo name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s, expected owner/repo", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Create the comment
	comment := &github.IssueComment{
		Body: github.Ptr(message),
	}

	// Create context
	ctx := context.Background()

	// Post the comment using the app's GitHub client
	_, _, err := o.app.GitHubClient.Issues.CreateComment(ctx, owner, repoName, issueNum, comment)
	if err != nil {
		return fmt.Errorf("failed to post GitHub comment: %w", err)
	}

	slog.Info("GitHub comment posted successfully",
		"repo", repo,
		"issue_num", issueNum)
	return nil
}

// Shutdown implements the ModuleShutdowner interface.
func (o *OnCallModule) Shutdown(ctx context.Context) error {
	// Nothing to clean up
	return nil
}

func (o *OnCallModule) HandleEvent(eventType string, event any, raw json.RawMessage) error {
	db := o.database.DB()
	if db == nil {
		return internal.LogAndWrapError(
			nil,
			internal.ErrorTypeCommand,
			"no_db_connection",
			map[string]any{
				"module": "oncall",
			},
		)
	}

	switch eventType {
	case "issues":
		// Cast to GitHub issues event
		issuesEvent, ok := event.(*github.IssuesEvent)
		if !ok {
			return internal.LogAndWrapError(
				nil,
				internal.ErrorTypeCommand,
				"invalid_event_type",
				map[string]any{
					"event_type": eventType,
				},
			)
		}

		// Check if the issue is closed
		if issuesEvent.GetAction() == "closed" {
			// Find the task associated with this issue
			repo := issuesEvent.GetRepo().GetFullName()
			issueNum := issuesEvent.GetIssue().GetNumber()

			task, err := GetTaskByIssueNumber(db, repo, issueNum)
			if err != nil {
				return LogAndWrapError(err, ErrorTypeCommand, "get_task", map[string]any{
					"repo":  repo,
					"issue": issueNum,
				})
			}

			// If task exists and is not already done, mark it as done
			if task != nil && task.Status != "done" {
				if err := UpdateTaskStatus(db, task.ID, "done"); err != nil {
					return LogAndWrapError(
						err,
						ErrorTypeCommand,
						"update_task_status",
						map[string]any{
							"task_id": task.ID,
							"status":  "done",
						},
					)
				}
				slog.Info("Task marked as done due to issue closure",
					"task_id", task.ID,
					"repo", repo,
					"issue_num", issueNum)
			}
		}
	case "comment":
		commentEvent, ok := event.(*github.IssueCommentEvent)
		if !ok {
			return LogAndWrapError(nil, ErrorTypeCommand, "invalid_event_type", map[string]any{
				"event_type": "comment",
			})
		}
		task, err := GetTaskByIssueNumber(db, *commentEvent.Repo.Name, *commentEvent.Issue.Number)
		if err != nil {
			return LogAndWrapError(
				err,
				ErrorTypeCommand,
				"get_task_by_issue_number",
				map[string]any{
					"repo":      *commentEvent.Repo.Name,
					"issue_num": *commentEvent.Issue.Number,
				},
			)
		}
		if strings.Contains(*commentEvent.GetComment().Body, "/ack") {
			currentOnCall, err := GetCurrentOnCallUser(db, "primary")
			if err != nil {
				return LogAndWrapError(
					err,
					ErrorTypeCommand,
					"get_current_oncall_user",
					map[string]any{
						"schedule_name": "primary",
					},
				)
			}
			if currentOnCall.GitHub == *commentEvent.GetComment().User.Login {
				if err := UpdateTaskStatus(db, task.ID, "ack"); err != nil {
					return LogAndWrapError(
						err,
						ErrorTypeCommand,
						"update_task_status",
						map[string]any{
							"task_id": task.ID,
							"status":  "ack",
						},
					)
				}
				slog.Info("Task marked as acknowledged.",
					"task_id", task.ID,
					"repo", task.Repo,
					"issue_num", task.IssueNum,
					"acknowledged_by", currentOnCall.GitHub)
			}
		}
	}
	return nil
}
