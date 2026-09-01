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
	spaceID               string
	accessToken           string
	baseURL               *url.URL
	apiVersion            APIVersion
	timeout               time.Duration
	retries               int
	retryDelay            time.Duration
	retryBackoff          float64
	httpClient            *http.Client
	banks                 *BankService
	mobileMoney           *MobileMoneyService
	providerKYC           *ProviderKYCService
	financialAccounts     *FinancialAccountService
	financialTransactions *FinancialTransactionService
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

	client := &Client{
		spaceID:      config.SpaceID,
		accessToken:  config.AccessToken,
		baseURL:      baseURL,
		apiVersion:   config.APIVersion,
		timeout:      config.Timeout,
		retries:      config.Retries,
		retryDelay:   config.RetryDelay,
		retryBackoff: config.RetryBackoff,
		httpClient:   &http.Client{},
	}
	client.banks = &BankService{client: client}
	client.mobileMoney = &MobileMoneyService{client: client}
	client.providerKYC = &ProviderKYCService{client: client}
	client.financialAccounts = &FinancialAccountService{client: client}
	client.financialTransactions = &FinancialTransactionService{client: client}

	return client, nil
}

// Banks returns the bank provider service.
func (c *Client) Banks() *BankService {
	return c.banks
}

// MobileMoney returns the mobile money provider service.
func (c *Client) MobileMoney() *MobileMoneyService {
	return c.mobileMoney
}

// ProviderKYC returns the provider KYC service.
func (c *Client) ProviderKYC() *ProviderKYCService {
	return c.providerKYC
}

// FinancialAccounts returns the financial account service.
func (c *Client) FinancialAccounts() *FinancialAccountService {
	return c.financialAccounts
}

// FinancialTransactions returns the financial transaction service.
func (c *Client) FinancialTransactions() *FinancialTransactionService {
	return c.financialTransactions
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
