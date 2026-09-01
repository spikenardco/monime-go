package monime

import (
	"context"
	"net/url"
)

// FinancialTransactionService retrieves financial transactions.
type FinancialTransactionService struct {
	client *Client
}

// Get retrieves a financial transaction by ID.
func (s *FinancialTransactionService) Get(ctx context.Context, id string, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/financial-transactions/"+url.PathEscape(id), nil, config)
}

// List retrieves financial transactions.
func (s *FinancialTransactionService) List(ctx context.Context, params url.Values, config *RequestConfig) (*APIListResponse, error) {
	return s.client.getList(ctx, "/financial-transactions", params, config)
}
