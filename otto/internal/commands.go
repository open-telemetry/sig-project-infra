// SPDX-License-Identifier: Apache-2.0

// commands.go provides utilities for parsing slash commands from comment text.
// Note: Direct command dispatch has been removed. Each module should parse and handle
// commands in their HandleEvent implementation.

package internal

import (
	"context"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IsSlashCommand checks if a comment body contains a slash command.
func IsSlashCommand(body string) bool {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "/") && !strings.HasPrefix(trimmed, "//") {
			return true
		}
	}
	return false
}

// LogSlashCommand logs information about a detected slash command for tracing purposes.
func LogSlashCommand(ctx context.Context, command string, args []string, issuer, repo string, issueNum int) {
	// Create a span for command logging
	// Intentionally ignoring returned context since we don't use it further in this method
	// but keeping this pattern for consistency with other tracing code and to make it easier
	// to use the context in the future if needed
	_, span := OttoTracer().Start(ctx, "otto.slash_command",
		trace.WithAttributes(
			attribute.String("repo", repo),
			attribute.Int("issue_num", issueNum),
			attribute.String("issuer", issuer),
			attribute.String("command.name", command),
			attribute.Int("command.args_count", len(args)),
		))
	defer span.End()

	slog.Debug("Slash command detected",
		"command", command,
		"args_count", len(args),
		"issuer", issuer,
		"repo", repo,
		"issue", issueNum)
}
