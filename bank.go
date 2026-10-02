package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ProviderStatus reports whether a financial provider is active.
type ProviderStatus struct {
	Active bool `json:"active"`
}

// PayoutFeature describes a provider's payout support.
type PayoutFeature struct {
	CanPayTo bool     `json:"canPayTo"`
	Schemes  []string `json:"schemes,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// PaymentFeature describes a provider's payment support.
type PaymentFeature struct {
	CanPayFrom bool     `json:"canPayFrom"`
	Schemes    []string `json:"schemes,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
}

// KYCVerificationFeature describes a provider's account-verification support.
type KYCVerificationFeature struct {
	CanVerifyAccount bool     `json:"canVerifyAccount"`
	Metadata         Metadata `json:"metadata,omitempty"`
}

// ProviderFeatureSet describes the Monime operations a provider supports.
type ProviderFeatureSet struct {
	Payout          PayoutFeature          `json:"payout"`
	Payment         PaymentFeature         `json:"payment"`
	KYCVerification KYCVerificationFeature `json:"kycVerification"`
}

// Bank is a supported bank provider.
type Bank struct {
	ProviderID string             `json:"providerId"`
	Name       string             `json:"name"`
	Country    string             `json:"country"`
	Status     ProviderStatus     `json:"status"`
	FeatureSet ProviderFeatureSet `json:"featureSet"`
	CreateTime time.Time          `json:"createTime"`
	UpdateTime *time.Time         `json:"updateTime,omitempty"`
}

// BankListParams filters a bank list request. Country is required.
type BankListParams struct {
	Country string
	Limit   int
	After   string
}

// ListBanks returns banks available in a country.
func (c *Client) ListBanks(ctx context.Context, params BankListParams) (*Page[Bank], *Response, error) {
	if params.Country == "" {
		return nil, nil, newValidationError("Country", "must be non-empty")
	}
	page, response, err := doList[Bank](c, ctx, "/banks", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// GetBank returns a bank by provider ID.
func (c *Client) GetBank(ctx context.Context, providerID string) (*Bank, *Response, error) {
	if providerID == "" {
		return nil, nil, newValidationError("ProviderID", "must be non-empty")
	}
	var bank Bank
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/banks/" + url.PathEscape(providerID),
		output: &bank,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &bank, &response, nil
}

func (p BankListParams) values() url.Values {
	values := url.Values{"country": {p.Country}}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	return values
}
