package monime

import (
	"encoding/json"
	"time"
)

// Response contains HTTP metadata for a completed API operation.
type Response struct {
	StatusCode     int
	RequestID      string
	Attempts       int
	IdempotencyKey string
	RetryAfter     time.Duration
}

// Page is a cursor-paginated API result.
type Page[T any] struct {
	Items    []T
	PageInfo PageInfo
}

// PageInfo describes the cursor state returned by a list operation.
type PageInfo struct {
	Count int
	Next  string
}

// HasNext reports whether the API returned a cursor for another page.
func (p PageInfo) HasNext() bool { return p.Next != "" }

type apiResponse struct {
	Success  bool              `json:"success"`
	Messages []json.RawMessage `json:"messages"`
	Result   json.RawMessage   `json:"result"`
}

type apiListResponse struct {
	Success    bool                `json:"success"`
	Messages   []json.RawMessage   `json:"messages"`
	Result     json.RawMessage     `json:"result"`
	Pagination *paginationEnvelope `json:"pagination"`
}

type paginationEnvelope struct {
	Count int    `json:"count"`
	Next  string `json:"next"`
}
