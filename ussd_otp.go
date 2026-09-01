package monime

import (
	"context"
	"net/url"
)

// USSDOTPService manages USSD OTP verification requests.
type USSDOTPService struct {
	client *Client
}

// Create creates a USSD OTP verification request.
func (s *USSDOTPService) Create(ctx context.Context, input any, config *RequestConfig) (*APIResponse, error) {
	return s.client.post(ctx, "/ussd-otps", input, config)
}

// Get retrieves a USSD OTP by ID.
func (s *USSDOTPService) Get(ctx context.Context, id string, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/ussd-otps/"+url.PathEscape(id), nil, config)
}

// List retrieves USSD OTP verification requests.
func (s *USSDOTPService) List(ctx context.Context, params url.Values, config *RequestConfig) (*APIListResponse, error) {
	return s.client.getList(ctx, "/ussd-otps", params, config)
}

// Delete deletes a USSD OTP verification request.
func (s *USSDOTPService) Delete(ctx context.Context, id string, config *RequestConfig) (*APIDeleteResponse, error) {
	return s.client.delete(ctx, "/ussd-otps/"+url.PathEscape(id), config)
}
