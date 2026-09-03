package monime

import (
	"context"
	"net/url"
)

// InternalTransferService manages internal transfers.
type InternalTransferService struct {
	client *Client
}

// Create creates an internal transfer.
func (s *InternalTransferService) Create(ctx context.Context, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.post(ctx, "/internal-transfers", input, config)
}

// Get retrieves an internal transfer by ID.
func (s *InternalTransferService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/internal-transfers/"+url.PathEscape(id), nil, config)
}

// List retrieves internal transfers.
func (s *InternalTransferService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/internal-transfers", params, config)
}

// Update updates an internal transfer.
func (s *InternalTransferService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.patch(ctx, "/internal-transfers/"+url.PathEscape(id), input, config)
}

// Delete deletes an internal transfer.
func (s *InternalTransferService) Delete(ctx context.Context, id string, config *RequestConfig) (*apiDeleteResponse, error) {
	return s.client.delete(ctx, "/internal-transfers/"+url.PathEscape(id), config)
}
