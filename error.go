package monime

import (
	"encoding/json"
	"fmt"
	"time"
)

// APIError is returned for a non-success API response.
type APIError struct {
	Status     int
	Code       int
	Reason     string
	Message    string
	RequestID  string
	Details    json.RawMessage
	RetryAfter time.Duration
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("monime api error: %s", e.Reason)
	}
	return fmt.Sprintf("monime api error: http %d", e.Status)
}

// TimeoutError is returned when the SDK operation deadline expires.
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("monime request timed out after %s", e.Timeout)
}

// NetworkError is returned for a transport failure that is not context cancellation.
type NetworkError struct {
	Cause error
}

func (e *NetworkError) Error() string { return "monime network request failed" }

// Unwrap returns the underlying transport error.
func (e *NetworkError) Unwrap() error { return e.Cause }

// ValidationIssue identifies an invalid SDK execution setting.
type ValidationIssue struct {
	Field   string
	Message string
}

// ValidationError is returned when configuration cannot safely execute requests.
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
	return &ValidationError{Issues: []ValidationIssue{{Field: field, Message: message}}}
}
