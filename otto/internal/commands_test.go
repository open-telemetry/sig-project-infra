// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"
)

func TestIsSlashCommand(t *testing.T) {
	tests := []struct {
		body     string
		expected bool
	}{
		{"/echo foo bar", true},
		{"not a command", false},
		{"//not a command", false},
		{"/   ", true}, // Still starts with a slash
		{"/echo", true},
		{"prefix text\n/echo hey\nmore", true},
	}

	for _, tt := range tests {
		got := IsSlashCommand(tt.body)
		if got != tt.expected {
			t.Errorf("IsSlashCommand(%q) = %v, want %v", tt.body, got, tt.expected)
		}
	}
}
