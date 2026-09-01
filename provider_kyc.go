package monime

import (
	"context"
	"net/url"
)

// ProviderKYCService retrieves KYC information from financial providers.
type ProviderKYCService struct {
	client *Client
}

// Get retrieves KYC information for an account at a provider.
func (s *ProviderKYCService) Get(ctx context.Context, providerID string, params url.Values, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/provider-kyc/"+url.PathEscape(providerID), params, config)
}
