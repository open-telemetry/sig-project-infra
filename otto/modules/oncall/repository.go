// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
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

// SQLiteOnCallRepository implements OnCallRepository for SQLite databases.
type SQLiteOnCallRepository struct {
	db database.Provider

	// Generic repositories for each entity type
	userRepo       database.Repository[OnCallUser]
	rotationRepo   database.Repository[OnCallRotation]
	assignmentRepo database.Repository[OnCallAssignment]
	escalationRepo database.Repository[OnCallEscalation]
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

// validateMapperTableNames validates all table names used by the repository.
func validateMapperTableNames() error {
	mappers := []string{
		(&OnCallUserMapper{}).Table(),
		(&OnCallRotationMapper{}).Table(),
		(&OnCallAssignmentMapper{}).Table(),
		(&OnCallEscalationMapper{}).Table(),
	}

	for _, tableName := range mappers {
		if err := validateTableName(tableName); err != nil {
			return err
		}
	}

	return nil
}

// NewSQLiteOnCallRepository creates a new SQLite-backed OnCallRepository.
func NewSQLiteOnCallRepository(db database.Provider) (OnCallRepository, error) {
	// Validate all table names at initialization time
	if err := validateMapperTableNames(); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create mappers for each entity type
	userMapper := &OnCallUserMapper{}
	rotationMapper := &OnCallRotationMapper{}
	assignmentMapper := &OnCallAssignmentMapper{}
	escalationMapper := &OnCallEscalationMapper{}

	// Create repositories using the mappers
	var err error
	userRepo, err := database.NewSQLRepository(db, userMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	rotationRepo, err := database.NewSQLRepository(db, rotationMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create rotation repository: %w", err)
	}

	assignmentRepo, err := database.NewSQLRepository(db, assignmentMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment repository: %w", err)
	}

	escalationRepo, err := database.NewSQLRepository(db, escalationMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create escalation repository: %w", err)
	}

	return &SQLiteOnCallRepository{
		db:             db,
		userRepo:       userRepo,
		rotationRepo:   rotationRepo,
		assignmentRepo: assignmentRepo,
		escalationRepo: escalationRepo,
	}, nil
}

// FindUserByID retrieves a user by their ID.
func (r *SQLiteOnCallRepository) FindUserByID(ctx context.Context, id string) (*OnCallUser, error) {
	return r.userRepo.FindByID(ctx, id)
}

// FindUserByGitHubUsername retrieves a user by their GitHub username.
func (r *SQLiteOnCallRepository) FindUserByGitHubUsername(ctx context.Context, username string) (*OnCallUser, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallUserMapper{}).Columns()) +
		" FROM " + (&OnCallUserMapper{}).Table() + " WHERE github_username = ?"

	row := r.db.DB().QueryRowContext(ctx, query, username)
	return (&OnCallUserMapper{}).FromRow(row)
}

// FindAllUsers retrieves all users.
func (r *SQLiteOnCallRepository) FindAllUsers(ctx context.Context) ([]OnCallUser, error) {
	return r.userRepo.FindAll(ctx)
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
	return r.userRepo.Create(ctx, user)
}

// UpdateUser updates an existing user.
func (r *SQLiteOnCallRepository) UpdateUser(ctx context.Context, user *OnCallUser) error {
	user.UpdatedAt = time.Now()
	return r.userRepo.Update(ctx, user)
}

// DeleteUser removes a user by their ID.
func (r *SQLiteOnCallRepository) DeleteUser(ctx context.Context, id string) error {
	return r.userRepo.Delete(ctx, id)
}

// FindRotationByID retrieves a rotation by its ID.
func (r *SQLiteOnCallRepository) FindRotationByID(ctx context.Context, id string) (*OnCallRotation, error) {
	return r.rotationRepo.FindByID(ctx, id)
}

// FindRotationsByRepository retrieves rotations by repository.
func (r *SQLiteOnCallRepository) FindRotationsByRepository(
	ctx context.Context,
	repository string,
) ([]OnCallRotation, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallRotationMapper{}).Columns()) +
		" FROM " + (&OnCallRotationMapper{}).Table() + " WHERE repository = ?"

	rows, err := r.db.DB().QueryContext(ctx, query, repository)
	if err != nil {
		return nil, fmt.Errorf("error querying rotations by repository: %w", err)
	}
	defer rows.Close()

	return (&OnCallRotationMapper{}).FromRows(rows)
}

