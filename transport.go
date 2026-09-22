package monime

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	mrand "math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	maxResponseBody = 10 << 20
	maxRetryJitter  = 500 * time.Millisecond
)

// operation is one typed API call.
type operation struct {
	method string
	path   string
	query  url.Values
	input  any
	output any
	list   bool
}

func validateConfig(config Config) error {
	if config.SpaceID == "" {
		return newValidationError("SpaceID", "must be non-empty")
	}
	if config.AccessToken == "" {
		return newValidationError("AccessToken", "must be non-empty")
	}
	if config.APIVersion == "" {
		return newValidationError("APIVersion", "must be non-empty")
	}
	if config.Timeout < 0 {
		return newValidationError("Timeout", "must be non-negative")
	}
	if config.Retries < 0 {
		return newValidationError("Retries", "must be non-negative")
	}
	if config.RetryDelay < 0 {
		return newValidationError("RetryDelay", "must be non-negative")
	}
	if math.IsNaN(config.RetryBackoff) ||
		math.IsInf(config.RetryBackoff, 0) ||
		config.RetryBackoff < 0 {
		return newValidationError("RetryBackoff", "must be a finite non-negative number")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil || !isAPIURL(baseURL) {
		return newValidationError("BaseURL", "must be a valid HTTPS URL")
	}
	return nil
}

func isAPIURL(rawURL *url.URL) bool {
	httpsOnly := rawURL.Scheme == "https"
	hasHost := rawURL.Host != ""
	noUser := rawURL.User == nil
	noQuery := rawURL.RawQuery == ""
	noFragment := rawURL.Fragment == ""
	return httpsOnly && hasHost && noUser && noQuery && noFragment
}

// sameHostRedirectPolicy prevents credentials leaking to another host.
func sameHostRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	if req.URL.Host != via[0].URL.Host {
		return fmt.Errorf("monime: refusing cross-host redirect to %s", req.URL.Host)
	}
	if len(via) >= 10 {
		return fmt.Errorf("monime: too many redirects")
	}
	return nil
}

func (c *Client) buildURL(path string, query url.Values) *url.URL {
	requestURL := *c.baseURL
	baseEscapedPath := requestURL.EscapedPath()
	escapedPath := "/v1" + path
	decodedPath, _ := url.PathUnescape(escapedPath)
	requestURL.Path += decodedPath
	requestURL.RawPath = baseEscapedPath + escapedPath
	requestURL.RawQuery = query.Encode()
	return &requestURL
}

func marshalBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("monime: encode request body: %w", err)
	}
	return encoded, nil
}

