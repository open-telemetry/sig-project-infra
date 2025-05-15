// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
)

// OnCallRepository defines the interface for oncall data operations.
type OnCallRepository interface {
	// User operations
	FindUserByID(ctx context.Context, id string) (*OnCallUser, error)
	FindUserByGitHubUsername(ctx context.Context, username string) (*OnCallUser, error)
	FindAllUsers(ctx context.Context) ([]OnCallUser, error)
	CreateUser(ctx context.Context, user *OnCallUser) error
	UpdateUser(ctx context.Context, user *OnCallUser) error
	DeleteUser(ctx context.Context, id string) error

	// Rotation operations
	FindRotationByID(ctx context.Context, id string) (*OnCallRotation, error)
	FindRotationsByRepository(ctx context.Context, repository string) ([]OnCallRotation, error)
	FindAllRotations(ctx context.Context) ([]OnCallRotation, error)
	CreateRotation(ctx context.Context, rotation *OnCallRotation) error
	UpdateRotation(ctx context.Context, rotation *OnCallRotation) error
	DeleteRotation(ctx context.Context, id string) error

	// Assignment operations
	FindAssignmentByID(ctx context.Context, id string) (*OnCallAssignment, error)
	FindCurrentAssignmentByRotation(ctx context.Context, rotationID string) (*OnCallAssignment, error)
	FindAssignmentsByUser(ctx context.Context, userID string) ([]OnCallAssignment, error)
	FindAssignmentsByRotation(ctx context.Context, rotationID string) ([]OnCallAssignment, error)
	CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error
	UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error
	DeleteAssignment(ctx context.Context, id string) error

	// Escalation operations
	FindEscalationByID(ctx context.Context, id string) (*OnCallEscalation, error)
	FindEscalationsByStatus(ctx context.Context, status EscalationStatus) ([]OnCallEscalation, error)
	FindEscalationsByAssignment(ctx context.Context, assignmentID string) ([]OnCallEscalation, error)
	FindEscalationByIssue(ctx context.Context, repository string, issueNumber int) (*OnCallEscalation, error)
	FindEscalationByPR(ctx context.Context, repository string, prNumber int) (*OnCallEscalation, error)
	CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error
	UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error
	DeleteEscalation(ctx context.Context, id string) error

	// Transaction support
	Transaction(ctx context.Context, fn func(tx OnCallTransaction) error) error

	// Schema management
	EnsureSchema(ctx context.Context) error
}

// OnCallTransaction defines transactional operations for oncall data.
type OnCallTransaction interface {
	// User operations
	CreateUser(ctx context.Context, user *OnCallUser) error
	UpdateUser(ctx context.Context, user *OnCallUser) error
	DeleteUser(ctx context.Context, id string) error

	// Rotation operations
	CreateRotation(ctx context.Context, rotation *OnCallRotation) error
	UpdateRotation(ctx context.Context, rotation *OnCallRotation) error
	DeleteRotation(ctx context.Context, id string) error

	// Assignment operations
	CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error
	UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error
	DeleteAssignment(ctx context.Context, id string) error

	// Escalation operations
	CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error
	UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error
	DeleteEscalation(ctx context.Context, id string) error
}

// SQLiteOnCallRepository implements OnCallRepository for SQLite databases using sqlx.
type SQLiteOnCallRepository struct {
	db database.Provider
}

// Ensure SQLiteOnCallRepository implements OnCallRepository.
var _ OnCallRepository = (*SQLiteOnCallRepository)(nil)

// validateTableName ensures table names are safe for use in SQL queries.
func validateTableName(tableName string) error {
	// Restrict table names to alphanumeric characters, underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validPattern.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}
	return nil
}

// validateTableNames validates all table names used by the repository.
func validateTableNames() error {
	tables := []string{
		"oncall_users",
		"oncall_rotations",
		"oncall_assignments",
		"oncall_escalations",
	}

	for _, tableName := range tables {
		if err := validateTableName(tableName); err != nil {
			return err
		}
	}

	return nil
}

