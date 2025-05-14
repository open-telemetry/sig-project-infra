// SPDX-License-Identifier: Apache-2.0

// Package oncall provides on-call rotation management functionality for handling GitHub issues and pull requests.
package oncall

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
)

// OnCallUserMapper implements SQLMapper for OnCallUser.
type OnCallUserMapper struct{}

// Ensure OnCallUserMapper implements database.SQLMapper.
var _ database.SQLMapper[OnCallUser] = (*OnCallUserMapper)(nil)

// Table returns the database table name.
func (m *OnCallUserMapper) Table() string {
	return "oncall_users"
}

// Columns returns the database column names.
func (m *OnCallUserMapper) Columns() []string {
	return []string{
		"id", "github_username", "name", "email", "is_active", "created_at", "updated_at",
	}
}

// ToRow maps an entity to database column values.
func (m *OnCallUserMapper) ToRow(entity *OnCallUser) []interface{} {
	return []interface{}{
		entity.ID,
		entity.GitHubUsername,
		entity.Name,
		entity.Email,
		entity.IsActive,
		entity.CreatedAt,
		entity.UpdatedAt,
	}
}

// FromRow creates an entity from a database row.
func (m *OnCallUserMapper) FromRow(row *sql.Row) (*OnCallUser, error) {
	var user OnCallUser
	var createdAt, updatedAt string

	err := row.Scan(
		&user.ID,
		&user.GitHubUsername,
		&user.Name,
		&user.Email,
		&user.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %w", err)
	}

	user.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %w", err)
	}

	return &user, nil
}

// FromRows creates entities from database rows.
func (m *OnCallUserMapper) FromRows(rows *sql.Rows) ([]OnCallUser, error) {
	var users []OnCallUser

	for rows.Next() {
		var user OnCallUser
		var createdAt, updatedAt string

		err := rows.Scan(
			&user.ID,
			&user.GitHubUsername,
			&user.Name,
			&user.Email,
			&user.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		user.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %w", err)
		}

		user.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// OnCallRotationMapper implements SQLMapper for OnCallRotation.
type OnCallRotationMapper struct{}

// Ensure OnCallRotationMapper implements database.SQLMapper.
var _ database.SQLMapper[OnCallRotation] = (*OnCallRotationMapper)(nil)

// Table returns the database table name.
func (m *OnCallRotationMapper) Table() string {
	return "oncall_rotations"
}

// Columns returns the database column names.
func (m *OnCallRotationMapper) Columns() []string {
	return []string{
		"id", "name", "description", "repository", "is_active", "created_at", "updated_at",
	}
}

// ToRow maps an entity to database column values.
func (m *OnCallRotationMapper) ToRow(entity *OnCallRotation) []interface{} {
	return []interface{}{
		entity.ID,
		entity.Name,
		entity.Description,
		entity.Repository,
		entity.IsActive,
		entity.CreatedAt,
		entity.UpdatedAt,
	}
}

// FromRow creates an entity from a database row.
func (m *OnCallRotationMapper) FromRow(row *sql.Row) (*OnCallRotation, error) {
	var rotation OnCallRotation
	var createdAt, updatedAt string

	err := row.Scan(
		&rotation.ID,
		&rotation.Name,
		&rotation.Description,
		&rotation.Repository,
		&rotation.IsActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	rotation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %w", err)
	}

	rotation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %w", err)
	}

	return &rotation, nil
}

// FromRows creates entities from database rows.
func (m *OnCallRotationMapper) FromRows(rows *sql.Rows) ([]OnCallRotation, error) {
	var rotations []OnCallRotation

	for rows.Next() {
		var rotation OnCallRotation
		var createdAt, updatedAt string

		err := rows.Scan(
			&rotation.ID,
			&rotation.Name,
			&rotation.Description,
			&rotation.Repository,
			&rotation.IsActive,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		rotation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %w", err)
		}

		rotation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at: %w", err)
		}

		rotations = append(rotations, rotation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rotations, nil
}

// OnCallAssignmentMapper implements SQLMapper for OnCallAssignment.
type OnCallAssignmentMapper struct{}

// Ensure OnCallAssignmentMapper implements database.SQLMapper.
var _ database.SQLMapper[OnCallAssignment] = (*OnCallAssignmentMapper)(nil)

// Table returns the database table name.
func (m *OnCallAssignmentMapper) Table() string {
	return "oncall_assignments"
}

// Columns returns the database column names.
func (m *OnCallAssignmentMapper) Columns() []string {
	return []string{
		"id", "rotation_id", "user_id", "start_time", "end_time", "is_current", "created_at", "updated_at",
	}
}

// ToRow maps an entity to database column values.
func (m *OnCallAssignmentMapper) ToRow(entity *OnCallAssignment) []interface{} {
	return []interface{}{
		entity.ID,
		entity.RotationID,
		entity.UserID,
		entity.StartTime,
		entity.EndTime,
		entity.IsCurrent,
		entity.CreatedAt,
		entity.UpdatedAt,
	}
}

// FromRow creates an entity from a database row.
func (m *OnCallAssignmentMapper) FromRow(row *sql.Row) (*OnCallAssignment, error) {
	var assignment OnCallAssignment
	var startTime, endTime, createdAt, updatedAt string
	var endTimeNull sql.NullString

	err := row.Scan(
		&assignment.ID,
		&assignment.RotationID,
		&assignment.UserID,
		&startTime,
		&endTimeNull,
		&assignment.IsCurrent,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	assignment.StartTime, err = time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("error parsing start_time: %w", err)
	}

	if endTimeNull.Valid {
		endTime = endTimeNull.String
		assignment.EndTime, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			return nil, fmt.Errorf("error parsing end_time: %w", err)
		}
	}

	assignment.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %w", err)
	}

	assignment.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %w", err)
	}

	return &assignment, nil
}

