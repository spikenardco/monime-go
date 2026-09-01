package monime

import (
	"context"
	"net/url"
)

// BankService retrieves bank providers.
type BankService struct {
	client *Client
}

// List retrieves bank providers.
func (s *BankService) List(ctx context.Context, params url.Values, config *RequestConfig) (*APIListResponse, error) {
	return s.client.getList(ctx, "/banks", params, config)
}

// Get retrieves a bank provider by ID.
func (s *BankService) Get(ctx context.Context, providerID string, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/banks/"+url.PathEscape(providerID), nil, config)
}
