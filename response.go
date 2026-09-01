package monime

import "encoding/json"

// APIResponse is the API envelope for a single resource operation.
type APIResponse struct {
	Success  bool            `json:"success"`
	Messages []string        `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

// APIListResponse is the API envelope for a list operation.
type APIListResponse struct {
	Success    bool            `json:"success"`
	Messages   []string        `json:"messages"`
	Result     json.RawMessage `json:"result"`
	Pagination json.RawMessage `json:"pagination"`
}

// APIDeleteResponse is the API envelope for a delete operation.
type APIDeleteResponse struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
}