// FromRows creates entities from database rows.
func (m *OnCallAssignmentMapper) FromRows(rows *sql.Rows) ([]OnCallAssignment, error) {
	var assignments []OnCallAssignment

	for rows.Next() {
		var assignment OnCallAssignment
		var startTime, endTime, createdAt, updatedAt string
		var endTimeNull sql.NullString

		err := rows.Scan(
			&assignment.ID,
			&assignment.RotationID,
			&assignment.UserID,
			&startTime,
			&endTimeNull,
			&assignment.IsCurrent,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		assignment.StartTime, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			return nil, fmt.Errorf("error parsing start_time: %w", err)
		}

		if endTimeNull.Valid {
			endTime = endTimeNull.String
			assignment.EndTime, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				return nil, fmt.Errorf("error parsing end_time: %w", err)
			}
		}

		assignment.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %w", err)
		}

		assignment.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

// OnCallEscalationMapper implements SQLMapper for OnCallEscalation.
type OnCallEscalationMapper struct{}

// Ensure OnCallEscalationMapper implements database.SQLMapper.
var _ database.SQLMapper[OnCallEscalation] = (*OnCallEscalationMapper)(nil)

// Table returns the database table name.
func (m *OnCallEscalationMapper) Table() string {
	return "oncall_escalations"
}

// Columns returns the database column names.
func (m *OnCallEscalationMapper) Columns() []string {
	return []string{
		"id", "assignment_id", "issue_number", "pr_number", "repository", "status",
		"escalation_time", "resolution_time", "created_at", "updated_at",
	}
}

