// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"database/sql"
	"testing"
)

func openTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := AutoMigrateOnCall(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestAddTaskAndGetTaskByIssueNumber(t *testing.T) {
	db := openTestDB(t)
	sch, _ := AddSchedule(db, "primary", "round-robin")
	user, _ := AddUser(db, "testuser", "Test User")
	_ = AssignUserToSchedule(db, sch.ID, user.ID, 0)
	task, err := AddTask(db, sch.ID, "org/repo", 42, "Issue #42", "desc", user.ID)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	got, err := GetTaskByIssueNumber(db, "org/repo", 42)
	if err != nil || got == nil {
		t.Fatalf("GetTaskByIssueNumber failed: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("wrong task ID: want %d, got %d", task.ID, got.ID)
	}
}

func TestTaskAcknowledge(t *testing.T) {
	db := openTestDB(t)
	sch, _ := AddSchedule(db, "primary", "round-robin")
	user, _ := AddUser(db, "a", "A")
	_ = AssignUserToSchedule(db, sch.ID, user.ID, 0)
	task, err := AddTask(db, sch.ID, "repo", 1, "t", "desc", user.ID)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if err := UpdateTaskStatus(db, task.ID, "ack"); err != nil {
		t.Errorf("acknowledge failed: %v", err)
	}
	updated, err := GetTask(db, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if updated == nil {
		t.Fatalf("updated task not found in db")
	}
	if updated.Status != "ack" {
		t.Errorf("expected status 'ack', got %q", updated.Status)
	}
}
