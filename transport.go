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

// RequestConfig overrides client settings for a single request.
type RequestConfig struct {
	Timeout        time.Duration
	Retries        int
	IdempotencyKey string
}

type requestOptions struct {
	Method string
	Path   string
	Body   any
	Query  url.Values
	Config *RequestConfig
}

func validateConfig(config Config) error {
	if config.SpaceID == "" {
		return newConfigValidationError("SpaceID", "must be non-empty")
	}
	if config.AccessToken == "" {
		return newConfigValidationError("AccessToken", "must be non-empty")
	}
	if config.APIVersion == "" {
		return newConfigValidationError("APIVersion", "must be non-empty")
	}
	if config.Timeout < 0 {
		return newConfigValidationError("Timeout", "must be non-negative")
	}
	if config.Retries < 0 {
		return newConfigValidationError("Retries", "must be non-negative")
	}
	if config.RetryDelay < 0 {
		return newConfigValidationError("RetryDelay", "must be non-negative")
	}
	if math.IsNaN(config.RetryBackoff) || math.IsInf(config.RetryBackoff, 0) || config.RetryBackoff < 0 {
		return newConfigValidationError("RetryBackoff", "must be a finite non-negative number")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil || baseURL.Scheme != "https" || baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return newConfigValidationError("BaseURL", "must be a valid HTTPS URL")
	}

	return nil
}

func validateRequestConfig(config *RequestConfig) error {
	if config == nil {
		return nil
	}
	if config.Timeout < 0 {
		return newConfigValidationError("Timeout", "must be non-negative")
	}
	if config.Retries < 0 {
		return newConfigValidationError("Retries", "must be non-negative")
	}

	return nil
}

