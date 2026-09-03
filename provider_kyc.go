package monime

import (
	"context"
	"net/url"
)

type ProviderKYC struct {
	Account  ProviderKYCAccount  `json:"account"`
	Provider ProviderKYCProvider `json:"provider"`
}
type ProviderKYCAccount struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	HolderName string            `json:"holderName"`
	Metadata   map[string]string `json:"metadata"`
}
type ProviderKYCProvider struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
type ProviderKYCQuery struct{ AccountID string }
type ProviderKYCService struct{ client *Client }

func (s *ProviderKYCService) Get(ctx context.Context, providerID string, query ProviderKYCQuery) (ProviderKYC, Response, error) {
	var result ProviderKYC
	values := url.Values{}
	if query.AccountID != "" {
		values.Set("accountId", query.AccountID)
	}
	response, err := s.client.do(ctx, "GET", "/provider-kyc/"+url.PathEscape(providerID), values, nil, &result)
	return result, response, err
}
