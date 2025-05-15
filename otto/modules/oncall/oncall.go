// SPDX-License-Identifier: Apache-2.0

package oncall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/google/uuid"

	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
	githubclient "github.com/open-telemetry/sig-project-infra/otto/internal/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
	"github.com/open-telemetry/sig-project-infra/otto/internal/module"
	"github.com/open-telemetry/sig-project-infra/otto/internal/telemetry"
)

const (
	// ModuleName is the name of the oncall module.
	ModuleName = "oncall"

	// EscalationCheckInterval is the interval at which to check for unacknowledged escalations.
	EscalationCheckInterval = 5 * time.Minute

	// EscalationThreshold is the time after which an unacknowledged escalation should be escalated.
	EscalationThreshold = 24 * time.Hour
)

var (
	// Command patterns.
	ackPattern         = regexp.MustCompile(`^/ack\b`)
	escalatePattern    = regexp.MustCompile(`^/escalate\b`)
	resolvePattern     = regexp.MustCompile(`^/resolve\b`)
	addUserPattern     = regexp.MustCompile(`^/oncall\s+add\s+user\s+([a-zA-Z0-9_-]+)\s+(.+)$`)
	addRotationPattern = regexp.MustCompile(`^/oncall\s+add\s+rotation\s+(.+)$`)
	assignUserPattern  = regexp.MustCompile(`^/oncall\s+assign\s+([a-zA-Z0-9_-]+)\s+to\s+(.+)$`)
)

// Config contains configuration for the OnCall module.
type Config struct {
	// EnabledRepositories is the list of repositories for which this module is enabled.
	EnabledRepositories []string `json:"enabled_repositories"`
}

// Dependencies contains all the dependencies for the OnCall module.
type Dependencies struct {
	Logger     logging.Logger
	Telemetry  telemetry.Provider
	Database   database.Provider
	Repository database.Repository[database.AnyEntity]
	GitHub     githubclient.Provider
}

// Module implements the OnCall module.
type Module struct {
	*module.BaseModule
	Config Config
	Repo   OnCallRepository
	ticker *time.Ticker
	stopCh chan struct{}
	logger logging.Logger
	github githubclient.Provider
}

// Ensure Module implements module.Module.
var (
	_ module.Module      = (*Module)(nil)
	_ module.Initializer = (*Module)(nil)
	_ module.Shutdowner  = (*Module)(nil)
)

// New creates a new OnCall module.
func New(config Config, deps Dependencies) *Module {
	baseMod := module.NewBaseModule(ModuleName, module.Dependencies{
		Logger:     deps.Logger,
		Telemetry:  deps.Telemetry,
		Database:   deps.Database,
		Repository: deps.Repository,
		GitHub:     deps.GitHub,
	})

	logger := deps.Logger
	if logger == nil {
		// Use NoopLogger as fallback
		logger = logging.NewNoopLogger()
	}

	moduleLogger := logger.With("module", ModuleName)

	// Determine the repository to use
	var repo OnCallRepository
	var err error
	if deps.Database != nil {
		// Create repository with the provided database
		repo, err = NewSQLiteOnCallRepository(deps.Database)
		if err != nil {
			return nil
		}
	}

	return &Module{
		BaseModule: baseMod,
		Config:     config,
		Repo:       repo,
		stopCh:     make(chan struct{}),
		logger:     moduleLogger,
		github:     deps.GitHub,
	}
}

// Initialize sets up the OnCall module and starts background tasks.
func (o *Module) Initialize(ctx context.Context) error {
	o.logger.Info("Initializing oncall module")

	// Verify dependencies
	if o.Repo == nil {
		return errors.New("oncall repository is required")
	}

	if o.github == nil {
		return errors.New("github client is required")
	}

	// Initialize database tables
	if err := o.Repo.EnsureSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize oncall database: %w", err)
	}

	// Start background ticker for checking escalations
	o.ticker = time.NewTicker(EscalationCheckInterval)
	go o.runEscalationCheck()

	return nil
}

// Shutdown stops any background tasks.
func (o *Module) Shutdown(ctx context.Context) error {
	o.logger.Info("Shutting down oncall module")

	if o.ticker != nil {
		o.ticker.Stop()
	}

	close(o.stopCh)

	return nil
}

