// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockOnCallRepository implements OnCallRepository for testing.
type MockOnCallRepository struct {
	users       map[string]*OnCallUser
	rotations   map[string]*OnCallRotation
	assignments map[string]*OnCallAssignment
	escalations map[string]*OnCallEscalation
	mu          sync.RWMutex
}

// Ensure MockOnCallRepository implements OnCallRepository.
var _ OnCallRepository = (*MockOnCallRepository)(nil)

// NewMockOnCallRepository creates a new in-memory repository for testing.
func NewMockOnCallRepository() *MockOnCallRepository {
	return &MockOnCallRepository{
		users:       make(map[string]*OnCallUser),
		rotations:   make(map[string]*OnCallRotation),
		assignments: make(map[string]*OnCallAssignment),
		escalations: make(map[string]*OnCallEscalation),
	}
}

// FindUserByID retrieves a user by their ID.
func (r *MockOnCallRepository) FindUserByID(ctx context.Context, id string) (*OnCallUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

// FindUserByGitHubUsername retrieves a user by their GitHub username.
func (r *MockOnCallRepository) FindUserByGitHubUsername(ctx context.Context, username string) (*OnCallUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.GitHubUsername == username {
			return user, nil
		}
	}
	return nil, nil
}

// FindAllUsers retrieves all users.
func (r *MockOnCallRepository) FindAllUsers(ctx context.Context) ([]OnCallUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]OnCallUser, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, nil
}

// CreateUser inserts a new user.
func (r *MockOnCallRepository) CreateUser(ctx context.Context, user *OnCallUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	r.users[user.ID] = user
	return nil
}

// UpdateUser updates an existing user.
func (r *MockOnCallRepository) UpdateUser(ctx context.Context, user *OnCallUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return nil
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

// DeleteUser removes a user by their ID.
func (r *MockOnCallRepository) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, id)
	return nil
}

// FindRotationByID retrieves a rotation by its ID.
func (r *MockOnCallRepository) FindRotationByID(ctx context.Context, id string) (*OnCallRotation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rotation, ok := r.rotations[id]
	if !ok {
		return nil, nil
	}
	return rotation, nil
}

// FindRotationsByRepository retrieves rotations by repository.
func (r *MockOnCallRepository) FindRotationsByRepository(
	ctx context.Context,
	repository string,
) ([]OnCallRotation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var rotations []OnCallRotation
	for _, rotation := range r.rotations {
		if rotation.Repository == repository {
			rotations = append(rotations, *rotation)
		}
	}
	return rotations, nil
}

// FindAllRotations retrieves all rotations.
func (r *MockOnCallRepository) FindAllRotations(ctx context.Context) ([]OnCallRotation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rotations := make([]OnCallRotation, 0, len(r.rotations))
	for _, rotation := range r.rotations {
		rotations = append(rotations, *rotation)
	}
	return rotations, nil
}

// CreateRotation inserts a new rotation.
func (r *MockOnCallRepository) CreateRotation(ctx context.Context, rotation *OnCallRotation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rotation.ID == "" {
		rotation.ID = uuid.New().String()
	}
	if rotation.CreatedAt.IsZero() {
		rotation.CreatedAt = time.Now()
	}
	if rotation.UpdatedAt.IsZero() {
		rotation.UpdatedAt = time.Now()
	}

	r.rotations[rotation.ID] = rotation
	return nil
}

// UpdateRotation updates an existing rotation.
func (r *MockOnCallRepository) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.rotations[rotation.ID]; !ok {
		return nil
	}

	rotation.UpdatedAt = time.Now()
	r.rotations[rotation.ID] = rotation
	return nil
}

// DeleteRotation removes a rotation by its ID.
func (r *MockOnCallRepository) DeleteRotation(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rotations, id)
	return nil
}

// FindAssignmentByID retrieves an assignment by its ID.
func (r *MockOnCallRepository) FindAssignmentByID(ctx context.Context, id string) (*OnCallAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	assignment, ok := r.assignments[id]
	if !ok {
		return nil, nil
	}
	return assignment, nil
}

// FindCurrentAssignmentByRotation retrieves the current assignment for a rotation.
func (r *MockOnCallRepository) FindCurrentAssignmentByRotation(
	ctx context.Context,
	rotationID string,
) (*OnCallAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, assignment := range r.assignments {
		if assignment.RotationID == rotationID && assignment.IsCurrent {
			return assignment, nil
		}
	}
	return nil, nil
}

// FindAssignmentsByUser retrieves assignments for a user.
func (r *MockOnCallRepository) FindAssignmentsByUser(ctx context.Context, userID string) ([]OnCallAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var assignments []OnCallAssignment
	for _, assignment := range r.assignments {
		if assignment.UserID == userID {
			assignments = append(assignments, *assignment)
		}
	}
	return assignments, nil
}

