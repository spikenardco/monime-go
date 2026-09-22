package monime

import (
	"context"
	"net/url"
)

// ProviderKYCProviderType identifies the provider hosting a KYC account.
type ProviderKYCProviderType string

const (
	ProviderKYCTypeMomo   ProviderKYCProviderType = "momo"
	ProviderKYCTypeBank   ProviderKYCProviderType = "bank"
	ProviderKYCTypeWallet ProviderKYCProviderType = "wallet"
)

type ProviderKYC struct {
	Account  ProviderKYCAccount  `json:"account"`
	Provider ProviderKYCProvider `json:"provider"`
}

type ProviderKYCAccount struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	HolderName string   `json:"holderName"`
	Metadata   Metadata `json:"metadata,omitempty"`
}

type ProviderKYCProvider struct {
	ID   string                  `json:"id"`
	Type ProviderKYCProviderType `json:"type"`
	Name string                  `json:"name"`
}

type ProviderKYCGetQuery struct{ AccountID string }

type ProviderKYCService struct{ client *Client }

func (s *ProviderKYCService) Get(
	ctx context.Context,
	providerID string,
	query ProviderKYCGetQuery,
) (ProviderKYC, Response, error) {
	if providerID == "" {
		return ProviderKYC{}, Response{}, newValidationError("ProviderID", "must be non-empty")
	}
	if query.AccountID == "" {
		return ProviderKYC{}, Response{}, newValidationError("AccountID", "must be non-empty")
	}
	var result ProviderKYC
	values := url.Values{}
	values.Set("accountId", query.AccountID)
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/provider-kyc/" + url.PathEscape(providerID),
		query:  values,
		output: &result,
	})
	return result, response, err
}
