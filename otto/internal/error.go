// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"log/slog"
)

// ErrorType represents categories of application errors.
type ErrorType string

const (
	// ErrorTypeConfig represents configuration errors.
	ErrorTypeConfig ErrorType = "config"
	// ErrorTypeDatabase represents database errors.
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeServer represents HTTP server errors.
	ErrorTypeServer ErrorType = "server"
	// ErrorTypeModule represents module-related errors.
	ErrorTypeModule ErrorType = "module"
	// ErrorTypeCommand represents command processing errors.
	ErrorTypeCommand ErrorType = "command"
	// ErrorTypeGeneral represents general errors.
	ErrorTypeGeneral ErrorType = "general"
)

// AppError represents an application-level error with context.
type AppError struct {
	Type    ErrorType      // Category of error
	Op      string         // Operation that failed
	Err     error          // Original error
	Details map[string]any // Additional context
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s operation failed", e.Type, e.Op)
	}
	return fmt.Sprintf("%s: %s: %v", e.Type, e.Op, e.Err)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// LogAndWrapError logs an error with context and returns a wrapped error.
func LogAndWrapError(err error, errType ErrorType, op string, details map[string]any) error {
	// Don't wrap nil errors
	if err == nil {
		return nil
	}

	// If it's already an AppError, just add details
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Add any new details
		for k, v := range details {
			if _, exists := appErr.Details[k]; !exists {
				appErr.Details[k] = v
			}
		}

		// Log at appropriate level
		logAppError(appErr)
		return appErr
	}

	// Create details map if nil
	if details == nil {
		details = make(map[string]any)
	}

	// Create a new AppError
	appErr = &AppError{
		Type:    errType,
		Op:      op,
		Err:     err,
		Details: details,
	}

	// Log at appropriate level
	logAppError(appErr)
	return appErr
}

// logAppError logs an AppError at an appropriate level based on type.
func logAppError(err *AppError) {
	// Extract log attributes from details
	attrs := make([]any, 0, len(err.Details)*2+4)
	attrs = append(attrs, "error_type", err.Type, "operation", err.Op)

	for k, v := range err.Details {
		attrs = append(attrs, k, v)
	}

	// Determine appropriate log level based on error type
	switch err.Type {
	case ErrorTypeConfig, ErrorTypeDatabase, ErrorTypeServer:
		// These errors are often fatal or require admin attention
		slog.Error(err.Error(), attrs...)
	case ErrorTypeModule, ErrorTypeCommand:
		// Module and command errors might be due to user input
		slog.Warn(err.Error(), attrs...)
	default:
		// Default to info level for unknown error types
		slog.Info(err.Error(), attrs...)
	}
}

// IsErrorType checks if an error or any wrapped error is an AppError of specific type.
func IsErrorType(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// GetErrorDetails extracts the details map from an AppError.
// Returns nil if the error is not an AppError.
func GetErrorDetails(err error) map[string]any {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Details
	}
	return nil
}
