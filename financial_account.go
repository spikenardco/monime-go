package monime

import (
	"context"
	"net/url"
)

// FinancialAccountService manages financial accounts.
type FinancialAccountService struct {
	client *Client
}

// Create creates a financial account.
func (s *FinancialAccountService) Create(ctx context.Context, input any, config *RequestConfig) (*APIResponse, error) {
	return s.client.post(ctx, "/financial-accounts", input, config)
}

// Get retrieves a financial account by ID.
func (s *FinancialAccountService) Get(ctx context.Context, id string, params url.Values, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/financial-accounts/"+url.PathEscape(id), params, config)
}

// List retrieves financial accounts.
func (s *FinancialAccountService) List(ctx context.Context, params url.Values, config *RequestConfig) (*APIListResponse, error) {
	return s.client.getList(ctx, "/financial-accounts", params, config)
}

// Update updates a financial account.
func (s *FinancialAccountService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*APIResponse, error) {
	return s.client.patch(ctx, "/financial-accounts/"+url.PathEscape(id), input, config)
}
