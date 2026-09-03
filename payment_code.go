package monime

import (
	"context"
	"net/url"
)

// PaymentCodeService manages payment codes.
type PaymentCodeService struct {
	client *Client
}

// Create creates a payment code.
func (s *PaymentCodeService) Create(ctx context.Context, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.post(ctx, "/payment-codes", input, config)
}

// Get retrieves a payment code by ID.
func (s *PaymentCodeService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/payment-codes/"+url.PathEscape(id), nil, config)
}

// List retrieves payment codes.
func (s *PaymentCodeService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/payment-codes", params, config)
}

// Update updates a payment code.
func (s *PaymentCodeService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.patch(ctx, "/payment-codes/"+url.PathEscape(id), input, config)
}

// Delete deletes a payment code.
func (s *PaymentCodeService) Delete(ctx context.Context, id string, config *RequestConfig) (*apiDeleteResponse, error) {
	return s.client.delete(ctx, "/payment-codes/"+url.PathEscape(id), config)
}
