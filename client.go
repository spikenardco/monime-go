package monime

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL      = "https://api.monime.io"
	defaultTimeout      = 30 * time.Second
	defaultRetries      = 2
	defaultRetryDelay   = time.Second
	defaultRetryBackoff = 2
)

// Config configures a Client.
type Config struct {
	SpaceID      string
	AccessToken  string
	BaseURL      string
	APIVersion   APIVersion
	Timeout      time.Duration
	Retries      int
	RetryDelay   time.Duration
	RetryBackoff float64
}

// Client is a thin client for the Monime API.
type Client struct {
	spaceID      string
	accessToken  string
	baseURL      *url.URL
	apiVersion   APIVersion
	timeout      time.Duration
	retries      int
	retryDelay   time.Duration
	retryBackoff float64
	httpClient   *http.Client
}

// New creates a Client from config.
func New(config Config) (*Client, error) {
	config = withDefaults(config)
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, newConfigValidationError("BaseURL", "must be a valid HTTPS URL")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	return &Client{
		spaceID:      config.SpaceID,
		accessToken:  config.AccessToken,
		baseURL:      baseURL,
		apiVersion:   config.APIVersion,
		timeout:      config.Timeout,
		retries:      config.Retries,
		retryDelay:   config.RetryDelay,
		retryBackoff: config.RetryBackoff,
		httpClient:   &http.Client{},
	}, nil
}

func withDefaults(config Config) Config {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.APIVersion == "" {
		config.APIVersion = APIVersionCaph20250823
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Retries == 0 {
		config.Retries = defaultRetries
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = defaultRetryBackoff
	}

	return config
}