// FindAllRotations retrieves all rotations.
func (r *SQLiteOnCallRepository) FindAllRotations(ctx context.Context) ([]OnCallRotation, error) {
	return r.rotationRepo.FindAll(ctx)
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
	return r.rotationRepo.Create(ctx, rotation)
}

// UpdateRotation updates an existing rotation.
func (r *SQLiteOnCallRepository) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	rotation.UpdatedAt = time.Now()
	return r.rotationRepo.Update(ctx, rotation)
}

// DeleteRotation removes a rotation by its ID.
func (r *SQLiteOnCallRepository) DeleteRotation(ctx context.Context, id string) error {
	return r.rotationRepo.Delete(ctx, id)
}

// FindAssignmentByID retrieves an assignment by its ID.
func (r *SQLiteOnCallRepository) FindAssignmentByID(ctx context.Context, id string) (*OnCallAssignment, error) {
	return r.assignmentRepo.FindByID(ctx, id)
}

// FindCurrentAssignmentByRotation retrieves the current assignment for a rotation.
func (r *SQLiteOnCallRepository) FindCurrentAssignmentByRotation(
	ctx context.Context,
	rotationID string,
) (*OnCallAssignment, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallAssignmentMapper{}).Columns()) +
		" FROM " + (&OnCallAssignmentMapper{}).Table() + " WHERE rotation_id = ? AND is_current = 1"

	row := r.db.DB().QueryRowContext(ctx, query, rotationID)
	return (&OnCallAssignmentMapper{}).FromRow(row)
}

// FindAssignmentsByUser retrieves assignments for a user.
func (r *SQLiteOnCallRepository) FindAssignmentsByUser(ctx context.Context, userID string) ([]OnCallAssignment, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallAssignmentMapper{}).Columns()) +
		" FROM " + (&OnCallAssignmentMapper{}).Table() + " WHERE user_id = ?"

	rows, err := r.db.DB().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying assignments by user: %w", err)
	}
	defer rows.Close()

	return (&OnCallAssignmentMapper{}).FromRows(rows)
}

// FindAssignmentsByRotation retrieves assignments for a rotation.
func (r *SQLiteOnCallRepository) FindAssignmentsByRotation(
	ctx context.Context,
	rotationID string,
) ([]OnCallAssignment, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallAssignmentMapper{}).Columns()) +
		" FROM " + (&OnCallAssignmentMapper{}).Table() + " WHERE rotation_id = ?"

	rows, err := r.db.DB().QueryContext(ctx, query, rotationID)
	if err != nil {
		return nil, fmt.Errorf("error querying assignments by rotation: %w", err)
	}
	defer rows.Close()

	return (&OnCallAssignmentMapper{}).FromRows(rows)
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
			_, err := r.db.DB().ExecContext(ctx, resetQuery, assignment.RotationID, assignment.ID)
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

	return r.assignmentRepo.Create(ctx, assignment)
}

// UpdateAssignment updates an existing assignment.
func (r *SQLiteOnCallRepository) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	assignment.UpdatedAt = time.Now()

	// If this is being marked as current, ensure no other assignments are current
	if assignment.IsCurrent {
		err := r.Transaction(ctx, func(tx OnCallTransaction) error {
			// Clear current flag from all other assignments for this rotation
			resetQuery := "UPDATE oncall_assignments SET is_current = 0, updated_at = CURRENT_TIMESTAMP WHERE rotation_id = ? AND id != ?"
			_, err := r.db.DB().ExecContext(ctx, resetQuery, assignment.RotationID, assignment.ID)
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

	return r.assignmentRepo.Update(ctx, assignment)
}

// DeleteAssignment removes an assignment by its ID.
func (r *SQLiteOnCallRepository) DeleteAssignment(ctx context.Context, id string) error {
	return r.assignmentRepo.Delete(ctx, id)
}

// FindEscalationByID retrieves an escalation by its ID.
func (r *SQLiteOnCallRepository) FindEscalationByID(ctx context.Context, id string) (*OnCallEscalation, error) {
	return r.escalationRepo.FindByID(ctx, id)
}

// FindEscalationsByStatus retrieves escalations by status.
func (r *SQLiteOnCallRepository) FindEscalationsByStatus(
	ctx context.Context,
	status EscalationStatus,
) ([]OnCallEscalation, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallEscalationMapper{}).Columns()) +
		" FROM " + (&OnCallEscalationMapper{}).Table() + " WHERE status = ?"

	rows, err := r.db.DB().QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("error querying escalations by status: %w", err)
	}
	defer rows.Close()

	return (&OnCallEscalationMapper{}).FromRows(rows)
}

