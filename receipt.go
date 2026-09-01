package monime

import (
	"context"
	"net/url"
)

// ReceiptService manages receipts.
type ReceiptService struct {
	client *Client
}

// Get retrieves a receipt by order number.
func (s *ReceiptService) Get(ctx context.Context, orderNumber string, config *RequestConfig) (*APIResponse, error) {
	return s.client.get(ctx, "/receipts/"+url.PathEscape(orderNumber), nil, config)
}

// Redeem redeems a receipt.
func (s *ReceiptService) Redeem(ctx context.Context, orderNumber string, input any, config *RequestConfig) (*APIResponse, error) {
	return s.client.post(ctx, "/receipts/"+url.PathEscape(orderNumber)+"/redeem", input, config)
}