// HandleEvent processes events from GitHub webhooks.
func (o *Module) HandleEvent(eventType string, event interface{}, raw json.RawMessage) error {
	switch eventType {
	case "issues":
		return o.handleIssueEvent(event.(*github.IssuesEvent), raw)
	case "issue_comment":
		return o.handleIssueCommentEvent(event.(*github.IssueCommentEvent), raw)
	case "pull_request":
		return o.handlePullRequestEvent(event.(*github.PullRequestEvent), raw)
	case "pull_request_review_comment":
		return o.handlePullRequestReviewCommentEvent(event.(*github.PullRequestReviewCommentEvent), raw)
	}

	return nil
}

// handleIssueEvent processes GitHub issue events.
// The raw parameter is not used but required by the module interface.
func (o *Module) handleIssueEvent(event *github.IssuesEvent, _ json.RawMessage) error {
	if event.GetAction() != "opened" {
		return nil
	}

	// Check if repository is enabled
	if !o.isRepositoryEnabled(event.GetRepo().GetFullName()) {
		return nil
	}

	// Process new issue - check if it needs escalation
	// This would typically involve routing logic based on labels, repo, etc.
	// For now, we'll just log it
	o.logger.Info("New issue opened",
		"repo", event.GetRepo().GetFullName(),
		"number", event.GetIssue().GetNumber(),
		"title", event.GetIssue().GetTitle())

	return nil
}

// handleIssueCommentEvent processes GitHub issue comment events.
// The raw parameter is not used but required by the module interface.
func (o *Module) handleIssueCommentEvent(event *github.IssueCommentEvent, _ json.RawMessage) error {
	if event.GetAction() != "created" {
		return nil
	}

	// Check if repository is enabled
	if !o.isRepositoryEnabled(event.GetRepo().GetFullName()) {
		return nil
	}

	commentBody := event.GetComment().GetBody()
	issueNumber := event.GetIssue().GetNumber()
	repo := event.GetRepo().GetFullName()
	commenter := event.GetComment().GetUser().GetLogin()

	// Process commands in comment
	if ackPattern.MatchString(commentBody) {
		return o.handleAckCommand(repo, issueNumber, commenter)
	} else if escalatePattern.MatchString(commentBody) {
		return o.handleEscalateCommand(repo, issueNumber, commenter)
	} else if resolvePattern.MatchString(commentBody) {
		return o.handleResolveCommand(repo, issueNumber, commenter)
	} else if match := addUserPattern.FindStringSubmatch(commentBody); len(match) > 0 {
		return o.handleAddUserCommand(repo, issueNumber, commenter, match[1], match[2])
	} else if match := addRotationPattern.FindStringSubmatch(commentBody); len(match) > 0 {
		return o.handleAddRotationCommand(repo, issueNumber, commenter, match[1])
	} else if match := assignUserPattern.FindStringSubmatch(commentBody); len(match) > 0 {
		return o.handleAssignUserCommand(repo, issueNumber, commenter, match[1], match[2])
	}

	return nil
}

// handlePullRequestEvent processes GitHub pull request events.
// The raw parameter is not used but required by the module interface.
func (o *Module) handlePullRequestEvent(event *github.PullRequestEvent, _ json.RawMessage) error {
	if event.GetAction() != "opened" {
		return nil
	}

	// Check if repository is enabled
	if !o.isRepositoryEnabled(event.GetRepo().GetFullName()) {
		return nil
	}

	// Process new PR - check if it needs escalation
	o.logger.Info("New pull request opened",
		"repo", event.GetRepo().GetFullName(),
		"number", event.GetPullRequest().GetNumber(),
		"title", event.GetPullRequest().GetTitle())

	return nil
}

// handlePullRequestReviewCommentEvent processes GitHub PR review comment events.
// The raw parameter is not used but required by the module interface.
func (o *Module) handlePullRequestReviewCommentEvent(
	event *github.PullRequestReviewCommentEvent,
	_ json.RawMessage,
) error {
	if event.GetAction() != "created" {
		return nil
	}

	// Check if repository is enabled
	if !o.isRepositoryEnabled(event.GetRepo().GetFullName()) {
		return nil
	}

	commentBody := event.GetComment().GetBody()
	prNumber := event.GetPullRequest().GetNumber()
	repo := event.GetRepo().GetFullName()
	commenter := event.GetComment().GetUser().GetLogin()

	// Process commands in comment
	if ackPattern.MatchString(commentBody) {
		return o.handleAckPRCommand(repo, prNumber, commenter)
	} else if escalatePattern.MatchString(commentBody) {
		return o.handleEscalatePRCommand(repo, prNumber, commenter)
	} else if resolvePattern.MatchString(commentBody) {
		return o.handleResolvePRCommand(repo, prNumber, commenter)
	}

	return nil
}

