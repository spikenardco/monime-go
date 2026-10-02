package monime

import (
	"context"
	"net/http"
	"net/url"
)

// ProviderKYCProviderType identifies the provider hosting a KYC account.
type ProviderKYCProviderType string

const (
	ProviderKYCTypeMomo   ProviderKYCProviderType = "momo"
	ProviderKYCTypeBank   ProviderKYCProviderType = "bank"
	ProviderKYCTypeWallet ProviderKYCProviderType = "wallet"
)

// ProviderKYC contains provider account-holder information.
type ProviderKYC struct {
	Account  ProviderKYCAccount  `json:"account"`
	Provider ProviderKYCProvider `json:"provider"`
}

// ProviderKYCAccount describes the account verified by the provider.
type ProviderKYCAccount struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	HolderName string   `json:"holderName"`
	Metadata   Metadata `json:"metadata,omitempty"`
}

// ProviderKYCProvider identifies the provider that supplied KYC data.
type ProviderKYCProvider struct {
	ID   string                  `json:"id"`
	Type ProviderKYCProviderType `json:"type"`
	Name string                  `json:"name"`
}

// GetProviderKYC returns KYC data for an account held by a provider.
func (c *Client) GetProviderKYC(
	ctx context.Context,
	providerID string,
	accountID string,
) (*ProviderKYC, *Response, error) {
	if providerID == "" {
		return nil, nil, newValidationError("ProviderID", "must be non-empty")
	}
	if accountID == "" {
		return nil, nil, newValidationError("AccountID", "must be non-empty")
	}
	var result ProviderKYC
	query := url.Values{"accountId": {accountID}}
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/provider-kyc/" + url.PathEscape(providerID),
		query:  query,
		output: &result,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &result, &response, nil
}
