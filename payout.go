package monime

import (
	"context"
	"net/url"
)

// PayoutService manages payouts.
type PayoutService struct {
	client *Client
}

// Create creates a payout.
func (s *PayoutService) Create(ctx context.Context, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.post(ctx, "/payouts", input, config)
}

// Get retrieves a payout by ID.
func (s *PayoutService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/payouts/"+url.PathEscape(id), nil, config)
}

// List retrieves payouts.
func (s *PayoutService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/payouts", params, config)
}

// Update updates a payout.
func (s *PayoutService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.patch(ctx, "/payouts/"+url.PathEscape(id), input, config)
}

// Delete deletes a payout.
func (s *PayoutService) Delete(ctx context.Context, id string, config *RequestConfig) (*apiDeleteResponse, error) {
	return s.client.delete(ctx, "/payouts/"+url.PathEscape(id), config)
}