func (c *Client) requestHeaders(method string, hasBody bool) (http.Header, error) {
	headers := http.Header{
		"Authorization":   {"Bearer " + c.accessToken},
		"Monime-Space-Id": {c.spaceID},
		"Monime-Version":  {c.apiVersion},
		"Accept":          {"application/json"},
		"User-Agent":      {c.userAgent},
	}
	if hasBody {
		headers.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost {
		key, err := newUUIDv4()
		if err != nil {
			return nil, fmt.Errorf("monime: generate idempotency key: %w", err)
		}
		headers.Set("Idempotency-Key", key)
	}
	return headers, nil
}

// do executes a single-resource or delete operation.
func (c *Client) do(ctx context.Context, op operation) (Response, error) {
	resp, _, err := c.doRequest(ctx, op)
	return resp, err
}

// doList executes a list operation and decodes the {result, pagination} envelope.
func doList[T any](
	c *Client,
	ctx context.Context,
	path string,
	query url.Values,
) (Page[T], Response, error) {
	var items []T
	op := operation{
		method: http.MethodGet,
		path:   path,
		query:  query,
		output: &items,
		list:   true,
	}
	resp, pageInfo, err := c.doRequest(ctx, op)
	if err != nil {
		return Page[T]{}, resp, err
	}
	page := Page[T]{Items: items}
	if pageInfo != nil {
		page.PageInfo = *pageInfo
	}
	return page, resp, nil
}

func (c *Client) doRequest(
	ctx context.Context,
	op operation,
) (Response, *PageInfo, error) {
	if ctx == nil {
		return Response{}, nil, errors.New("monime: nil context")
	}

	operationCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && c.timeout > 0 {
		operationCtx, cancel = context.WithTimeout(ctx, c.timeout)
	}
	defer cancel()

	body, err := marshalBody(op.input)
	if err != nil {
		return Response{}, nil, err
	}
	requestURL := c.buildURL(op.path, op.query)
	headers, err := c.requestHeaders(op.method, body != nil)
	if err != nil {
		return Response{}, nil, err
	}
	idempotencyKey := headers.Get("Idempotency-Key")

	var lastErr error
	for attempt := 0; ; attempt++ {
		resp, pageInfo, err := c.doAttempt(operationCtx, attemptRequest{
			method:  op.method,
			url:     requestURL,
			body:    body,
			headers: headers,
			output:  op.output,
			list:    op.list,
		})
		if err == nil {
			resp.Attempts = attempt + 1
			resp.IdempotencyKey = idempotencyKey
			return resp, pageInfo, nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok {
			apiErr.Attempts = attempt + 1
			apiErr.IdempotencyKey = idempotencyKey
		}
		if operationCtx.Err() != nil {
			if ctx.Err() != nil {
				return Response{}, nil, ctx.Err()
			}
			return Response{}, nil, &TimeoutError{Timeout: c.timeout}
		}
		if attempt >= c.retries || !isReplaySafe(op.method) || !isRetryable(err) {
			return Response{}, nil, err
		}
		delay := c.calculateRetryDelay(attempt, lastErr)
		if err := sleep(operationCtx, delay); err != nil {
			return Response{}, nil, err
		}
	}
}

// attemptRequest is one HTTP attempt within an operation's retry loop.
type attemptRequest struct {
	method  string
	url     *url.URL
	body    []byte
	headers http.Header
	output  any
	list    bool
}

func (c *Client) doAttempt(
	ctx context.Context,
	req attemptRequest,
) (Response, *PageInfo, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		req.method,
		req.url.String(),
		bytes.NewReader(req.body),
	)
	if err != nil {
		return Response{}, nil, fmt.Errorf("monime: create request: %w", err)
	}
	request.Header = req.headers.Clone()

	response, err := c.httpClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return Response{}, nil, ctx.Err()
		}
		return Response{}, nil, &NetworkError{Cause: err}
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBody))
	if err != nil {
		return Response{}, nil, &NetworkError{Cause: err}
	}
	retryAfter, _ := parseRetryAfter(response.Header.Get("Retry-After"), time.Now())
	metadata := Response{
		StatusCode:     response.StatusCode,
		RequestID:      response.Header.Get("Monime-Request-Id"),
		Attempts:       1,
		IdempotencyKey: req.headers.Get("Idempotency-Key"),
		RetryAfter:     retryAfter,
		Header:         response.Header.Clone(),
	}
	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		err := parseAPIError(response.StatusCode, response.Header, bodyBytes, time.Now())
		return metadata, nil, err
	}
	if len(bytes.TrimSpace(bodyBytes)) == 0 || req.output == nil {
		return metadata, nil, nil
	}
	if req.list {
		pageInfo, err := decodeList(bodyBytes, req.output, response.StatusCode)
		if err != nil {
			return metadata, nil, err
		}
		return metadata, pageInfo, nil
	}
	if err := decodeResult(bodyBytes, req.output, response.StatusCode); err != nil {
		return metadata, nil, err
	}
	return metadata, nil, nil
}

func decodeList(body []byte, output any, status int) (*PageInfo, error) {
	var envelope apiListResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, malformedSuccess(status, body)
	}
	if len(envelope.Result) != 0 && string(envelope.Result) != "null" {
		if err := json.Unmarshal(envelope.Result, output); err != nil {
			return nil, invalidResult(status, body)
		}
	}
	var pageInfo *PageInfo
	if len(envelope.Pagination) != 0 && string(envelope.Pagination) != "null" {
		var pagination paginationEnvelope
		if err := json.Unmarshal(envelope.Pagination, &pagination); err == nil {
			pageInfo = &PageInfo{Count: pagination.Count, Next: pagination.Next}
		}
	}
	return pageInfo, nil
}

