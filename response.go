package monime

import "net/http"

// Response contains HTTP metadata for a completed API operation.
type Response struct {
	StatusCode int
	RequestID  string
	Header     http.Header
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

func newResponse(statusCode int, header http.Header) Response {
	return Response{
		StatusCode: statusCode,
		RequestID:  header.Get("Monime-Request-Id"),
		Header:     header.Clone(),
	}
}

type apiEnvelope struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
}

type paginationEnvelope struct {
	Count int    `json:"count"`
	Next  string `json:"next"`
}

type apiResponse struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Result   []byte   `json:"result"`
}

type apiListResponse struct {
	Success    bool     `json:"success"`
	Messages   []string `json:"messages"`
	Result     []byte   `json:"result"`
	Pagination []byte   `json:"pagination"`
}

type apiDeleteResponse struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
}
