package monime

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

const (
	apiEndpoint        = "https://api.monime.io"
	defaultTimeout     = 30 * time.Second
	defaultMaxAttempts = 3
	defaultBaseDelay   = time.Second
	defaultMaxDelay    = 8 * time.Second
)

// Config configures a Client.
type Config struct {
	SpaceID     string
	AccessToken string
	HTTPClient  *http.Client
	Timeout     time.Duration
	Retry       RetryConfig
}

// RetryConfig controls retries for GET and idempotent POST requests.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// Client is an immutable Monime API client that is safe for concurrent use.
type Client struct {
	spaceID     string
	accessToken string
	baseURL     *url.URL
	timeout     time.Duration
	retry       RetryConfig
	userAgent   string
	httpClient  *http.Client
}

// New creates a Client from config.
func New(config Config) (*Client, error) {
	config = withDefaults(config)
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, newValidationError("APIEndpoint", "is invalid")
	}

	httpClient := &http.Client{}
	if config.HTTPClient != nil {
		clone := *config.HTTPClient
		httpClient = &clone
	}
	httpClient.Timeout = 0
	httpClient.CheckRedirect = refuseRedirects

	return &Client{
		spaceID:     config.SpaceID,
		accessToken: config.AccessToken,
		baseURL:     baseURL,
		timeout:     config.Timeout,
		retry:       config.Retry,
		userAgent:   "monime-go/" + sdkVersion,
		httpClient:  httpClient,
	}, nil
}

func withDefaults(config Config) Config {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Retry.MaxAttempts == 0 {
		config.Retry.MaxAttempts = defaultMaxAttempts
	}
	if config.Retry.BaseDelay == 0 {
		config.Retry.BaseDelay = defaultBaseDelay
	}
	if config.Retry.MaxDelay == 0 {
		config.Retry.MaxDelay = defaultMaxDelay
	}
	return config
}

func validateConfig(config Config) error {
	if config.SpaceID == "" {
		return newValidationError("SpaceID", "must be non-empty")
	}
	if config.AccessToken == "" {
		return newValidationError("AccessToken", "must be non-empty")
	}
	if config.Timeout < 0 {
		return newValidationError("Timeout", "must be non-negative")
	}
	if config.Retry.MaxAttempts < 0 {
		return newValidationError("Retry.MaxAttempts", "must be non-negative")
	}
	if config.Retry.BaseDelay < 0 {
		return newValidationError("Retry.BaseDelay", "must be non-negative")
	}
	if config.Retry.MaxDelay < 0 {
		return newValidationError("Retry.MaxDelay", "must be non-negative")
	}
	if config.Retry.MaxDelay < config.Retry.BaseDelay {
		return newValidationError("Retry.MaxDelay", "must not be less than Retry.BaseDelay")
	}
	if config.HTTPClient != nil && config.HTTPClient.Timeout < 0 {
		return newValidationError("HTTPClient.Timeout", "must be non-negative")
	}
	return nil
}

func refuseRedirects(_ *http.Request, _ []*http.Request) error {
	return errors.New("monime: redirects are not allowed")
}
