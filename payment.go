package monime

import (
	"context"
	"net/url"
)

// PaymentService retrieves and updates payments.
type PaymentService struct {
	client *Client
}

// Get retrieves a payment by ID.
func (s *PaymentService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/payments/"+url.PathEscape(id), nil, config)
}

// List retrieves payments.
func (s *PaymentService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/payments", params, config)
}

// Update updates a payment.
func (s *PaymentService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.patch(ctx, "/payments/"+url.PathEscape(id), input, config)
}