func decodeResult(body []byte, output any, status int) error {
	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return malformedSuccess(status, body)
	}
	if len(envelope.Result) != 0 && string(envelope.Result) != "null" {
		if err := json.Unmarshal(envelope.Result, output); err != nil {
			return invalidResult(status, body)
		}
	}
	return nil
}

func malformedSuccess(status int, body []byte) *APIError {
	return &APIError{
		Status:  status,
		Code:    status,
		Reason:  "invalid_json",
		Message: "invalid json response from server",
		Body:    body,
	}
}

func invalidResult(status int, body []byte) *APIError {
	return &APIError{
		Status:  status,
		Code:    status,
		Reason:  "invalid_result",
		Message: "invalid result response from server",
		Body:    body,
	}
}

func parseAPIError(
	status int,
	headers http.Header,
	body []byte,
	now time.Time,
) *APIError {
	apiError := &APIError{
		Status:  status,
		Code:    status,
		Reason:  "http_error",
		Message: "api request failed",
	}
	apiError.RetryAfter, _ = parseRetryAfter(headers.Get("Retry-After"), now)
	apiError.RequestID = headers.Get("Monime-Request-Id")

	var envelope struct {
		Error *struct {
			Code    int             `json:"code"`
			Reason  string          `json:"reason"`
			Message string          `json:"message"`
			Details json.RawMessage `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		apiError.Reason = "invalid_json"
		apiError.Message = "invalid json response from server"
		return apiError
	}
	if envelope.Error == nil {
		return apiError
	}
	if envelope.Error.Code != 0 {
		apiError.Code = envelope.Error.Code
	}
	if envelope.Error.Reason != "" {
		apiError.Reason = envelope.Error.Reason
	}
	if envelope.Error.Message != "" {
		apiError.Message = envelope.Error.Message
	}
	apiError.Details = envelope.Error.Details
	return apiError
}

// isReplaySafe reports whether a method may be retried. Only GET and POST
// with a stable idempotency key are replay-safe; PATCH and DELETE have no
// documented replay guarantee.
func isReplaySafe(method string) bool {
	return method == http.MethodGet || method == http.MethodPost
}

func isRetryable(err error) bool {
	var apiError *APIError
	if errors.As(err, &apiError) {
		return isRetryableStatus(apiError.Status)
	}
	var networkError *NetworkError
	return errors.As(err, &networkError)
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *Client) calculateRetryDelay(retryIndex int, err error) time.Duration {
	var apiError *APIError
	if errors.As(err, &apiError) && apiError.RetryAfter > 0 {
		return apiError.RetryAfter
	}
	baseDelay := float64(c.retryDelay) * math.Pow(c.retryBackoff, float64(retryIndex))
	maxDelay := time.Duration(math.MaxInt64 - int64(maxRetryJitter))
	if baseDelay > float64(maxDelay) {
		baseDelay = float64(maxDelay)
	}
	jitter := time.Duration(mrand.IntN(int(maxRetryJitter / time.Millisecond)))
	delay := time.Duration(baseDelay) + jitter*time.Millisecond
	if delay < 0 {
		return maxRetryJitter
	}
	return delay
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
		if seconds >= 0 && seconds <= int64(math.MaxInt64/time.Second) {
			return time.Duration(seconds) * time.Second, true
		}
		return 0, false
	}
	if date, err := http.ParseTime(value); err == nil && date.After(now) {
		return date.Sub(now), true
	}
	return 0, false
}

func sleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func newUUIDv4() (string, error) {
	var value [16]byte
	if _, err := crand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80

	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], value[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], value[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], value[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], value[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], value[10:16])
	return string(encoded), nil
}
