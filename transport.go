package monime

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBody = 10 << 20

type operation struct {
	method         string
	path           string
	query          url.Values
	input          any
	output         any
	idempotencyKey string
	list           bool
	emptyOK        bool
}

func (c *Client) do(ctx context.Context, op operation) (Response, error) {
	response, _, err := c.doRequest(ctx, op)
	return response, err
}

func doList[T any](
	c *Client,
	ctx context.Context,
	path string,
	query url.Values,
) (Page[T], Response, error) {
	items := []T{}
	response, pagination, err := c.doRequest(ctx, operation{
		method: http.MethodGet,
		path:   path,
		query:  query,
		output: &items,
		list:   true,
	})
	if err != nil {
		return Page[T]{}, response, err
	}
	return Page[T]{
		Items: items,
		PageInfo: PageInfo{
			Count: pagination.Count,
			Next:  pagination.Next,
		},
	}, response, nil
}

func (c *Client) doRequest(
	ctx context.Context,
	op operation,
) (Response, *paginationEnvelope, error) {
	if err := validateOperation(ctx, op); err != nil {
		return Response{}, nil, err
	}

	operationCtx := ctx
	cancel := func() {}
	if c.timeout > 0 {
		operationCtx, cancel = context.WithTimeout(ctx, c.timeout)
	}
	defer cancel()

	body, err := marshalBody(op.input)
	if err != nil {
		return Response{}, nil, err
	}
	requestURL := c.buildURL(op.path, op.query)
	headers := c.requestHeaders(body != nil, op.idempotencyKey)

	var lastResponse Response
	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		response, pagination, err := c.doAttempt(operationCtx, attemptRequest{
			method:  op.method,
			url:     requestURL,
			body:    body,
			headers: headers,
			output:  op.output,
			list:    op.list,
			emptyOK: op.emptyOK,
		})
		response.Attempts = attempt
		response.IdempotencyKey = op.idempotencyKey
		lastResponse = response
		if apiErr := apiErrorFrom(err); apiErr != nil {
			apiErr.Attempts = attempt
			apiErr.IdempotencyKey = op.idempotencyKey
		}
		if err == nil {
			return response, pagination, nil
		}
		if operationCtx.Err() != nil {
			return response, nil, operationContextError(ctx, c.timeout)
		}
		if attempt == c.retry.MaxAttempts || !isReplaySafe(op.method) || !isRetryable(err) {
			return response, nil, err
		}
		if err := sleep(operationCtx, c.retryDelay(attempt-1, err)); err != nil {
			return lastResponse, nil, operationContextError(ctx, c.timeout)
		}
	}

	return lastResponse, nil, errors.New("monime: request attempts exhausted")
}

func validateOperation(ctx context.Context, op operation) error {
	if ctx == nil {
		return newValidationError("Context", "must be non-nil")
	}
	if op.method == http.MethodPost {
		if op.idempotencyKey == "" {
			return newValidationError("IdempotencyKey", "must be non-empty for POST requests")
		}
		if len(op.idempotencyKey) > 64 {
			return newValidationError("IdempotencyKey", "must contain at most 64 characters")
		}
		return nil
	}
	if op.idempotencyKey != "" {
		return newValidationError("IdempotencyKey", "must be omitted for non-POST requests")
	}
	return nil
}

func operationContextError(parent context.Context, timeout time.Duration) error {
	if err := parent.Err(); err != nil {
		return err
	}
	return fmt.Errorf("monime: operation timed out after %s: %w", timeout, context.DeadlineExceeded)
}

func (c *Client) buildURL(path string, query url.Values) *url.URL {
	requestURL := *c.baseURL
	escapedPath := "/v1" + path
	decodedPath, err := url.PathUnescape(escapedPath)
	if err != nil {
		decodedPath = escapedPath
	}
	requestURL.Path = decodedPath
	requestURL.RawPath = escapedPath
	requestURL.RawQuery = query.Encode()
	return &requestURL
}

func marshalBody(input any) ([]byte, error) {
	if input == nil {
		return nil, nil
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("monime: encode request body: %w", err)
	}
	return body, nil
}

func (c *Client) requestHeaders(hasBody bool, idempotencyKey string) http.Header {
	headers := http.Header{
		"Authorization":   {"Bearer " + c.accessToken},
		"Monime-Space-Id": {c.spaceID},
		"Monime-Version":  {apiVersion},
		"Accept":          {"application/json"},
		"User-Agent":      {c.userAgent},
	}
	if hasBody {
		headers.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		headers.Set("Idempotency-Key", idempotencyKey)
	}
	return headers
}

type attemptRequest struct {
	method  string
	url     *url.URL
	body    []byte
	headers http.Header
	output  any
	list    bool
	emptyOK bool
}