// NewSQLiteOnCallRepository creates a new SQLite-backed OnCallRepository using sqlx.
func NewSQLiteOnCallRepository(db database.Provider) (OnCallRepository, error) {
	// Validate all table names at initialization time
	if err := validateTableNames(); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &SQLiteOnCallRepository{
		db: db,
	}, nil
}

// FindUserByID retrieves a user by their ID.
func (r *SQLiteOnCallRepository) FindUserByID(ctx context.Context, id string) (*OnCallUser, error) {
	var user OnCallUser
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_users WHERE id = ?"
	err := r.db.Sqlx().GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding user with id %s: %w", id, err)
	}
	return &user, nil
}

// FindUserByGitHubUsername retrieves a user by their GitHub username.
func (r *SQLiteOnCallRepository) FindUserByGitHubUsername(ctx context.Context, username string) (*OnCallUser, error) {
	var user OnCallUser
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_users WHERE github_username = ?"
	err := r.db.Sqlx().GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with GitHub username %s: %w", username, err)
		}
		return nil, fmt.Errorf("error finding user with GitHub username %s: %w", username, err)
	}
	return &user, nil
}

// FindAllUsers retrieves all users.
func (r *SQLiteOnCallRepository) FindAllUsers(ctx context.Context) ([]OnCallUser, error) {
	var users []OnCallUser
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_users"
	err := r.db.Sqlx().SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all users: %w", err)
	}
	return users, nil
}

// CreateUser inserts a new user.
func (r *SQLiteOnCallRepository) CreateUser(ctx context.Context, user *OnCallUser) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_users 
		(id, github_username, name, email, is_active, created_at, updated_at) 
		VALUES (:id, :github_username, :name, :email, :is_active, :created_at, :updated_at)
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

// UpdateUser updates an existing user.
func (r *SQLiteOnCallRepository) UpdateUser(ctx context.Context, user *OnCallUser) error {
	user.UpdatedAt = time.Now()

	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_users 
		SET github_username = :github_username, 
			name = :name, 
			email = :email, 
			is_active = :is_active, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

// DeleteUser removes a user by their ID.
func (r *SQLiteOnCallRepository) DeleteUser(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_users WHERE id = ?"
	_, err := r.db.Sqlx().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user with id %s: %w", id, err)
	}
	return nil
}

// FindRotationByID retrieves a rotation by its ID.
func (r *SQLiteOnCallRepository) FindRotationByID(ctx context.Context, id string) (*OnCallRotation, error) {
	var rotation OnCallRotation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_rotations WHERE id = ?"
	err := r.db.Sqlx().GetContext(ctx, &rotation, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rotation not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding rotation with id %s: %w", id, err)
	}
	return &rotation, nil
}

// FindRotationsByRepository retrieves rotations by repository.
func (r *SQLiteOnCallRepository) FindRotationsByRepository(ctx context.Context, repository string) ([]OnCallRotation, error) {
	var rotations []OnCallRotation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_rotations WHERE repository = ?"
	err := r.db.Sqlx().SelectContext(ctx, &rotations, query, repository)
	if err != nil {
		return nil, fmt.Errorf("error querying rotations by repository: %w", err)
	}
	return rotations, nil
}

// FindAllRotations retrieves all rotations.
func (r *SQLiteOnCallRepository) FindAllRotations(ctx context.Context) ([]OnCallRotation, error) {
	var rotations []OnCallRotation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_rotations"
	err := r.db.Sqlx().SelectContext(ctx, &rotations, query)
	if err != nil {
		return nil, fmt.Errorf("error querying all rotations: %w", err)
	}
	return rotations, nil
}

// CreateRotation inserts a new rotation.
func (r *SQLiteOnCallRepository) CreateRotation(ctx context.Context, rotation *OnCallRotation) error {
	if rotation.ID == "" {
		rotation.ID = uuid.New().String()
	}
	if rotation.CreatedAt.IsZero() {
		rotation.CreatedAt = time.Now()
	}
	if rotation.UpdatedAt.IsZero() {
		rotation.UpdatedAt = time.Now()
	}

	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_rotations 
		(id, name, description, repository, is_active, created_at, updated_at) 
		VALUES (:id, :name, :description, :repository, :is_active, :created_at, :updated_at)
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, rotation)
	if err != nil {
		return fmt.Errorf("error creating rotation: %w", err)
	}
	return nil
}

