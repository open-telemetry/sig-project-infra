// SPDX-License-Identifier: Apache-2.0

// Package logging provides interfaces and implementations for logging.
package logging

import (
	"log/slog"
)

// Logger defines the interface for application logging.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, args ...any)
	// Info logs an informational message.
	Info(msg string, args ...any)
	// Warn logs a warning message.
	Warn(msg string, args ...any)
	// Error logs an error message.
	Error(msg string, args ...any)
	// With returns a new logger with the given attributes.
	With(args ...any) Logger
}

// SlogLogger is an implementation of Logger using Go's standard log/slog package.
type SlogLogger struct {
	logger *slog.Logger
}

// Ensure SlogLogger implements Logger.
var _ Logger = (*SlogLogger)(nil)

// NewSlogLogger creates a new logger using slog.
func NewSlogLogger(logger *slog.Logger) Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogLogger{logger: logger}
}

// Debug logs a debug message.
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an informational message.
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With returns a new logger with the given attributes.
func (l *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{logger: l.logger.With(args...)}
}

// MockLogger is an implementation of Logger for testing.
type MockLogger struct {
	// Messages stores logged messages for testing.
	Messages []LogMessage
}

// LogMessage represents a logged message for testing.
type LogMessage struct {
	Level   slog.Level
	Message string
	Args    []any
}

// Ensure MockLogger implements Logger.
var _ Logger = (*MockLogger)(nil)

// NewMockLogger creates a new mock logger for testing.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: make([]LogMessage, 0),
	}
}

// Debug logs a debug message.
func (l *MockLogger) Debug(msg string, args ...any) {
	l.Messages = append(l.Messages, LogMessage{
		Level:   slog.LevelDebug,
		Message: msg,
		Args:    args,
	})
}

// Info logs an informational message.
func (l *MockLogger) Info(msg string, args ...any) {
	l.Messages = append(l.Messages, LogMessage{
		Level:   slog.LevelInfo,
		Message: msg,
		Args:    args,
	})
}

// Warn logs a warning message.
func (l *MockLogger) Warn(msg string, args ...any) {
	l.Messages = append(l.Messages, LogMessage{
		Level:   slog.LevelWarn,
		Message: msg,
		Args:    args,
	})
}

// Error logs an error message.
func (l *MockLogger) Error(msg string, args ...any) {
	l.Messages = append(l.Messages, LogMessage{
		Level:   slog.LevelError,
		Message: msg,
		Args:    args,
	})
}

// With returns a new logger with the given attributes.
func (l *MockLogger) With(args ...any) Logger {
	// For simplicity, return the same logger in tests
	return l
}

// Clear clears all logged messages.
func (l *MockLogger) Clear() {
	l.Messages = make([]LogMessage, 0)
}

// GetMessages returns all logged messages.
func (l *MockLogger) GetMessages() []LogMessage {
	return l.Messages
}

// GetMessagesByLevel returns all logged messages of a specific level.
func (l *MockLogger) GetMessagesByLevel(level slog.Level) []LogMessage {
	var filtered []LogMessage
	for _, msg := range l.Messages {
		if msg.Level == level {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// HasMessage checks if a specific message was logged.
func (l *MockLogger) HasMessage(level slog.Level, msg string) bool {
	for _, logMsg := range l.Messages {
		if logMsg.Level == level && logMsg.Message == msg {
			return true
		}
	}
	return false
}

// NoopLogger is a logger that does nothing.
type NoopLogger struct{}

// Ensure NoopLogger implements Logger.
var _ Logger = (*NoopLogger)(nil)

// NewNoopLogger creates a new noop logger.
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// Debug does nothing.
func (l *NoopLogger) Debug(msg string, args ...any) {}

// Info does nothing.
func (l *NoopLogger) Info(msg string, args ...any) {}

// Warn does nothing.
func (l *NoopLogger) Warn(msg string, args ...any) {}

// Error does nothing.
func (l *NoopLogger) Error(msg string, args ...any) {}

// With returns the same noop logger.
func (l *NoopLogger) With(args ...any) Logger {
	return l
}