// ToRow maps an entity to database column values.
func (m *OnCallEscalationMapper) ToRow(entity *OnCallEscalation) []interface{} {
	var issueNumberSQL, prNumberSQL sql.NullInt32
	if entity.IssueNumber != 0 {
		// Check for potential integer overflow before conversion
		if entity.IssueNumber > 0 && entity.IssueNumber <= 2147483647 {
			issueNumberSQL = sql.NullInt32{Int32: int32(entity.IssueNumber), Valid: true}
		} else {
			// If issue number is too large, use a safer default
			issueNumberSQL = sql.NullInt32{Int32: 0, Valid: true}
		}
	}
	if entity.PRNumber != 0 {
		// Check for potential integer overflow before conversion
		if entity.PRNumber > 0 && entity.PRNumber <= 2147483647 {
			prNumberSQL = sql.NullInt32{Int32: int32(entity.PRNumber), Valid: true}
		} else {
			// If PR number is too large, use a safer default
			prNumberSQL = sql.NullInt32{Int32: 0, Valid: true}
		}
	}

	var escalationTimeSQL, resolutionTimeSQL sql.NullTime
	if !entity.EscalationTime.IsZero() {
		escalationTimeSQL = sql.NullTime{Time: entity.EscalationTime, Valid: true}
	}
	if !entity.ResolutionTime.IsZero() {
		resolutionTimeSQL = sql.NullTime{Time: entity.ResolutionTime, Valid: true}
	}

	return []interface{}{
		entity.ID,
		entity.AssignmentID,
		issueNumberSQL,
		prNumberSQL,
		entity.Repository,
		entity.Status,
		escalationTimeSQL,
		resolutionTimeSQL,
		entity.CreatedAt,
		entity.UpdatedAt,
	}
}

// FromRow creates an entity from a database row.
func (m *OnCallEscalationMapper) FromRow(row *sql.Row) (*OnCallEscalation, error) {
	var escalation OnCallEscalation
	var createdAt, updatedAt string
	var issueNumber, prNumber sql.NullInt32
	var escalationTime, resolutionTime sql.NullString

	err := row.Scan(
		&escalation.ID,
		&escalation.AssignmentID,
		&issueNumber,
		&prNumber,
		&escalation.Repository,
		&escalation.Status,
		&escalationTime,
		&resolutionTime,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if issueNumber.Valid {
		escalation.IssueNumber = int(issueNumber.Int32)
	}

	if prNumber.Valid {
		escalation.PRNumber = int(prNumber.Int32)
	}

	if escalationTime.Valid {
		escalation.EscalationTime, err = time.Parse(time.RFC3339, escalationTime.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing escalation_time: %w", err)
		}
	}

	if resolutionTime.Valid {
		escalation.ResolutionTime, err = time.Parse(time.RFC3339, resolutionTime.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing resolution_time: %w", err)
		}
	}

	escalation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %w", err)
	}

	escalation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %w", err)
	}

	return &escalation, nil
}

// FromRows creates entities from database rows.
func (m *OnCallEscalationMapper) FromRows(rows *sql.Rows) ([]OnCallEscalation, error) {
	var escalations []OnCallEscalation

	for rows.Next() {
		var escalation OnCallEscalation
		var createdAt, updatedAt string
		var issueNumber, prNumber sql.NullInt32
		var escalationTime, resolutionTime sql.NullString

		err := rows.Scan(
			&escalation.ID,
			&escalation.AssignmentID,
			&issueNumber,
			&prNumber,
			&escalation.Repository,
			&escalation.Status,
			&escalationTime,
			&resolutionTime,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if issueNumber.Valid {
			escalation.IssueNumber = int(issueNumber.Int32)
		}

		if prNumber.Valid {
			escalation.PRNumber = int(prNumber.Int32)
		}

		if escalationTime.Valid {
			escalation.EscalationTime, err = time.Parse(time.RFC3339, escalationTime.String)
			if err != nil {
				return nil, fmt.Errorf("error parsing escalation_time: %w", err)
			}
		}

		if resolutionTime.Valid {
			escalation.ResolutionTime, err = time.Parse(time.RFC3339, resolutionTime.String)
			if err != nil {
				return nil, fmt.Errorf("error parsing resolution_time: %w", err)
			}
		}

		escalation.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %w", err)
		}

		escalation.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at: %w", err)
		}

		escalations = append(escalations, escalation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return escalations, nil
}