// UpdateRotation updates an existing rotation.
func (r *SQLiteOnCallRepository) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	rotation.UpdatedAt = time.Now()

	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_rotations 
		SET name = :name, 
			description = :description, 
			repository = :repository, 
			is_active = :is_active, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, rotation)
	if err != nil {
		return fmt.Errorf("error updating rotation: %w", err)
	}
	return nil
}

// DeleteRotation removes a rotation by its ID.
func (r *SQLiteOnCallRepository) DeleteRotation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_rotations WHERE id = ?"
	_, err := r.db.Sqlx().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting rotation with id %s: %w", id, err)
	}
	return nil
}

// FindAssignmentByID retrieves an assignment by its ID.
func (r *SQLiteOnCallRepository) FindAssignmentByID(ctx context.Context, id string) (*OnCallAssignment, error) {
	var assignment OnCallAssignment
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_assignments WHERE id = ?"
	err := r.db.Sqlx().GetContext(ctx, &assignment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("assignment not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding assignment with id %s: %w", id, err)
	}
	return &assignment, nil
}

// FindCurrentAssignmentByRotation retrieves the current assignment for a rotation.
func (r *SQLiteOnCallRepository) FindCurrentAssignmentByRotation(ctx context.Context, rotationID string) (*OnCallAssignment, error) {
	var assignment OnCallAssignment
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_assignments WHERE rotation_id = ? AND is_current = 1"
	err := r.db.Sqlx().GetContext(ctx, &assignment, query, rotationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no current assignment found for rotation %s: %w", rotationID, err)
		}
		return nil, fmt.Errorf("error finding current assignment for rotation %s: %w", rotationID, err)
	}
	return &assignment, nil
}

// FindAssignmentsByUser retrieves assignments for a user.
func (r *SQLiteOnCallRepository) FindAssignmentsByUser(ctx context.Context, userID string) ([]OnCallAssignment, error) {
	var assignments []OnCallAssignment
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_assignments WHERE user_id = ?"
	err := r.db.Sqlx().SelectContext(ctx, &assignments, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying assignments by user: %w", err)
	}
	return assignments, nil
}

// FindAssignmentsByRotation retrieves assignments for a rotation.
func (r *SQLiteOnCallRepository) FindAssignmentsByRotation(ctx context.Context, rotationID string) ([]OnCallAssignment, error) {
	var assignments []OnCallAssignment
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_assignments WHERE rotation_id = ?"
	err := r.db.Sqlx().SelectContext(ctx, &assignments, query, rotationID)
	if err != nil {
		return nil, fmt.Errorf("error querying assignments by rotation: %w", err)
	}
	return assignments, nil
}

// CreateAssignment inserts a new assignment.
func (r *SQLiteOnCallRepository) CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	if assignment.ID == "" {
		assignment.ID = uuid.New().String()
	}
	if assignment.CreatedAt.IsZero() {
		assignment.CreatedAt = time.Now()
	}
	if assignment.UpdatedAt.IsZero() {
		assignment.UpdatedAt = time.Now()
	}

	// If this is the current assignment, ensure no other assignments are marked as current
	if assignment.IsCurrent {
		err := r.Transaction(ctx, func(tx OnCallTransaction) error {
			// Clear current flag from all other assignments for this rotation
			resetQuery := "UPDATE oncall_assignments SET is_current = 0, updated_at = CURRENT_TIMESTAMP WHERE rotation_id = ? AND id != ?"
			_, err := r.db.Sqlx().ExecContext(ctx, resetQuery, assignment.RotationID, assignment.ID)
			if err != nil {
				return fmt.Errorf("error resetting current assignments: %w", err)
			}

			// Create the new assignment
			return tx.CreateAssignment(ctx, assignment)
		})
		if err != nil {
			return err
		}
		return nil
	}

	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_assignments 
		(id, rotation_id, user_id, start_time, end_time, is_current, created_at, updated_at) 
		VALUES (:id, :rotation_id, :user_id, :start_time, :end_time, :is_current, :created_at, :updated_at)
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, assignment)
	if err != nil {
		return fmt.Errorf("error creating assignment: %w", err)
	}
	return nil
}

