// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"sync"

	"github.com/google/go-github/v71/github"
)

// MockClient implements Provider for testing.
type MockClient struct {
	client *github.Client

	// Call tracking
	GetIssueCalls           []GetIssueCall
	CreateIssueCalls        []CreateIssueCall
	CreateIssueCommentCalls []CreateIssueCommentCall
	GetPullRequestCalls     []GetPullRequestCall
	CreatePullRequestCalls  []CreatePullRequestCall

	// Return values
	ReturnIssue             *github.Issue
	ReturnIssueError        error
	ReturnIssueComment      *github.IssueComment
	ReturnIssueCommentError error
	ReturnPullRequest       *github.PullRequest
	ReturnPullRequestError  error

	mu sync.RWMutex
}

// GetIssueCall tracks parameters for GetIssue calls.
type GetIssueCall struct {
	Owner  string
	Repo   string
	Number int
}

type CreateIssueCall struct {
	Owner string
	Repo  string
	Issue *github.IssueRequest
}

type CreateIssueCommentCall struct {
	Owner   string
	Repo    string
	Number  int
	Comment *github.IssueComment
}

type GetPullRequestCall struct {
	Owner  string
	Repo   string
	Number int
}

type CreatePullRequestCall struct {
	Owner string
	Repo  string
	Pull  *github.NewPullRequest
}

// Ensure MockClient implements Provider.
var _ Provider = (*MockClient)(nil)

// NewMockClient creates a new mock GitHub client.
func NewMockClient() *MockClient {
	return &MockClient{
		client:                  github.NewClient(nil),
		GetIssueCalls:           make([]GetIssueCall, 0),
		CreateIssueCalls:        make([]CreateIssueCall, 0),
		CreateIssueCommentCalls: make([]CreateIssueCommentCall, 0),
		GetPullRequestCalls:     make([]GetPullRequestCall, 0),
		CreatePullRequestCalls:  make([]CreatePullRequestCall, 0),
	}
}

// GetClient returns the underlying GitHub client.
func (c *MockClient) GetClient() *github.Client {
	return c.client
}

// GetIssue mocks getting an issue from GitHub.
func (c *MockClient) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.GetIssueCalls = append(c.GetIssueCalls, GetIssueCall{
		Owner:  owner,
		Repo:   repo,
		Number: number,
	})

	return c.ReturnIssue, c.ReturnIssueError
}

// CreateIssue mocks creating a new issue on GitHub.
func (c *MockClient) CreateIssue(
	ctx context.Context,
	owner, repo string,
	issue *github.IssueRequest,
) (*github.Issue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CreateIssueCalls = append(c.CreateIssueCalls, CreateIssueCall{
		Owner: owner,
		Repo:  repo,
		Issue: issue,
	})

	return c.ReturnIssue, c.ReturnIssueError
}

// CreateIssueComment mocks creating a new comment on an issue.
func (c *MockClient) CreateIssueComment(
	ctx context.Context,
	owner, repo string,
	number int,
	comment *github.IssueComment,
) (*github.IssueComment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CreateIssueCommentCalls = append(c.CreateIssueCommentCalls, CreateIssueCommentCall{
		Owner:   owner,
		Repo:    repo,
		Number:  number,
		Comment: comment,
	})

	return c.ReturnIssueComment, c.ReturnIssueCommentError
}

// GetPullRequest mocks getting a pull request from GitHub.
func (c *MockClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.GetPullRequestCalls = append(c.GetPullRequestCalls, GetPullRequestCall{
		Owner:  owner,
		Repo:   repo,
		Number: number,
	})

	return c.ReturnPullRequest, c.ReturnPullRequestError
}

// CreatePullRequest mocks creating a new pull request on GitHub.
func (c *MockClient) CreatePullRequest(
	ctx context.Context,
	owner, repo string,
	pull *github.NewPullRequest,
) (*github.PullRequest, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CreatePullRequestCalls = append(c.CreatePullRequestCalls, CreatePullRequestCall{
		Owner: owner,
		Repo:  repo,
		Pull:  pull,
	})

	return c.ReturnPullRequest, c.ReturnPullRequestError
}