// FindEscalationsByAssignment retrieves escalations for an assignment.
func (r *SQLiteOnCallRepository) FindEscalationsByAssignment(
	ctx context.Context,
	assignmentID string,
) ([]OnCallEscalation, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallEscalationMapper{}).Columns()) +
		" FROM " + (&OnCallEscalationMapper{}).Table() + " WHERE assignment_id = ?"

	rows, err := r.db.DB().QueryContext(ctx, query, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("error querying escalations by assignment: %w", err)
	}
	defer rows.Close()

	return (&OnCallEscalationMapper{}).FromRows(rows)
}

// FindEscalationByIssue retrieves an escalation for an issue.
func (r *SQLiteOnCallRepository) FindEscalationByIssue(
	ctx context.Context,
	repository string,
	issueNumber int,
) (*OnCallEscalation, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallEscalationMapper{}).Columns()) +
		" FROM " + (&OnCallEscalationMapper{}).Table() + " WHERE repository = ? AND issue_number = ?"

	row := r.db.DB().QueryRowContext(ctx, query, repository, issueNumber)
	return (&OnCallEscalationMapper{}).FromRow(row)
}

// FindEscalationByPR retrieves an escalation for a PR.
func (r *SQLiteOnCallRepository) FindEscalationByPR(
	ctx context.Context,
	repository string,
	prNumber int,
) (*OnCallEscalation, error) {
	// #nosec G202 -- Table name is validated at initialization
	query := "SELECT " + joinColumns((&OnCallEscalationMapper{}).Columns()) +
		" FROM " + (&OnCallEscalationMapper{}).Table() + " WHERE repository = ? AND pr_number = ?"

	row := r.db.DB().QueryRowContext(ctx, query, repository, prNumber)
	return (&OnCallEscalationMapper{}).FromRow(row)
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
	return r.escalationRepo.Create(ctx, escalation)
}

// UpdateEscalation updates an existing escalation.
func (r *SQLiteOnCallRepository) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	escalation.UpdatedAt = time.Now()
	return r.escalationRepo.Update(ctx, escalation)
}

// DeleteEscalation removes an escalation by its ID.
func (r *SQLiteOnCallRepository) DeleteEscalation(ctx context.Context, id string) error {
	return r.escalationRepo.Delete(ctx, id)
}

// Transaction executes a function within a database transaction.
func (r *SQLiteOnCallRepository) Transaction(ctx context.Context, fn func(tx OnCallTransaction) error) error {
	// Begin a database transaction
	sqlTx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Create a transaction-scoped repository
	txRepo := &SQLiteOnCallTransaction{
		tx: sqlTx,
	}

	// Execute the function
	if err := fn(txRepo); err != nil {
		// Attempt rollback, but don't override original error
		rollbackErr := sqlTx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback failed: %w)", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Commit the transaction
	if err := sqlTx.Commit(); err != nil {
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
	tx *sql.Tx
}

// Ensure SQLiteOnCallTransaction implements OnCallTransaction.
var _ OnCallTransaction = (*SQLiteOnCallTransaction)(nil)

// CreateUser inserts a new user within a transaction.
func (t *SQLiteOnCallTransaction) CreateUser(ctx context.Context, user *OnCallUser) error {
	columns := (&OnCallUserMapper{}).Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + (&OnCallUserMapper{}).Table() +
		" (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := t.tx.ExecContext(ctx, query, (&OnCallUserMapper{}).ToRow(user)...)
	if err != nil {
		return fmt.Errorf("error creating user in transaction: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user within a transaction.
func (t *SQLiteOnCallTransaction) UpdateUser(ctx context.Context, user *OnCallUser) error {
	columns := (&OnCallUserMapper{}).Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + (&OnCallUserMapper{}).Table() +
		" SET " + setClause + " WHERE id = ?"

	args := (&OnCallUserMapper{}).ToRow(user)
	args = append(args, user.ID)

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating user in transaction: %w", err)
	}

	return nil
}

// DeleteUser removes a user by their ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteUser(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + (&OnCallUserMapper{}).Table() + " WHERE id = ?"

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user in transaction: %w", err)
	}

	return nil
}

// CreateRotation inserts a new rotation within a transaction.
func (t *SQLiteOnCallTransaction) CreateRotation(ctx context.Context, rotation *OnCallRotation) error {
	columns := (&OnCallRotationMapper{}).Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + (&OnCallRotationMapper{}).Table() +
		" (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := t.tx.ExecContext(ctx, query, (&OnCallRotationMapper{}).ToRow(rotation)...)
	if err != nil {
		return fmt.Errorf("error creating rotation in transaction: %w", err)
	}

	return nil
}

// UpdateRotation updates an existing rotation within a transaction.
func (t *SQLiteOnCallTransaction) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	columns := (&OnCallRotationMapper{}).Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + (&OnCallRotationMapper{}).Table() +
		" SET " + setClause + " WHERE id = ?"

	args := (&OnCallRotationMapper{}).ToRow(rotation)
	args = append(args, rotation.ID)

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating rotation in transaction: %w", err)
	}

	return nil
}