// UpdateAssignment updates an existing assignment.
func (r *SQLiteOnCallRepository) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	assignment.UpdatedAt = time.Now()

	// If this is being marked as current, ensure no other assignments are current
	if assignment.IsCurrent {
		err := r.Transaction(ctx, func(tx OnCallTransaction) error {
			// Clear current flag from all other assignments for this rotation
			resetQuery := "UPDATE oncall_assignments SET is_current = 0, updated_at = CURRENT_TIMESTAMP WHERE rotation_id = ? AND id != ?"
			_, err := r.db.Sqlx().ExecContext(ctx, resetQuery, assignment.RotationID, assignment.ID)
			if err != nil {
				return fmt.Errorf("error resetting current assignments: %w", err)
			}

			// Update the assignment
			return tx.UpdateAssignment(ctx, assignment)
		})
		if err != nil {
			return err
		}
		return nil
	}

	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_assignments 
		SET rotation_id = :rotation_id, 
			user_id = :user_id, 
			start_time = :start_time, 
			end_time = :end_time, 
			is_current = :is_current, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, assignment)
	if err != nil {
		return fmt.Errorf("error updating assignment: %w", err)
	}
	return nil
}

// DeleteAssignment removes an assignment by its ID.
func (r *SQLiteOnCallRepository) DeleteAssignment(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_assignments WHERE id = ?"
	_, err := r.db.Sqlx().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting assignment with id %s: %w", id, err)
	}
	return nil
}

// FindEscalationByID retrieves an escalation by its ID.
func (r *SQLiteOnCallRepository) FindEscalationByID(ctx context.Context, id string) (*OnCallEscalation, error) {
	var escalation OnCallEscalation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_escalations WHERE id = ?"
	err := r.db.Sqlx().GetContext(ctx, &escalation, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("escalation not found with id %s: %w", id, err)
		}
		return nil, fmt.Errorf("error finding escalation with id %s: %w", id, err)
	}
	return &escalation, nil
}

// FindEscalationsByStatus retrieves escalations by status.
func (r *SQLiteOnCallRepository) FindEscalationsByStatus(ctx context.Context, status EscalationStatus) ([]OnCallEscalation, error) {
	var escalations []OnCallEscalation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_escalations WHERE status = ?"
	err := r.db.Sqlx().SelectContext(ctx, &escalations, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("error querying escalations by status: %w", err)
	}
	return escalations, nil
}

// FindEscalationsByAssignment retrieves escalations for an assignment.
func (r *SQLiteOnCallRepository) FindEscalationsByAssignment(ctx context.Context, assignmentID string) ([]OnCallEscalation, error) {
	var escalations []OnCallEscalation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_escalations WHERE assignment_id = ?"
	err := r.db.Sqlx().SelectContext(ctx, &escalations, query, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("error querying escalations by assignment: %w", err)
	}
	return escalations, nil
}

// FindEscalationByIssue retrieves an escalation for an issue.
func (r *SQLiteOnCallRepository) FindEscalationByIssue(ctx context.Context, repository string, issueNumber int) (*OnCallEscalation, error) {
	var escalation OnCallEscalation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_escalations WHERE repository = ? AND issue_number = ?"
	err := r.db.Sqlx().GetContext(ctx, &escalation, query, repository, issueNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("escalation not found for issue %d in repository %s: %w", issueNumber, repository, err)
		}
		return nil, fmt.Errorf("error finding escalation for issue %d in repository %s: %w", issueNumber, repository, err)
	}
	return &escalation, nil
}