// isRepositoryEnabled checks if a repository is enabled for this module.
func (o *Module) isRepositoryEnabled(repo string) bool {
	if len(o.Config.EnabledRepositories) == 0 {
		// If no repositories are specified, all are enabled
		return true
	}

	for _, r := range o.Config.EnabledRepositories {
		if r == repo {
			return true
		}
	}

	return false
}

// Command handlers

// handleAckCommand processes an acknowledgment command for an issue.
func (o *Module) handleAckCommand(repo string, issueNumber int, commenter string) error {
	ctx := context.Background()

	// Find the escalation for this issue
	escalation, err := o.Repo.FindEscalationByIssue(ctx, repo, issueNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - create one
		user, err := o.Repo.FindUserByGitHubUsername(ctx, commenter)
		if err != nil {
			return fmt.Errorf("error finding user: %w", err)
		}

		if user == nil {
			// Auto-register user
			user = &OnCallUser{
				ID:             uuid.New().String(),
				GitHubUsername: commenter,
				Name:           commenter,
				IsActive:       true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := o.Repo.CreateUser(ctx, user); err != nil {
				return fmt.Errorf("error creating user: %w", err)
			}
		}

		// Find rotation for this repository
		rotations, err := o.Repo.FindRotationsByRepository(ctx, repo)
		if err != nil {
			return fmt.Errorf("error finding rotations: %w", err)
		}

		var rotationID string
		if len(rotations) == 0 {
			// Create a default rotation
			rotation := &OnCallRotation{
				ID:          uuid.New().String(),
				Name:        repo + " Default Rotation",
				Description: "Default rotation created automatically",
				Repository:  repo,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := o.Repo.CreateRotation(ctx, rotation); err != nil {
				return fmt.Errorf("error creating rotation: %w", err)
			}
			rotationID = rotation.ID
		} else {
			rotationID = rotations[0].ID
		}

		// Create assignment
		assignment := &OnCallAssignment{
			ID:         uuid.New().String(),
			RotationID: rotationID,
			UserID:     user.ID,
			StartTime:  time.Now(),
			IsCurrent:  true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := o.Repo.CreateAssignment(ctx, assignment); err != nil {
			return fmt.Errorf("error creating assignment: %w", err)
		}

		// Create escalation
		escalation = &OnCallEscalation{
			ID:           uuid.New().String(),
			AssignmentID: assignment.ID,
			IssueNumber:  issueNumber,
			Repository:   repo,
			Status:       StatusAcknowledged,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := o.Repo.CreateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error creating escalation: %w", err)
		}

		// Comment on the issue
		comment := fmt.Sprintf("@%s has acknowledged this issue.", commenter)
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}

		return nil
	}

	// Escalation exists - update it
	if escalation.Status == StatusPending || escalation.Status == StatusEscalated {
		escalation.Status = StatusAcknowledged
		escalation.UpdatedAt = time.Now()
		if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error updating escalation: %w", err)
		}

		// Comment on the issue
		comment := fmt.Sprintf("@%s has acknowledged this issue.", commenter)
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}
	}

	return nil
}

