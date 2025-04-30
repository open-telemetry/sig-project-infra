// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Migration, AddUser, AddSchedule, AssignUserToSchedule, etc.

func AutoMigrateOnCall(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS oncall_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			github TEXT UNIQUE NOT NULL,
			display_name TEXT,
			active BOOLEAN NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS oncall_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			policy TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT 1,
			current_rotation_idx INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS oncall_schedules_users (
			schedule_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			PRIMARY KEY (schedule_id, user_id),
			FOREIGN KEY(schedule_id) REFERENCES oncall_schedules(id),
			FOREIGN KEY(user_id) REFERENCES oncall_users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS oncall_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			schedule_id INTEGER NOT NULL,
			repo TEXT,
			issue_num INTEGER,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'open',
			assigned_to INTEGER,
			created_at TIMESTAMP NOT NULL,
			acked_at TIMESTAMP,
			completed_at TIMESTAMP,
			FOREIGN KEY(schedule_id) REFERENCES oncall_schedules(id),
			FOREIGN KEY(assigned_to) REFERENCES oncall_users(id)
		);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("failed migration: %w (SQL: %s)", err, s)
		}
	}
	return nil
}

func AddUser(db *sql.DB, gh, name string) (*OnCallUser, error) {
	now := time.Now()
	res, err := db.Exec(
		`INSERT INTO oncall_users (github, display_name, active, created_at) VALUES (?, ?, 1, ?)`,
		gh,
		name,
		now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &OnCallUser{ID: id, GitHub: gh, DisplayName: name, Active: true, CreatedAt: now}, nil
}

func AddSchedule(db *sql.DB, name, policyStr string) (*OnCallSchedule, error) {
	now := time.Now()

	// Convert string to OnCallScheduleRotationPolicy
	var policy OnCallScheduleRotationPolicy
	switch policyStr {
	case "round-robin":
		policy = RoundRobinPolicy
	case "sequential":
		policy = SequentialPolicy
	case "random":
		policy = RandomPolicy
	default:
		policy = RoundRobinPolicy // Default to round-robin if unrecognized
	}

	res, err := db.Exec(
		`INSERT INTO oncall_schedules (name, policy, enabled, current_rotation_idx, created_at, updated_at) VALUES (?, ?, 1, 0, ?, ?)`,
		name,
		string(policy),
		now,
		now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &OnCallSchedule{
		ID:                 id,
		Name:               name,
		Policy:             policy,
		Enabled:            true,
		CurrentRotationIdx: 0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

func AssignUserToSchedule(db *sql.DB, scheduleID, userID int64, position int) error {
	_, err := db.Exec(
		`INSERT INTO oncall_schedules_users (schedule_id, user_id, position) VALUES (?, ?, ?)`,
		scheduleID, userID, position,
	)
	return err
}

func GetScheduleByName(db *sql.DB, name string) (*OnCallSchedule, error) {
	row := db.QueryRow(
		`SELECT id, name, policy, enabled, current_rotation_idx, created_at, updated_at FROM oncall_schedules WHERE name = ?`,
		name,
	)
	var s OnCallSchedule
	err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Policy,
		&s.Enabled,
		&s.CurrentRotationIdx,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func GetCurrentOnCallUser(db *sql.DB, scheduleName string) (*OnCallUser, error) {
	// Get the schedule
	schedule, err := GetScheduleByName(db, scheduleName)
	if err != nil || schedule == nil {
		return nil, fmt.Errorf("schedule not found: %s", scheduleName)
	}

	// Get users in the schedule
	users, err := ListUsersForSchedule(db, schedule.ID)
	if err != nil || len(users) == 0 {
		return nil, fmt.Errorf("no users found in schedule: %s", scheduleName)
	}

	// For round-robin, use current rotation index
	var currentUser OnCallUser
	switch schedule.Policy {
	case RoundRobinPolicy:
		idx := schedule.CurrentRotationIdx % len(users)
		currentUserSchedule := users[idx]
		row := db.QueryRow(
			`SELECT id, github, display_name, active, created_at FROM oncall_users WHERE id = ?`,
			currentUserSchedule.UserID,
		)
		err = row.Scan(
			&currentUser.ID,
			&currentUser.GitHub,
			&currentUser.DisplayName,
			&currentUser.Active,
			&currentUser.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported schedule policy: %s", schedule.Policy)
	}

	return &currentUser, nil
}

func AdvanceOnCallSchedule(db *sql.DB, scheduleName string) error {
	// Get the schedule
	schedule, err := GetScheduleByName(db, scheduleName)
	if err != nil || schedule == nil {
		return fmt.Errorf("schedule not found: %s", scheduleName)
	}

	// Get users in the schedule
	users, err := ListUsersForSchedule(db, schedule.ID)
	if err != nil || len(users) == 0 {
		return fmt.Errorf("no users found in schedule: %s", scheduleName)
	}

	// Increment rotation index
	newRotationIdx := (schedule.CurrentRotationIdx + 1) % len(users)

	// Update the schedule's current rotation index
	_, err = db.Exec(
		`UPDATE oncall_schedules SET current_rotation_idx = ?, updated_at = ? WHERE id = ?`,
		newRotationIdx,
		time.Now(),
		schedule.ID,
	)

	return err
}

func ListUsersForSchedule(db *sql.DB, scheduleID int64) ([]OnCallScheduleUser, error) {
	rows, err := db.Query(
		`SELECT schedule_id, user_id, position FROM oncall_schedules_users WHERE schedule_id = ? ORDER BY position ASC`,
		scheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rels []OnCallScheduleUser
	for rows.Next() {
		var rel OnCallScheduleUser
		if err := rows.Scan(&rel.ScheduleID, &rel.UserID, &rel.Position); err != nil {
			return nil, err
		}
		rels = append(rels, rel)
	}
	return rels, nil
}

func AddTask(
	db *sql.DB,
	scheduleID int64,
	repo string,
	issueNum int,
	title, description string,
	assignedTo int64,
) (*OnCallTask, error) {
	now := time.Now()
	res, err := db.Exec(
		`INSERT INTO oncall_tasks (schedule_id, repo, issue_num, title, description, status, assigned_to, created_at) VALUES (?, ?, ?, ?, ?, 'open', ?, ?)`,
		scheduleID,
		repo,
		issueNum,
		title,
		description,
		assignedTo,
		now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &OnCallTask{
		ID:          id,
		ScheduleID:  scheduleID,
		Repo:        repo,
		IssueNum:    issueNum,
		Title:       title,
		Description: description,
		Status:      "open",
		AssignedTo:  assignedTo,
		CreatedAt:   now,
	}, nil
}

func GetTaskByIssueNumber(db *sql.DB, repo string, issueNum int) (*OnCallTask, error) {
	row := db.QueryRow(
		`SELECT id, schedule_id, repo, issue_num, title, description, status, assigned_to, created_at, acked_at, completed_at
		 FROM oncall_tasks WHERE repo = ? AND issue_num = ?`,
		repo,
		issueNum,
	)
	var t OnCallTask
	err := row.Scan(
		&t.ID,
		&t.ScheduleID,
		&t.Repo,
		&t.IssueNum,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.AssignedTo,
		&t.CreatedAt,
		&t.AckedAt,
		&t.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func UpdateTaskStatus(db *sql.DB, id int64, status string) error {
	now := time.Now()
	var tsField string
	switch status {
	case "ack":
		tsField = "acked_at"
	case "done":
		tsField = "completed_at"
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// Start a transaction to ensure the update and verify it
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Failed to rollback transaction", "error", err)
		}
	}() // Rollback in case of error, won't do anything if commit succeeds

	// Execute the update
	result, err := tx.Exec(
		fmt.Sprintf(`UPDATE oncall_tasks SET status = ?, %s = ? WHERE id = ?`, tsField),
		status, now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Check that a row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no task found with id %d", id)
	}

	// Verify the update
	var currentStatus string
	err = tx.QueryRow(`SELECT status FROM oncall_tasks WHERE id = ?`, id).Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("failed to verify task status: %w", err)
	}
	if currentStatus != status {
		return fmt.Errorf(
			"failed to update task status: current status is %s, expected %s",
			currentStatus,
			status,
		)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetTask(db *sql.DB, id int64) (*OnCallTask, error) {
	row := db.QueryRow(
		`SELECT id, schedule_id, repo, issue_num, title, description, status, assigned_to, created_at, acked_at, completed_at FROM oncall_tasks WHERE id = ?`,
		id,
	)
	var t OnCallTask
	err := row.Scan(
		&t.ID,
		&t.ScheduleID,
		&t.Repo,
		&t.IssueNum,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.AssignedTo,
		&t.CreatedAt,
		&t.AckedAt,
		&t.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}