// FindAssignmentsByRotation retrieves assignments for a rotation.
func (r *MockOnCallRepository) FindAssignmentsByRotation(
	ctx context.Context,
	rotationID string,
) ([]OnCallAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var assignments []OnCallAssignment
	for _, assignment := range r.assignments {
		if assignment.RotationID == rotationID {
			assignments = append(assignments, *assignment)
		}
	}
	return assignments, nil
}

// CreateAssignment inserts a new assignment.
func (r *MockOnCallRepository) CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
		for id, a := range r.assignments {
			if a.RotationID == assignment.RotationID && a.ID != assignment.ID {
				a.IsCurrent = false
				a.UpdatedAt = time.Now()
				r.assignments[id] = a
			}
		}
	}

	r.assignments[assignment.ID] = assignment
	return nil
}

// UpdateAssignment updates an existing assignment.
func (r *MockOnCallRepository) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.assignments[assignment.ID]; !ok {
		return nil
	}

	assignment.UpdatedAt = time.Now()

	// If this is being marked as current, ensure no other assignments are current
	if assignment.IsCurrent {
		for id, a := range r.assignments {
			if a.RotationID == assignment.RotationID && a.ID != assignment.ID {
				a.IsCurrent = false
				a.UpdatedAt = time.Now()
				r.assignments[id] = a
			}
		}
	}

	r.assignments[assignment.ID] = assignment
	return nil
}

// DeleteAssignment removes an assignment by its ID.
func (r *MockOnCallRepository) DeleteAssignment(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.assignments, id)
	return nil
}

// FindEscalationByID retrieves an escalation by its ID.
func (r *MockOnCallRepository) FindEscalationByID(ctx context.Context, id string) (*OnCallEscalation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	escalation, ok := r.escalations[id]
	if !ok {
		return nil, nil
	}
	return escalation, nil
}

// FindEscalationsByStatus retrieves escalations by status.
func (r *MockOnCallRepository) FindEscalationsByStatus(
	ctx context.Context,
	status EscalationStatus,
) ([]OnCallEscalation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var escalations []OnCallEscalation
	for _, escalation := range r.escalations {
		if escalation.Status == status {
			escalations = append(escalations, *escalation)
		}
	}
	return escalations, nil
}

// FindEscalationsByAssignment retrieves escalations for an assignment.
func (r *MockOnCallRepository) FindEscalationsByAssignment(
	ctx context.Context,
	assignmentID string,
) ([]OnCallEscalation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var escalations []OnCallEscalation
	for _, escalation := range r.escalations {
		if escalation.AssignmentID == assignmentID {
			escalations = append(escalations, *escalation)
		}
	}
	return escalations, nil
}

// FindEscalationByIssue retrieves an escalation for an issue.
func (r *MockOnCallRepository) FindEscalationByIssue(
	ctx context.Context,
	repository string,
	issueNumber int,
) (*OnCallEscalation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, escalation := range r.escalations {
		if escalation.Repository == repository && escalation.IssueNumber == issueNumber {
			return escalation, nil
		}
	}
	return nil, nil
}

// FindEscalationByPR retrieves an escalation for a PR.
func (r *MockOnCallRepository) FindEscalationByPR(
	ctx context.Context,
	repository string,
	prNumber int,
) (*OnCallEscalation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, escalation := range r.escalations {
		if escalation.Repository == repository && escalation.PRNumber == prNumber {
			return escalation, nil
		}
	}
	return nil, nil
}

// CreateEscalation inserts a new escalation.
func (r *MockOnCallRepository) CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if escalation.ID == "" {
		escalation.ID = uuid.New().String()
	}
	if escalation.CreatedAt.IsZero() {
		escalation.CreatedAt = time.Now()
	}
	if escalation.UpdatedAt.IsZero() {
		escalation.UpdatedAt = time.Now()
	}

	r.escalations[escalation.ID] = escalation
	return nil
}

// UpdateEscalation updates an existing escalation.
func (r *MockOnCallRepository) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.escalations[escalation.ID]; !ok {
		return nil
	}

	escalation.UpdatedAt = time.Now()
	r.escalations[escalation.ID] = escalation
	return nil
}

// DeleteEscalation removes an escalation by its ID.
func (r *MockOnCallRepository) DeleteEscalation(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.escalations, id)
	return nil
}

// Transaction executes a function within a database transaction.
func (r *MockOnCallRepository) Transaction(ctx context.Context, fn func(tx OnCallTransaction) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	txRepo := &MockOnCallTransaction{repo: r}
	return fn(txRepo)
}

