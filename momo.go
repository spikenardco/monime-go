package monime

import (
	"context"
	"net/url"
)

// MobileMoneyService retrieves mobile money providers.
type MobileMoneyService struct {
	client *Client
}

// List retrieves mobile money providers.
func (s *MobileMoneyService) List(ctx context.Context, params url.Values, config *RequestConfig) (*APIListResponse, error) {
	return s.client.getList(ctx, "/momos", params, config)
}

// Get retrieves a mobile money provider by ID.
func (s *MobileMoneyService) Get(ctx context.Context, providerID string, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/momos/"+url.PathEscape(providerID), nil, config)
}
