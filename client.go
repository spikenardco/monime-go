package monime

import (
	"net/http"
	"net/url"
	"strings"
	"time"
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
