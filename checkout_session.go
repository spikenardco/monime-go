package monime

import (
	"context"
	"net/url"
)

// CheckoutSessionService manages checkout sessions.
type CheckoutSessionService struct {
	client *Client
}

// Create creates a checkout session.
func (s *CheckoutSessionService) Create(ctx context.Context, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.post(ctx, "/checkout-sessions", input, config)
}

// Get retrieves a checkout session by ID.
func (s *CheckoutSessionService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/checkout-sessions/"+url.PathEscape(id), nil, config)
}

// List retrieves checkout sessions.
func (s *CheckoutSessionService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/checkout-sessions", params, config)
}

// Delete deletes a checkout session.
func (s *CheckoutSessionService) Delete(ctx context.Context, id string, config *RequestConfig) (*apiDeleteResponse, error) {
	return s.client.delete(ctx, "/checkout-sessions/"+url.PathEscape(id), config)
}