// EnsureSchema ensures the database schema is up to date.
func (r *MockOnCallRepository) EnsureSchema(ctx context.Context) error {
	return nil
}

// MockOnCallTransaction implements OnCallTransaction for testing.
type MockOnCallTransaction struct {
	repo *MockOnCallRepository
}

// Ensure MockOnCallTransaction implements OnCallTransaction.
var _ OnCallTransaction = (*MockOnCallTransaction)(nil)

// CreateUser inserts a new user within a transaction.
func (t *MockOnCallTransaction) CreateUser(ctx context.Context, user *OnCallUser) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	t.repo.users[user.ID] = user
	return nil
}

// UpdateUser updates an existing user within a transaction.
func (t *MockOnCallTransaction) UpdateUser(ctx context.Context, user *OnCallUser) error {
	if _, ok := t.repo.users[user.ID]; !ok {
		return nil
	}

	user.UpdatedAt = time.Now()
	t.repo.users[user.ID] = user
	return nil
}

// DeleteUser removes a user by their ID within a transaction.
func (t *MockOnCallTransaction) DeleteUser(ctx context.Context, id string) error {
	delete(t.repo.users, id)
	return nil
}

// CreateRotation inserts a new rotation within a transaction.
func (t *MockOnCallTransaction) CreateRotation(ctx context.Context, rotation *OnCallRotation) error {
	if rotation.ID == "" {
		rotation.ID = uuid.New().String()
	}
	if rotation.CreatedAt.IsZero() {
		rotation.CreatedAt = time.Now()
	}
	if rotation.UpdatedAt.IsZero() {
		rotation.UpdatedAt = time.Now()
	}

	t.repo.rotations[rotation.ID] = rotation
	return nil
}

// UpdateRotation updates an existing rotation within a transaction.
func (t *MockOnCallTransaction) UpdateRotation(ctx context.Context, rotation *OnCallRotation) error {
	if _, ok := t.repo.rotations[rotation.ID]; !ok {
		return nil
	}

	rotation.UpdatedAt = time.Now()
	t.repo.rotations[rotation.ID] = rotation
	return nil
}

// DeleteRotation removes a rotation by its ID within a transaction.
func (t *MockOnCallTransaction) DeleteRotation(ctx context.Context, id string) error {
	delete(t.repo.rotations, id)
	return nil
}

// CreateAssignment inserts a new assignment within a transaction.
func (t *MockOnCallTransaction) CreateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
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
		for id, a := range t.repo.assignments {
			if a.RotationID == assignment.RotationID && a.ID != assignment.ID {
				a.IsCurrent = false
				a.UpdatedAt = time.Now()
				t.repo.assignments[id] = a
			}
		}
	}

	t.repo.assignments[assignment.ID] = assignment
	return nil
}

// UpdateAssignment updates an existing assignment within a transaction.
func (t *MockOnCallTransaction) UpdateAssignment(ctx context.Context, assignment *OnCallAssignment) error {
	if _, ok := t.repo.assignments[assignment.ID]; !ok {
		return nil
	}

	assignment.UpdatedAt = time.Now()

	// If this is being marked as current, ensure no other assignments are current
	if assignment.IsCurrent {
		for id, a := range t.repo.assignments {
			if a.RotationID == assignment.RotationID && a.ID != assignment.ID {
				a.IsCurrent = false
				a.UpdatedAt = time.Now()
				t.repo.assignments[id] = a
			}
		}
	}

	t.repo.assignments[assignment.ID] = assignment
	return nil
}

// DeleteAssignment removes an assignment by its ID within a transaction.
func (t *MockOnCallTransaction) DeleteAssignment(ctx context.Context, id string) error {
	delete(t.repo.assignments, id)
	return nil
}

// CreateEscalation inserts a new escalation within a transaction.
func (t *MockOnCallTransaction) CreateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	if escalation.ID == "" {
		escalation.ID = uuid.New().String()
	}
	if escalation.CreatedAt.IsZero() {
		escalation.CreatedAt = time.Now()
	}
	if escalation.UpdatedAt.IsZero() {
		escalation.UpdatedAt = time.Now()
	}

	t.repo.escalations[escalation.ID] = escalation
	return nil
}

// UpdateEscalation updates an existing escalation within a transaction.
func (t *MockOnCallTransaction) UpdateEscalation(ctx context.Context, escalation *OnCallEscalation) error {
	if _, ok := t.repo.escalations[escalation.ID]; !ok {
		return nil
	}

	escalation.UpdatedAt = time.Now()
	t.repo.escalations[escalation.ID] = escalation
	return nil
}

// DeleteEscalation removes an escalation by its ID within a transaction.
func (t *MockOnCallTransaction) DeleteEscalation(ctx context.Context, id string) error {
	delete(t.repo.escalations, id)
	return nil
}