func (c *Client) request(ctx context.Context, options requestOptions, result any) error {
	if ctx == nil {
		return errors.New("monime: nil context")
	}
	if err := validateRequestConfig(options.Config); err != nil {
		return err
	}

	requestURL := c.buildURL(options.Path, options.Query)
	body, err := marshalBody(options.Body)
	if err != nil {
		return err
	}

	timeout, retries := c.requestSettings(options.Config)
	headers, err := c.requestHeaders(options.Method, body != nil, options.Config)
	if err != nil {
		return err
	}

	for retryIndex := 0; ; retryIndex++ {
		err = c.executeAttempt(ctx, options.Method, requestURL, body, headers, timeout, result)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil || retryIndex >= retries || !isRetryable(err) {
			return err
		}

		delay := c.calculateRetryDelay(retryIndex, err)
		if err := sleep(ctx, delay); err != nil {
			return err
		}
	}
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

func (c *Client) requestSettings(config *RequestConfig) (time.Duration, int) {
	if config == nil {
		return c.timeout, c.retries
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = c.timeout
	}
	retries := config.Retries
	if retries == 0 {
		retries = c.retries
	}

	return timeout, retries
}

func (c *Client) requestHeaders(method string, hasBody bool, config *RequestConfig) (http.Header, error) {
	headers := http.Header{
		"Authorization":   {"Bearer " + c.accessToken},
		"Monime-Space-Id": {c.spaceID},
		"Monime-Version":  {c.apiVersion},
	}
	if hasBody {
		headers.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost {
		key := ""
		if config != nil {
			key = config.IdempotencyKey
		}
		if key == "" {
			var err error
			key, err = newUUIDv4()
			if err != nil {
				return nil, fmt.Errorf("monime: generate idempotency key: %w", err)
			}
		}
		headers.Set("Idempotency-Key", key)
	}

	return headers, nil
}

func (c *Client) executeAttempt(ctx context.Context, method string, requestURL *url.URL, body []byte, headers http.Header, timeout time.Duration, result any) error {
	attemptCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		attemptCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	request, err := http.NewRequestWithContext(attemptCtx, method, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("monime: create request: %w", err)
	}
	request.Header = headers.Clone()

	response, err := c.httpClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attemptCtx.Err() != nil {
			return &TimeoutError{Timeout: timeout}
		}
		return &NetworkError{Cause: err}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attemptCtx.Err() != nil {
			return &TimeoutError{Timeout: timeout}
		}
		return &NetworkError{Cause: err}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(response.StatusCode, response.Header, responseBody)
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return &APIError{
			Status:  response.StatusCode,
			Code:    response.StatusCode,
			Reason:  "invalid_json",
			Message: "invalid json response from server",
		}
	}

	return nil
}

func parseAPIError(status int, headers http.Header, body []byte) *APIError {
	apiError := &APIError{
		Status:  status,
		Code:    status,
		Reason:  "http_error",
		Message: "api request failed",
	}
	apiError.RetryAfter, _ = parseRetryAfter(headers.Get("Retry-After"), time.Now())

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
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
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
	maxDelay := time.Duration(math.MaxInt64 - int64(499*time.Millisecond))
	if baseDelay > float64(maxDelay) {
		baseDelay = float64(maxDelay)
	}

	return time.Duration(baseDelay) + time.Duration(mrand.IntN(500))*time.Millisecond
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

// do executes a typed API operation. Resource services use it to decode the
// documented result envelope into output.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, input any, output any) (Response, error) {
	if ctx == nil {
		return Response{}, errors.New("monime: nil context")
	}

	operationCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && c.timeout > 0 {
		operationCtx, cancel = context.WithTimeout(ctx, c.timeout)
	}
	defer cancel()

	body, err := marshalBody(input)
	if err != nil {
		return Response{}, err
	}
	requestURL := c.buildURL(path, query)
	headers, err := c.requestHeaders(method, body != nil, nil)
	if err != nil {
		return Response{}, err
	}

	var lastErr error
	for attempt := 0; ; attempt++ {
		response, err := c.doAttempt(operationCtx, method, requestURL, body, headers, output)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if operationCtx.Err() != nil {
			if ctx.Err() != nil {
				return Response{}, ctx.Err()
			}
			return Response{}, &TimeoutError{Timeout: c.timeout}
		}
		if attempt >= c.retries || (method != http.MethodGet && method != http.MethodPost) || !isRetryable(err) {
			return Response{}, err
		}
		if err := sleep(operationCtx, c.calculateRetryDelay(attempt, lastErr)); err != nil {
			return Response{}, err
		}
	}
}

func (c *Client) doAttempt(ctx context.Context, method string, requestURL *url.URL, body []byte, headers http.Header, output any) (Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("monime: create request: %w", err)
	}
	request.Header = headers.Clone()
	response, err := c.httpClient.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return Response{}, ctx.Err()
		}
		return Response{}, &NetworkError{Cause: err}
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		return Response{}, &NetworkError{Cause: err}
	}
	metadata := newResponse(response.StatusCode, response.Header)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return metadata, parseAPIError(response.StatusCode, response.Header, bodyBytes)
	}
	if len(bodyBytes) == 0 || output == nil {
		return metadata, nil
	}
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		return metadata, &APIError{Status: response.StatusCode, Code: response.StatusCode, Reason: "invalid_json", Message: "invalid json response from server", Body: bodyBytes}
	}
	if len(envelope.Result) != 0 && string(envelope.Result) != "null" {
		if err := json.Unmarshal(envelope.Result, output); err != nil {
			return metadata, &APIError{Status: response.StatusCode, Code: response.StatusCode, Reason: "invalid_result", Message: "invalid result response from server", Body: bodyBytes}
		}
	}
	return metadata, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, config *RequestConfig) (*apiResponse, error) {
	r := &apiResponse{}
	err := c.request(ctx, requestOptions{Method: http.MethodGet, Path: path, Query: query, Config: config}, r)
	return r, err
}
func (c *Client) getList(ctx context.Context, path string, query url.Values, config *RequestConfig) (*apiListResponse, error) {
	r := &apiListResponse{}
	err := c.request(ctx, requestOptions{Method: http.MethodGet, Path: path, Query: query, Config: config}, r)
	return r, err
}
func (c *Client) post(ctx context.Context, path string, body any, config *RequestConfig) (*apiResponse, error) {
	r := &apiResponse{}
	err := c.request(ctx, requestOptions{Method: http.MethodPost, Path: path, Body: body, Config: config}, r)
	return r, err
}
func (c *Client) patch(ctx context.Context, path string, body any, config *RequestConfig) (*apiResponse, error) {
	r := &apiResponse{}
	err := c.request(ctx, requestOptions{Method: http.MethodPatch, Path: path, Body: body, Config: config}, r)
	return r, err
}
func (c *Client) delete(ctx context.Context, path string, config *RequestConfig) (*apiDeleteResponse, error) {
	r := &apiDeleteResponse{}
	err := c.request(ctx, requestOptions{Method: http.MethodDelete, Path: path, Config: config}, r)
	return r, err
}
