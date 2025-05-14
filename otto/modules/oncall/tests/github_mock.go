// SPDX-License-Identifier: Apache-2.0

// Package tests provides mock implementations for testing the oncall module.
package tests

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v71/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"
)

// MockGitHubProvider mocks the GitHub client for testing.
type MockGitHubProvider struct {
	Comments map[string][]string // repo/123 -> []comments
}

func NewMockGitHubProvider() *MockGitHubProvider {
	return &MockGitHubProvider{
		Comments: make(map[string][]string),
	}
}

func (m *MockGitHubProvider) GetClient() *github.Client {
	return nil
}

func (m *MockGitHubProvider) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	return nil, nil
}

func (m *MockGitHubProvider) CreateIssue(
	ctx context.Context,
	owner, repo string,
	issue *github.IssueRequest,
) (*github.Issue, error) {
	return nil, nil
}

func (m *MockGitHubProvider) CreateIssueComment(
	ctx context.Context,
	owner, repo string,
	number int,
	comment *github.IssueComment,
) (*github.IssueComment, error) {
	key := owner + "/" + repo + "/" + strconv.Itoa(number)
	m.Comments[key] = append(m.Comments[key], *comment.Body)
	return nil, nil
}

func (m *MockGitHubProvider) GetPullRequest(
	ctx context.Context,
	owner, repo string,
	number int,
) (*github.PullRequest, error) {
	return nil, nil
}

func (m *MockGitHubProvider) CreatePullRequest(
	ctx context.Context,
	owner, repo string,
	pull *github.NewPullRequest,
) (*github.PullRequest, error) {
	return nil, nil
}

// MockLogger for tests.
type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) Debug(msg string, args ...any) {}
func (l *MockLogger) Info(msg string, args ...any)  {}
func (l *MockLogger) Warn(msg string, args ...any)  {}
func (l *MockLogger) Error(msg string, args ...any) {
	fmt.Printf("ERROR: %s %v\n", msg, args)
}

func (l *MockLogger) With(args ...any) logging.Logger {
	return l
}
