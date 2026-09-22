package monime

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Stable sentinel classifications for conditions callers branch on.
// Use errors.Is to match them against *APIError.
var (
	ErrUnauthorized = errors.New("monime: unauthorized")
	ErrForbidden    = errors.New("monime: forbidden")
	ErrNotFound     = errors.New("monime: not found")
	ErrConflict     = errors.New("monime: conflict")
	ErrRateLimited  = errors.New("monime: rate limited")
)

// APIError is returned for a non-2xx API response or a malformed success body.
type APIError struct {
	Status         int
	Code           int
	Reason         string
	Message        string
	Details        []byte
	RequestID      string
	RetryAfter     time.Duration
	Attempts       int
	IdempotencyKey string
	Body           []byte
}

func (e *APIError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("monime api error: %s", e.Reason)
	}
	return fmt.Sprintf("monime api error: http %d", e.Status)
}

// Is reports whether the error matches a sentinel classification.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.Status == http.StatusUnauthorized
	case ErrForbidden:
		return e.Status == http.StatusForbidden
	case ErrNotFound:
		return e.Status == http.StatusNotFound
	case ErrConflict:
		return e.Status == http.StatusConflict
	case ErrRateLimited:
		return e.Status == http.StatusTooManyRequests
	default:
		return false
	}
}

// IsRetryable reports whether the failed operation may be retried when it is
// replay-safe. Replay safety is decided by the transport, not here.
func (e *APIError) IsRetryable() bool {
	return isRetryableStatus(e.Status)
}

// TimeoutError is returned when the SDK operation deadline expires.
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("monime request timed out after %s", e.Timeout)
}

// NetworkError is returned for a transport failure that is not caller
// cancellation or deadline expiry.
type NetworkError struct {
	Cause error
}

func (e *NetworkError) Error() string { return "monime network request failed" }

// Unwrap returns the underlying transport error.
func (e *NetworkError) Unwrap() error { return e.Cause }

// ValidationIssue identifies one invalid field.
type ValidationIssue struct {
	Field   string
	Message string
}

// ValidationError is returned when client configuration or a request input
// cannot safely execute.
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "invalid monime client configuration"
	}
	return fmt.Sprintf("invalid monime client configuration: %s", e.Issues[0].Message)
}

func newValidationError(field, message string) *ValidationError {
	return &ValidationError{Issues: []ValidationIssue{{Field: field, Message: message}}}
}