// FindEscalationByPR retrieves an escalation for a PR.
func (r *SQLiteOnCallRepository) FindEscalationByPR(ctx context.Context, repository string, prNumber int) (*OnCallEscalation, error) {
	var escalation OnCallEscalation
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT * FROM oncall_escalations WHERE repository = ? AND pr_number = ?"
	err := r.db.Sqlx().GetContext(ctx, &escalation, query, repository, prNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("escalation not found for PR %d in repository %s: %w", prNumber, repository, err)
		}
		return nil, fmt.Errorf("error finding escalation for PR %d in repository %s: %w", prNumber, repository, err)
	}
	return &escalation, nil
}

// CreateEscalation inserts a new escalation.
func (r *SQLiteOnCallRepository) CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	if escalation.ID == "" {
		escalation.ID = uuid.New().String()
	}
	if escalation.CreatedAt.IsZero() {
		escalation.CreatedAt = time.Now()
	}
	if escalation.UpdatedAt.IsZero() {
		escalation.UpdatedAt = time.Now()
	}

	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_escalations 
		(id, assignment_id, issue_number, pr_number, repository, status, escalation_time, resolution_time, created_at, updated_at) 
		VALUES (:id, :assignment_id, :issue_number, :pr_number, :repository, :status, :escalation_time, :resolution_time, :created_at, :updated_at)
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, escalation)
	if err != nil {
		return fmt.Errorf("error creating escalation: %w", err)
	}
	return nil
}

// UpdateEscalation updates an existing escalation.
func (r *SQLiteOnCallRepository) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	escalation.UpdatedAt = time.Now()

	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_escalations 
		SET assignment_id = :assignment_id, 
			issue_number = :issue_number, 
			pr_number = :pr_number, 
			repository = :repository, 
			status = :status, 
			escalation_time = :escalation_time, 
			resolution_time = :resolution_time, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := r.db.Sqlx().NamedExecContext(ctx, query, escalation)
	if err != nil {
		return fmt.Errorf("error updating escalation: %w", err)
	}
	return nil
}

// DeleteEscalation removes an escalation by its ID.
func (r *SQLiteOnCallRepository) DeleteEscalation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_escalations WHERE id = ?"
	_, err := r.db.Sqlx().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting escalation with id %s: %w", id, err)
	}
	return nil
}

// Transaction executes a function within a database transaction.
func (r *SQLiteOnCallRepository) Transaction(ctx context.Context, fn func(tx OnCallTransaction) error) error {
	// Begin a database transaction
	sqlxTx, err := r.db.Sqlx().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Create a transaction-scoped repository
	txRepo := &SQLiteOnCallTransaction{
		tx: sqlxTx,
	}

	// Execute the function
	if err := fn(txRepo); err != nil {
		// Attempt rollback, but don't override original error
		rollbackErr := sqlxTx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback failed: %w)", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Commit the transaction
	if err := sqlxTx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// EnsureSchema ensures the database schema is up to date.
func (r *SQLiteOnCallRepository) EnsureSchema(ctx context.Context) error {
	// This method is no longer needed since we're using migrations
	// The migrator.MigrateUp function should be called from the application initialization
	return nil
}

// SQLiteOnCallTransaction implements OnCallTransaction for SQLite databases.
type SQLiteOnCallTransaction struct {
	tx *sqlx.Tx
}

// Ensure SQLiteOnCallTransaction implements OnCallTransaction.
var _ OnCallTransaction = (*SQLiteOnCallTransaction)(nil)

// CreateUser inserts a new user within a transaction.
func (t *SQLiteOnCallTransaction) CreateUser(ctx context.Context, user *OnCallUser) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_users 
		(id, github_username, name, email, is_active, created_at, updated_at) 
		VALUES (:id, :github_username, :name, :email, :is_active, :created_at, :updated_at)
	`
	_, err := t.tx.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("error creating user in transaction: %w", err)
	}
	return nil
}

// UpdateUser updates an existing user within a transaction.
func (t *SQLiteOnCallTransaction) UpdateUser(ctx context.Context, user *OnCallUser) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_users 
		SET github_username = :github_username, 
			name = :name, 
			email = :email, 
			is_active = :is_active, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := t.tx.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("error updating user in transaction: %w", err)
	}
	return nil
}

// DeleteUser removes a user by their ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteUser(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_users WHERE id = ?"
	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user in transaction with id %s: %w", id, err)
	}
	return nil
}

// CreateRotation inserts a new rotation within a transaction.
func (t *SQLiteOnCallTransaction) CreateRotation(ctx context.Context, rotation *OnCallRotation) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_rotations 
		(id, name, description, repository, is_active, created_at, updated_at) 
		VALUES (:id, :name, :description, :repository, :is_active, :created_at, :updated_at)
	`
	_, err := t.tx.NamedExecContext(ctx, query, rotation)
	if err != nil {
		return fmt.Errorf("error creating rotation in transaction: %w", err)
	}
	return nil
}