// handleEscalateCommand processes an escalation command for an issue.
// The commenter parameter is currently unused but may be used for permissions checks in the future.
func (o *Module) handleEscalateCommand(repo string, issueNumber int, _ string) error {
	ctx := context.Background()

	// Find the escalation for this issue
	escalation, err := o.Repo.FindEscalationByIssue(ctx, repo, issueNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - create one with status escalated
		escalation = &OnCallEscalation{
			ID:             uuid.New().String(),
			IssueNumber:    issueNumber,
			Repository:     repo,
			Status:         StatusEscalated,
			EscalationTime: time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := o.Repo.CreateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error creating escalation: %w", err)
		}

		// Try to find the current on-call user for this repository
		rotations, err := o.Repo.FindRotationsByRepository(ctx, repo)
		if err != nil {
			return fmt.Errorf("error finding rotations: %w", err)
		}

		if len(rotations) > 0 {
			assignment, err := o.Repo.FindCurrentAssignmentByRotation(ctx, rotations[0].ID)
			if err != nil {
				return fmt.Errorf("error finding current assignment: %w", err)
			}

			if assignment != nil {
				// Update the escalation with the assignment
				escalation.AssignmentID = assignment.ID
				if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
					return fmt.Errorf("error updating escalation: %w", err)
				}

				// Get the user
				user, err := o.Repo.FindUserByID(ctx, assignment.UserID)
				if err != nil {
					return fmt.Errorf("error finding user: %w", err)
				}

				if user != nil {
					// Comment on the issue
					comment := fmt.Sprintf("This issue has been escalated to @%s.", user.GitHubUsername)
					if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
						return fmt.Errorf("error posting comment: %w", err)
					}

					return nil
				}
			}
		}

		// No rotation or assignment found - generic comment
		comment := "This issue has been marked for escalation."
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}

		return nil
	}

	// Escalation exists - update it
	escalation.Status = StatusEscalated
	escalation.EscalationTime = time.Now()
	escalation.UpdatedAt = time.Now()
	if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
		return fmt.Errorf("error updating escalation: %w", err)
	}

	// Comment on the issue
	comment := "This issue has been re-escalated."
	if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// handleResolveCommand processes a resolve command for an issue.
func (o *Module) handleResolveCommand(repo string, issueNumber int, commenter string) error {
	ctx := context.Background()

	// Find the escalation for this issue
	escalation, err := o.Repo.FindEscalationByIssue(ctx, repo, issueNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - nothing to do
		return nil
	}

	// Update the escalation
	escalation.Status = StatusResolved
	escalation.ResolutionTime = time.Now()
	escalation.UpdatedAt = time.Now()
	if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
		return fmt.Errorf("error updating escalation: %w", err)
	}

	// Comment on the issue
	comment := fmt.Sprintf("This issue has been marked as resolved by @%s.", commenter)
	if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// handleAddUserCommand processes a command to add a new user.
