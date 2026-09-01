package monime

import (
	"encoding/json"
	"fmt"
	"time"
)

// APIError is returned when the API responds with a non-success HTTP status.
type APIError struct {
	Status        int
	Reason        string
	Message       string
	Details       json.RawMessage
	RetryAfter    time.Duration
	Retryable     bool
	retryAfterSet bool
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return fmt.Sprintf("monime API error: HTTP %d", e.Status)
}

// TimeoutError is returned when an individual request attempt exceeds its timeout.
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("monime request timed out after %s", e.Timeout)
}

// NetworkError is returned for transport failures that are not context cancellation.
type NetworkError struct {
	err error
}

func (e *NetworkError) Error() string {
	return "monime network request failed"
}

// Unwrap returns the underlying transport error.
func (e *NetworkError) Unwrap() error {
	return e.err
}

// ValidationIssue identifies an invalid client configuration field.
type ValidationIssue struct {
	Field   string
	Message string
}

// ValidationError is returned only when Config cannot safely execute requests.
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "invalid monime client configuration"
	}

	return fmt.Sprintf("invalid monime client configuration: %s", e.Issues[0].Message)
}

func newConfigValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Issues: []ValidationIssue{{
			Field:   field,
			Message: message,
		}},
	}
}