// UpdateRotation updates an existing rotation within a transaction.
func (t *SQLiteOnCallTransaction) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_rotations 
		SET name = :name, 
			description = :description, 
			repository = :repository, 
			is_active = :is_active, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := t.tx.NamedExecContext(ctx, query, rotation)
	if err != nil {
		return fmt.Errorf("error updating rotation in transaction: %w", err)
	}
	return nil
}

// DeleteRotation removes a rotation by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteRotation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_rotations WHERE id = ?"
	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting rotation in transaction with id %s: %w", id, err)
	}
	return nil
}

// CreateAssignment inserts a new assignment within a transaction.
func (t *SQLiteOnCallTransaction) CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_assignments 
		(id, rotation_id, user_id, start_time, end_time, is_current, created_at, updated_at) 
		VALUES (:id, :rotation_id, :user_id, :start_time, :end_time, :is_current, :created_at, :updated_at)
	`
	_, err := t.tx.NamedExecContext(ctx, query, assignment)
	if err != nil {
		return fmt.Errorf("error creating assignment in transaction: %w", err)
	}
	return nil
}

// UpdateAssignment updates an existing assignment within a transaction.
func (t *SQLiteOnCallTransaction) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_assignments 
		SET rotation_id = :rotation_id, 
			user_id = :user_id, 
			start_time = :start_time, 
			end_time = :end_time, 
			is_current = :is_current, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := t.tx.NamedExecContext(ctx, query, assignment)
	if err != nil {
		return fmt.Errorf("error updating assignment in transaction: %w", err)
	}
	return nil
}

// DeleteAssignment removes an assignment by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteAssignment(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_assignments WHERE id = ?"
	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting assignment in transaction with id %s: %w", id, err)
	}
	return nil
}

// CreateEscalation inserts a new escalation within a transaction.
func (t *SQLiteOnCallTransaction) CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		INSERT INTO oncall_escalations 
		(id, assignment_id, issue_number, pr_number, repository, status, escalation_time, resolution_time, created_at, updated_at) 
		VALUES (:id, :assignment_id, :issue_number, :pr_number, :repository, :status, :escalation_time, :resolution_time, :created_at, :updated_at)
	`
	_, err := t.tx.NamedExecContext(ctx, query, escalation)
	if err != nil {
		return fmt.Errorf("error creating escalation in transaction: %w", err)
	}
	return nil
}

// UpdateEscalation updates an existing escalation within a transaction.
func (t *SQLiteOnCallTransaction) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	// #nosec G202 -- Table name is validated at initialization
	query := `
		UPDATE oncall_escalations 
		SET assignment_id = :assignment_id, 
			issue_number = :issue_number, 
			pr_number = :pr_number, 
			repository = :repository, 
			status = :status, 
			escalation_time = :escalation_time, 
			resolution_time = :resolution_time, 
			updated_at = :updated_at 
		WHERE id = :id
	`
	_, err := t.tx.NamedExecContext(ctx, query, escalation)
	if err != nil {
		return fmt.Errorf("error updating escalation in transaction: %w", err)
	}
	return nil
}

// DeleteEscalation removes an escalation by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteEscalation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM oncall_escalations WHERE id = ?"
	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting escalation in transaction with id %s: %w", id, err)
	}
	return nil
}