// The commenter parameter is currently unused but may be used for permissions checks in the future.
func (o *Module) handleAddUserCommand(repo string, issueNumber int, _ string, username, name string) error {
	ctx := context.Background()

	// Check if user already exists
	existingUser, err := o.Repo.FindUserByGitHubUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("error finding user: %w", err)
	}

	if existingUser != nil {
		comment := fmt.Sprintf("User @%s already exists.", username)
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}
		return nil
	}

	// Create new user
	user := &OnCallUser{
		ID:             uuid.New().String(),
		GitHubUsername: username,
		Name:           name,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := o.Repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	// Comment on the issue
	comment := fmt.Sprintf("User @%s has been added to the on-call system.", username)
	if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// handleAddRotationCommand processes a command to add a new rotation.
func (o *Module) handleAddRotationCommand(repo string, issueNumber int, commenter, name string) error {
	ctx := context.Background()

	// Create new rotation
	rotation := &OnCallRotation{
		ID:          uuid.New().String(),
		Name:        name,
		Description: "Rotation created by @" + commenter,
		Repository:  repo,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := o.Repo.CreateRotation(ctx, rotation); err != nil {
		return fmt.Errorf("error creating rotation: %w", err)
	}

	// Comment on the issue
	comment := fmt.Sprintf("On-call rotation '%s' has been created for this repository.", name)
	if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// handleAssignUserCommand processes a command to assign a user to a rotation.
// The commenter parameter is currently unused but may be used for permissions checks in the future.
func (o *Module) handleAssignUserCommand(repo string, issueNumber int, _ string, username, rotationName string) error {
	ctx := context.Background()

	// Find the user
	user, err := o.Repo.FindUserByGitHubUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("error finding user: %w", err)
	}

	if user == nil {
		comment := fmt.Sprintf("User @%s does not exist. Please add the user first.", username)
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}
		return nil
	}

	// Find rotations for this repository
	rotations, err := o.Repo.FindRotationsByRepository(ctx, repo)
	if err != nil {
		return fmt.Errorf("error finding rotations: %w", err)
	}

	// Find the rotation by name
	var rotation *OnCallRotation
	for _, r := range rotations {
		if strings.EqualFold(r.Name, rotationName) {
			rotation = &r
			break
		}
	}

	if rotation == nil {
		comment := fmt.Sprintf("Rotation '%s' does not exist. Please create it first.", rotationName)
		if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}
		return nil
	}

	// Create assignment
	assignment := &OnCallAssignment{
		ID:         uuid.New().String(),
		RotationID: rotation.ID,
		UserID:     user.ID,
		StartTime:  time.Now(),
		IsCurrent:  true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := o.Repo.CreateAssignment(ctx, assignment); err != nil {
		return fmt.Errorf("error creating assignment: %w", err)
	}

	// Comment on the issue
	comment := fmt.Sprintf("@%s has been assigned to the '%s' on-call rotation.", username, rotation.Name)
	if err := o.postIssueComment(ctx, repo, issueNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// PR command handlers

// handleAckPRCommand processes an acknowledgment command for a PR.
func (o *Module) handleAckPRCommand(repo string, prNumber int, commenter string) error {
	// Similar to handleAckCommand but for PRs
	ctx := context.Background()

	// Find the escalation for this PR
	escalation, err := o.Repo.FindEscalationByPR(ctx, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - create one
		user, err := o.Repo.FindUserByGitHubUsername(ctx, commenter)
		if err != nil {
			return fmt.Errorf("error finding user: %w", err)
		}

		if user == nil {
			// Auto-register user
			user = &OnCallUser{
				ID:             uuid.New().String(),
				GitHubUsername: commenter,
				Name:           commenter,
				IsActive:       true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := o.Repo.CreateUser(ctx, user); err != nil {
				return fmt.Errorf("error creating user: %w", err)
			}
		}

		// Find rotation for this repository
		rotations, err := o.Repo.FindRotationsByRepository(ctx, repo)
		if err != nil {
			return fmt.Errorf("error finding rotations: %w", err)
		}

		var rotationID string
		if len(rotations) == 0 {
			// Create a default rotation
			rotation := &OnCallRotation{
				ID:          uuid.New().String(),
				Name:        repo + " Default Rotation",
				Description: "Default rotation created automatically",
				Repository:  repo,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := o.Repo.CreateRotation(ctx, rotation); err != nil {
				return fmt.Errorf("error creating rotation: %w", err)
			}
			rotationID = rotation.ID
		} else {
			rotationID = rotations[0].ID
		}

		// Create assignment
		assignment := &OnCallAssignment{
			ID:         uuid.New().String(),
			RotationID: rotationID,
			UserID:     user.ID,
			StartTime:  time.Now(),
			IsCurrent:  true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := o.Repo.CreateAssignment(ctx, assignment); err != nil {
			return fmt.Errorf("error creating assignment: %w", err)
		}

		// Create escalation
		escalation = &OnCallEscalation{
			ID:           uuid.New().String(),
			AssignmentID: assignment.ID,
			PRNumber:     prNumber,
			Repository:   repo,
			Status:       StatusAcknowledged,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := o.Repo.CreateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error creating escalation: %w", err)
		}

		// Comment on the PR
		comment := fmt.Sprintf("@%s has acknowledged this pull request.", commenter)
		if err := o.postPRComment(ctx, repo, prNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}

		return nil
	}

	// Escalation exists - update it
	if escalation.Status == StatusPending || escalation.Status == StatusEscalated {
		escalation.Status = StatusAcknowledged
		escalation.UpdatedAt = time.Now()
		if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error updating escalation: %w", err)
		}

		// Comment on the PR
		comment := fmt.Sprintf("@%s has acknowledged this pull request.", commenter)
		if err := o.postPRComment(ctx, repo, prNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}
	}

	return nil
}

// handleEscalatePRCommand processes an escalation command for a PR.
// The commenter parameter is currently unused but may be used for permissions checks in the future.
func (o *Module) handleEscalatePRCommand(repo string, prNumber int, _ string) error {
	// Similar to handleEscalateCommand but for PRs
	ctx := context.Background()

	// Find the escalation for this PR
	escalation, err := o.Repo.FindEscalationByPR(ctx, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - create one with status escalated
		escalation = &OnCallEscalation{
			ID:             uuid.New().String(),
			PRNumber:       prNumber,
			Repository:     repo,
			Status:         StatusEscalated,
			EscalationTime: time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := o.Repo.CreateEscalation(ctx, escalation); err != nil {
			return fmt.Errorf("error creating escalation: %w", err)
		}

		// Similar logic as in handleEscalateCommand to find on-call user
		// ...

		comment := "This pull request has been marked for escalation."
		if err := o.postPRComment(ctx, repo, prNumber, comment); err != nil {
			return fmt.Errorf("error posting comment: %w", err)
		}

		return nil
	}

	// Escalation exists - update it
	escalation.Status = StatusEscalated
	escalation.EscalationTime = time.Now()
	escalation.UpdatedAt = time.Now()
	if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
		return fmt.Errorf("error updating escalation: %w", err)
	}

	// Comment on the PR
	comment := "This pull request has been re-escalated."
	if err := o.postPRComment(ctx, repo, prNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// handleResolvePRCommand processes a resolve command for a PR.
func (o *Module) handleResolvePRCommand(repo string, prNumber int, commenter string) error {
	// Similar to handleResolveCommand but for PRs
	ctx := context.Background()

	// Find the escalation for this PR
	escalation, err := o.Repo.FindEscalationByPR(ctx, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error finding escalation: %w", err)
	}

	if escalation == nil {
		// No escalation exists - nothing to do
		return nil
	}

	// Update the escalation
	escalation.Status = StatusResolved
	escalation.ResolutionTime = time.Now()
	escalation.UpdatedAt = time.Now()
	if err := o.Repo.UpdateEscalation(ctx, escalation); err != nil {
		return fmt.Errorf("error updating escalation: %w", err)
	}

	// Comment on the PR
	comment := fmt.Sprintf("This pull request has been marked as resolved by @%s.", commenter)
	if err := o.postPRComment(ctx, repo, prNumber, comment); err != nil {
		return fmt.Errorf("error posting comment: %w", err)
	}

	return nil
}

// Helper methods

// postIssueComment posts a comment on an issue.
func (o *Module) postIssueComment(ctx context.Context, repo string, issueNumber int, comment string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Create the comment object
	issueComment := &github.IssueComment{
		Body: github.Ptr(comment),
	}

	_, err := o.github.CreateIssueComment(ctx, owner, repoName, issueNumber, issueComment)
	return err
}

// postPRComment posts a comment on a pull request.
func (o *Module) postPRComment(ctx context.Context, repo string, prNumber int, comment string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Create the comment object
	issueComment := &github.IssueComment{
		Body: github.Ptr(comment),
	}

	// PR comments are technically issue comments in GitHub's API
	_, err := o.github.CreateIssueComment(ctx, owner, repoName, prNumber, issueComment)
	return err
}

// runEscalationCheck periodically checks for unacknowledged escalations.
func (o *Module) runEscalationCheck() {
	for {
		select {
		case <-o.stopCh:
			return
		case <-o.ticker.C:
			ctx := context.Background()
			o.checkPendingEscalations(ctx)
		}
	}
}

// checkPendingEscalations looks for pending escalations that need attention.
func (o *Module) checkPendingEscalations(ctx context.Context) {
	// Find pending escalations
	pendingEscalations, err := o.Repo.FindEscalationsByStatus(ctx, StatusPending)
	if err != nil {
		o.logger.Error("Error finding pending escalations", "error", err)
		return
	}

	now := time.Now()
	for _, escalation := range pendingEscalations {
		// Skip if not old enough to escalate
		if escalation.CreatedAt.Add(EscalationThreshold).After(now) {
			continue
		}

		// Escalate this issue/PR
		escalation.Status = StatusEscalated
		escalation.EscalationTime = now
		escalation.UpdatedAt = now
		if err := o.Repo.UpdateEscalation(ctx, &escalation); err != nil {
			o.logger.Error("Error updating escalation", "error", err, "id", escalation.ID)
			continue
		}

		// Post a comment based on whether it's an issue or PR
		var comment string
		if escalation.IssueNumber != 0 {
			comment = "This issue has been automatically escalated due to lack of acknowledgment."
			if err := o.postIssueComment(ctx, escalation.Repository, escalation.IssueNumber, comment); err != nil {
				o.logger.Error(
					"Error posting issue comment",
					"error",
					err,
					"repo",
					escalation.Repository,
					"issue",
					escalation.IssueNumber,
				)
			}
		} else if escalation.PRNumber != 0 {
			comment = "This pull request has been automatically escalated due to lack of acknowledgment."
			if err := o.postPRComment(ctx, escalation.Repository, escalation.PRNumber, comment); err != nil {
				o.logger.Error("Error posting PR comment", "error", err, "repo", escalation.Repository, "pr", escalation.PRNumber)
			}
		}
	}
}
