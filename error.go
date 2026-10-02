package monime

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Stable sentinel classifications for conditions callers branch on.
var (
	ErrUnauthorized = errors.New("monime: unauthorized")
	ErrForbidden    = errors.New("monime: forbidden")
	ErrNotFound     = errors.New("monime: not found")
	ErrConflict     = errors.New("monime: conflict")
	ErrRateLimited  = errors.New("monime: rate limited")
)

// APIError is returned for a non-2xx response or invalid success envelope.
type APIError struct {
	Status         int
	Code           int
	Reason         string
	Message        string
	Details        json.RawMessage
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

// IsRetryable reports whether the request that produced the error is eligible for retry.
func (e *APIError) IsRetryable() bool { return isRetryable(e) }

// ValidationIssue identifies one invalid field.
type ValidationIssue struct {
	Field   string
	Message string
}

// ValidationError reports invalid client configuration or request input.
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "invalid monime input"
	}
	return fmt.Sprintf("invalid monime input: %s: %s", e.Issues[0].Field, e.Issues[0].Message)
}

func newValidationError(field, message string) *ValidationError {
	return &ValidationError{Issues: []ValidationIssue{{Field: field, Message: message}}}
}
