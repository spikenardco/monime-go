package monime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// WebhookService manages webhooks.
type WebhookService struct {
	client *Client
}

// Create creates a webhook.
//
// Deprecated: Create webhooks from the dashboard instead; API support is not guaranteed.
func (s *WebhookService) Create(ctx context.Context, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.post(ctx, "/webhooks", input, config)
}

// Get retrieves a webhook by ID.
func (s *WebhookService) Get(ctx context.Context, id string, config *RequestConfig) (*apiResponse, error) {
	return s.client.get(ctx, "/webhooks/"+url.PathEscape(id), nil, config)
}

// List retrieves webhooks.
func (s *WebhookService) List(ctx context.Context, params url.Values, config *RequestConfig) (*apiListResponse, error) {
	return s.client.getList(ctx, "/webhooks", params, config)
}

// Update updates a webhook.
func (s *WebhookService) Update(ctx context.Context, id string, input any, config *RequestConfig) (*apiResponse, error) {
	return s.client.patch(ctx, "/webhooks/"+url.PathEscape(id), input, config)
}

// Delete deletes a webhook.
func (s *WebhookService) Delete(ctx context.Context, id string, config *RequestConfig) (*apiDeleteResponse, error) {
	return s.client.delete(ctx, "/webhooks/"+url.PathEscape(id), config)
}

// WebhookEvent is the envelope sent by Monime for a webhook event.
type WebhookEvent struct {
	APIVersion string `json:"apiVersion"`
	Event      struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Timestamp string `json:"timestamp"`
	} `json:"event"`
	Object struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"object"`
	Data json.RawMessage `json:"data"`
}

// ParseWebhookEvent decodes a webhook event body.
func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	event := &WebhookEvent{}
	if err := json.Unmarshal(body, event); err != nil {
		return nil, fmt.Errorf("monime: decode webhook event: %w", err)
	}

	return event, nil
}
