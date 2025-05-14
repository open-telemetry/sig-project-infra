// SPDX-License-Identifier: Apache-2.0

// Package github provides GitHub API client interfaces and implementations.
package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v71/github"
	"github.com/jferrl/go-githubauth"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
	"github.com/open-telemetry/sig-project-infra/otto/internal/telemetry"
	"golang.org/x/oauth2"
)

// Provider defines the interface for GitHub API operations.
type Provider interface {
	// GetClient returns the underlying GitHub client.
	GetClient() *github.Client

	// Issues API
	GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error)
	CreateIssue(ctx context.Context, owner, repo string, issue *github.IssueRequest) (*github.Issue, error)
	CreateIssueComment(
		ctx context.Context,
		owner, repo string,
		number int,
		comment *github.IssueComment,
	) (*github.IssueComment, error)

	// Pull Requests API
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error)
	CreatePullRequest(ctx context.Context, owner, repo string, pull *github.NewPullRequest) (*github.PullRequest, error)
}

// Config contains configuration for GitHub client.
type Config struct {
	// AppID is the GitHub App ID
	AppID int64
	// InstallationID is the GitHub App Installation ID
	InstallationID int64
	// PrivateKey is the GitHub App private key
	PrivateKey string
}

// Client implements Provider using the go-github library.
type Client struct {
	client    *github.Client
	telemetry telemetry.Provider
	logger    logging.Logger
}

// Ensure Client implements Provider.
var _ Provider = (*Client)(nil)

// NewClient creates a new GitHub client.
func NewClient(ctx context.Context, config Config, telemetryProvider telemetry.Provider) (Provider, error) {
	logger := telemetryProvider.GetLogger().With("component", "github")
	var client *github.Client

	if config.AppID > 0 && config.InstallationID > 0 && config.PrivateKey != "" {
		// Use GitHub App authentication
		appTokenSource, err := githubauth.NewApplicationTokenSource(config.AppID, []byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub app token source: %w", err)
		}

		installationTokenSource := githubauth.NewInstallationTokenSource(config.InstallationID, appTokenSource)

		// Create an HTTP client that uses the installation token
		httpClient := oauth2.NewClient(ctx, installationTokenSource)

		// Create a new GitHub client with the custom HTTP client
		client = github.NewClient(httpClient)
		logger.Info("GitHub client initialized with App authentication",
			"app_id", config.AppID,
			"installation_id", config.InstallationID)
	} else {
		// If no authentication configured, use unauthenticated client
		client = github.NewClient(nil)
		logger.Info("GitHub client initialized (no auth)")
	}

	return &Client{
		client:    client,
		telemetry: telemetryProvider,
		logger:    logger,
	}, nil
}

// GetClient returns the underlying GitHub client.
func (c *Client) GetClient() *github.Client {
	return c.client
}

// GetIssue gets an issue from GitHub.
func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	ctx, span := c.telemetry.StartServerEventSpan(ctx, "github.get_issue")
	defer span.End()

	issue, _, err := c.client.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		c.logger.Error("Failed to get issue", "owner", owner, "repo", repo, "number", number, "err", err)
		return nil, err
	}

	return issue, nil
}

// CreateIssue creates a new issue on GitHub.
func (c *Client) CreateIssue(
	ctx context.Context,
	owner, repo string,
	issue *github.IssueRequest,
) (*github.Issue, error) {
	ctx, span := c.telemetry.StartServerEventSpan(ctx, "github.create_issue")
	defer span.End()

	createdIssue, _, err := c.client.Issues.Create(ctx, owner, repo, issue)
	if err != nil {
		c.logger.Error("Failed to create issue", "owner", owner, "repo", repo, "err", err)
		return nil, err
	}

	return createdIssue, nil
}

// CreateIssueComment creates a new comment on an issue.
func (c *Client) CreateIssueComment(
	ctx context.Context,
	owner, repo string,
	number int,
	comment *github.IssueComment,
) (*github.IssueComment, error) {
	ctx, span := c.telemetry.StartServerEventSpan(ctx, "github.create_issue_comment")
	defer span.End()

	createdComment, _, err := c.client.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		c.logger.Error("Failed to create issue comment", "owner", owner, "repo", repo, "number", number, "err", err)
		return nil, err
	}

	return createdComment, nil
}

// GetPullRequest gets a pull request from GitHub.
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	ctx, span := c.telemetry.StartServerEventSpan(ctx, "github.get_pull_request")
	defer span.End()

	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		c.logger.Error("Failed to get pull request", "owner", owner, "repo", repo, "number", number, "err", err)
		return nil, err
	}

	return pr, nil
}

// CreatePullRequest creates a new pull request on GitHub.
func (c *Client) CreatePullRequest(
	ctx context.Context,
	owner, repo string,
	pull *github.NewPullRequest,
) (*github.PullRequest, error) {
	ctx, span := c.telemetry.StartServerEventSpan(ctx, "github.create_pull_request")
	defer span.End()

	createdPR, _, err := c.client.PullRequests.Create(ctx, owner, repo, pull)
	if err != nil {
		c.logger.Error("Failed to create pull request", "owner", owner, "repo", repo, "err", err)
		return nil, err
	}

	return createdPR, nil
}