func (c *Client) doAttempt(
	ctx context.Context,
	req attemptRequest,
) (Response, *paginationEnvelope, error) {
	request, err := http.NewRequestWithContext(ctx, req.method, req.url.String(), bytes.NewReader(req.body))
	if err != nil {
		return Response{}, nil, fmt.Errorf("monime: create request: %w", err)
	}
	request.Header = req.headers.Clone()

	httpResponse, err := c.httpClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return Response{}, nil, ctx.Err()
		}
		return Response{}, nil, sanitizeTransportError(err)
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(io.LimitReader(httpResponse.Body, maxResponseBody+1))
	if err != nil {
		return Response{}, nil, fmt.Errorf("monime: read response: %w", err)
	}
	retryAfter, _ := parseRetryAfter(httpResponse.Header.Get("Retry-After"), time.Now())
	response := Response{
		StatusCode: httpResponse.StatusCode,
		RequestID:  httpResponse.Header.Get("Monime-Request-Id"),
		RetryAfter: retryAfter,
	}
	if len(body) > maxResponseBody {
		return response, nil, &APIError{
			Status:     httpResponse.StatusCode,
			Code:       httpResponse.StatusCode,
			Reason:     "response_too_large",
			Message:    "response body exceeds 10 MiB",
			RequestID:  response.RequestID,
			RetryAfter: response.RetryAfter,
		}
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return response, nil, parseAPIError(httpResponse.StatusCode, httpResponse.Header, body)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		if req.emptyOK || req.output == nil {
			return response, nil, nil
		}
		return response, nil, invalidSuccess(httpResponse.StatusCode, "missing_result", body)
	}
	if req.output == nil {
		return response, nil, nil
	}
	if req.list {
		pagination, err := decodeList(body, req.output, httpResponse.StatusCode)
		return response, pagination, err
	}
	return response, nil, decodeResult(body, req.output, httpResponse.StatusCode)
}

func sanitizeTransportError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("monime: HTTP request failed: %w", urlErr.Err)
	}
	return fmt.Errorf("monime: HTTP request failed: %w", err)
}

func decodeList(body []byte, output any, status int) (*paginationEnvelope, error) {
	var envelope apiListResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, invalidSuccess(status, "invalid_json", body)
	}
	if !envelope.Success {
		return nil, invalidSuccess(status, "unsuccessful_response", body)
	}
	if isMissingResult(envelope.Result) {
		return nil, invalidSuccess(status, "missing_result", body)
	}
	if envelope.Pagination == nil {
		return nil, invalidSuccess(status, "invalid_pagination", body)
	}
	if err := json.Unmarshal(envelope.Result, output); err != nil {
		return nil, invalidSuccess(status, "invalid_result", body)
	}
	return envelope.Pagination, nil
}

func decodeResult(body []byte, output any, status int) error {
	var envelope apiResponse
	if err := json.Unmarshal(body, &envelope); err != nil {
		return invalidSuccess(status, "invalid_json", body)
	}
	if !envelope.Success {
		return invalidSuccess(status, "unsuccessful_response", body)
	}
	if isMissingResult(envelope.Result) {
		return invalidSuccess(status, "missing_result", body)
	}
	if err := json.Unmarshal(envelope.Result, output); err != nil {
		return invalidSuccess(status, "invalid_result", body)
	}
	return nil
}

func isMissingResult(result json.RawMessage) bool {
	trimmed := bytes.TrimSpace(result)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func invalidSuccess(status int, reason string, body []byte) *APIError {
	return &APIError{
		Status:  status,
		Code:    status,
		Reason:  reason,
		Message: "invalid success response from server",
		Body:    body,
	}
}

func parseAPIError(status int, headers http.Header, body []byte) *APIError {
	apiErr := &APIError{
		Status:    status,
		Code:      status,
		Reason:    "http_error",
		Message:   "api request failed",
		RequestID: headers.Get("Monime-Request-Id"),
		Body:      body,
	}
	apiErr.RetryAfter, _ = parseRetryAfter(headers.Get("Retry-After"), time.Now())

	var envelope struct {
		Error *struct {
			Code    int             `json:"code"`
			Reason  string          `json:"reason"`
			Message string          `json:"message"`
			Details json.RawMessage `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		apiErr.Reason = "invalid_json"
		apiErr.Message = "invalid json response from server"
		return apiErr
	}
	if envelope.Error == nil {
		return apiErr
	}
	if envelope.Error.Code != 0 {
		apiErr.Code = envelope.Error.Code
	}
	if envelope.Error.Reason != "" {
		apiErr.Reason = envelope.Error.Reason
	}
	if envelope.Error.Message != "" {
		apiErr.Message = envelope.Error.Message
	}
	apiErr.Details = envelope.Error.Details
	return apiErr
}

func apiErrorFrom(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

func isReplaySafe(method string) bool {
	return method == http.MethodGet || method == http.MethodPost
}

func isRetryable(err error) bool {
	if apiErr := apiErrorFrom(err); apiErr != nil {
		return isRetryableStatus(apiErr.Status)
	}
	var networkErr net.Error
	return errors.As(err, &networkErr) && (networkErr.Timeout() || networkErr.Temporary())
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

func (c *Client) retryDelay(retryIndex int, err error) time.Duration {
	if apiErr := apiErrorFrom(err); apiErr != nil && apiErr.RetryAfter > 0 {
		return apiErr.RetryAfter
	}
	delay := c.retry.BaseDelay
	for range retryIndex {
		if delay >= c.retry.MaxDelay/2 {
			delay = c.retry.MaxDelay
			break
		}
		delay *= 2
	}
	if delay > c.retry.MaxDelay {
		delay = c.retry.MaxDelay
	}
	if delay <= 0 {
		return 0
	}
	return mathrand.N(delay + 1)
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
		if seconds >= 0 && seconds <= int64((1<<63-1)/time.Second) {
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
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// NewIdempotencyKey returns a random RFC 4122 UUIDv4 suitable for a POST request.
func NewIdempotencyKey() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("monime: generate idempotency key: %w", err)
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