// DeleteRotation removes a rotation by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteRotation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + (&OnCallRotationMapper{}).Table() + " WHERE id = ?"

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting rotation in transaction: %w", err)
	}

	return nil
}

// CreateAssignment inserts a new assignment within a transaction.
func (t *SQLiteOnCallTransaction) CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	columns := (&OnCallAssignmentMapper{}).Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + (&OnCallAssignmentMapper{}).Table() +
		" (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := t.tx.ExecContext(ctx, query, (&OnCallAssignmentMapper{}).ToRow(assignment)...)
	if err != nil {
		return fmt.Errorf("error creating assignment in transaction: %w", err)
	}

	return nil
}

// UpdateAssignment updates an existing assignment within a transaction.
func (t *SQLiteOnCallTransaction) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	columns := (&OnCallAssignmentMapper{}).Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + (&OnCallAssignmentMapper{}).Table() +
		" SET " + setClause + " WHERE id = ?"

	args := (&OnCallAssignmentMapper{}).ToRow(assignment)
	args = append(args, assignment.ID)

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating assignment in transaction: %w", err)
	}

	return nil
}

// DeleteAssignment removes an assignment by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteAssignment(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + (&OnCallAssignmentMapper{}).Table() + " WHERE id = ?"

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting assignment in transaction: %w", err)
	}

	return nil
}

// CreateEscalation inserts a new escalation within a transaction.
func (t *SQLiteOnCallTransaction) CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	columns := (&OnCallEscalationMapper{}).Columns()
	placeholders := createPlaceholders(len(columns))
	// #nosec G202 -- Table name is validated at initialization
	query := "INSERT INTO " + (&OnCallEscalationMapper{}).Table() +
		" (" + joinColumns(columns) + ") VALUES (" + placeholders + ")"

	_, err := t.tx.ExecContext(ctx, query, (&OnCallEscalationMapper{}).ToRow(escalation)...)
	if err != nil {
		return fmt.Errorf("error creating escalation in transaction: %w", err)
	}

	return nil
}

// UpdateEscalation updates an existing escalation within a transaction.
func (t *SQLiteOnCallTransaction) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	columns := (&OnCallEscalationMapper{}).Columns()
	setClause := createSetClause(columns)
	// #nosec G202 -- Table name is validated at initialization
	query := "UPDATE " + (&OnCallEscalationMapper{}).Table() +
		" SET " + setClause + " WHERE id = ?"

	args := (&OnCallEscalationMapper{}).ToRow(escalation)
	args = append(args, escalation.ID)

	_, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating escalation in transaction: %w", err)
	}

	return nil
}

// DeleteEscalation removes an escalation by its ID within a transaction.
func (t *SQLiteOnCallTransaction) DeleteEscalation(ctx context.Context, id string) error {
	// #nosec G202 -- Table name is validated at initialization
	query := "DELETE FROM " + (&OnCallEscalationMapper{}).Table() + " WHERE id = ?"

	_, err := t.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting escalation in transaction: %w", err)
	}

	return nil
}

// Helper functions for SQL query generation

func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

func createPlaceholders(count int) string {
	result := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ", "
		}
		result += "?"
	}
	return result
}

func createSetClause(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col + " = ?"
	}
	return result
}
