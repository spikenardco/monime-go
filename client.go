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
	defaultRetryDelay   = time.Second
	defaultRetryBackoff = 2
	sdkVersion          = "0.1.0"
)

// Config configures a Client.
type Config struct {
	SpaceID       string
	AccessToken   string
	BaseURL       string
	APIVersion    string
	Timeout       time.Duration
	Retries       int
	RetryDelay    time.Duration
	RetryBackoff  float64
	WebhookSecret string
	Transport     http.RoundTripper
}

// Client is the Monime API client. It is immutable after construction and
// safe for concurrent use.
type Client struct {
	spaceID               string
	accessToken           string
	baseURL               *url.URL
	apiVersion            string
	timeout               time.Duration
	retries               int
	retryDelay            time.Duration
	retryBackoff          float64
	userAgent             string
	webhookSecret         string
	httpClient            *http.Client
	banks                 *BankService
	mobileMoney           *MobileMoneyService
	providerKYC           *ProviderKYCService
	financialAccounts     *FinancialAccountService
	financialTransactions *FinancialTransactionService
	paymentCodes          *PaymentCodeService
	payments              *PaymentService
	checkoutSessions      *CheckoutSessionService
	ussdOTPs              *USSDOTPService
	payouts               *PayoutService
	internalTransfers     *InternalTransferService
	receipts              *ReceiptService
	webhooks              *WebhookService
}

// New creates a Client from config.
func New(config Config) (*Client, error) {
	config = withDefaults(config)
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, newValidationError("BaseURL", "must be a valid HTTPS URL")
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	httpClient := &http.Client{CheckRedirect: sameHostRedirectPolicy}
	if config.Transport != nil {
		httpClient.Transport = config.Transport
	}

	client := &Client{
		spaceID:       config.SpaceID,
		accessToken:   config.AccessToken,
		baseURL:       baseURL,
		apiVersion:    config.APIVersion,
		timeout:       config.Timeout,
		retries:       config.Retries,
		retryDelay:    config.RetryDelay,
		retryBackoff:  config.RetryBackoff,
		userAgent:     "monime-go/" + sdkVersion,
		webhookSecret: config.WebhookSecret,
		httpClient:    httpClient,
	}
	client.banks = &BankService{client: client}
	client.mobileMoney = &MobileMoneyService{client: client}
	client.providerKYC = &ProviderKYCService{client: client}
	client.financialAccounts = &FinancialAccountService{client: client}
	client.financialTransactions = &FinancialTransactionService{client: client}
	client.paymentCodes = &PaymentCodeService{client: client}
	client.payments = &PaymentService{client: client}
	client.checkoutSessions = &CheckoutSessionService{client: client}
	client.ussdOTPs = &USSDOTPService{client: client}
	client.payouts = &PayoutService{client: client}
	client.internalTransfers = &InternalTransferService{client: client}
	client.receipts = &ReceiptService{client: client}
	client.webhooks = &WebhookService{client: client}

	return client, nil
}

// Banks returns the bank provider service.
func (c *Client) Banks() *BankService { return c.banks }

// MobileMoney returns the mobile money provider service.
func (c *Client) MobileMoney() *MobileMoneyService { return c.mobileMoney }

// ProviderKYC returns the provider KYC service.
func (c *Client) ProviderKYC() *ProviderKYCService { return c.providerKYC }

// FinancialAccounts returns the financial account service.
func (c *Client) FinancialAccounts() *FinancialAccountService { return c.financialAccounts }

// FinancialTransactions returns the financial transaction service.
func (c *Client) FinancialTransactions() *FinancialTransactionService {
	return c.financialTransactions
}

// PaymentCodes returns the payment code service.
func (c *Client) PaymentCodes() *PaymentCodeService { return c.paymentCodes }

// Payments returns the payment service.
func (c *Client) Payments() *PaymentService { return c.payments }

// CheckoutSessions returns the checkout session service.
func (c *Client) CheckoutSessions() *CheckoutSessionService { return c.checkoutSessions }

// USSDOTPs returns the USSD OTP service.
func (c *Client) USSDOTPs() *USSDOTPService { return c.ussdOTPs }

// Payouts returns the payout service.
func (c *Client) Payouts() *PayoutService { return c.payouts }

// InternalTransfers returns the internal transfer service.
func (c *Client) InternalTransfers() *InternalTransferService { return c.internalTransfers }

// Receipts returns the receipt service.
func (c *Client) Receipts() *ReceiptService { return c.receipts }

// Webhooks returns the webhook service.
func (c *Client) Webhooks() *WebhookService { return c.webhooks }

func withDefaults(config Config) Config {
	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.APIVersion == "" {
		config.APIVersion = string(APIVersionCaph20250823)
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Retries > 0 && config.RetryDelay == 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.Retries > 0 && config.RetryBackoff == 0 {
		config.RetryBackoff = defaultRetryBackoff
	}
	return config
}
