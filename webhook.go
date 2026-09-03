package monime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type Webhook struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	URL                string                    `json:"url"`
	Enabled            bool                      `json:"enabled"`
	Events             []string                  `json:"events"`
	APIRelease         string                    `json:"apiRelease"`
	VerificationMethod WebhookVerificationMethod `json:"verificationMethod"`
	Headers            map[string]string         `json:"headers"`
	AlertEmails        []string                  `json:"alertEmails"`
	Metadata           map[string]string         `json:"metadata"`
}
type WebhookVerificationMethod struct {
	Type string `json:"type"`
}
type CreateWebhookInput struct {
	Name               string                     `json:"name"`
	URL                string                     `json:"url"`
	Events             []string                   `json:"events"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty"`
	Metadata           map[string]string          `json:"metadata,omitempty"`
}
type UpdateWebhookInput struct {
	Name     Field[string]            `json:"name,omitzero"`
	URL      Field[string]            `json:"url,omitzero"`
	Enabled  Field[bool]              `json:"enabled,omitzero"`
	Events   Field[[]string]          `json:"events,omitzero"`
	Metadata Field[map[string]string] `json:"metadata,omitzero"`
}
type WebhookListQuery struct {
	Limit int
	After string
}
type WebhookService struct{ client *Client }

func (s *WebhookService) Create(ctx context.Context, input CreateWebhookInput) (Webhook, Response, error) {
	var result Webhook
	response, err := s.client.do(ctx, "POST", "/webhooks", nil, input, &result)
	return result, response, err
}
func (s *WebhookService) Get(ctx context.Context, id string) (Webhook, Response, error) {
	var result Webhook
	response, err := s.client.do(ctx, "GET", "/webhooks/"+url.PathEscape(id), nil, nil, &result)
	return result, response, err
}
func (s *WebhookService) List(ctx context.Context, query WebhookListQuery) (Page[Webhook], Response, error) {
	var items []Webhook
	response, err := s.client.do(ctx, "GET", "/webhooks", query.values(), nil, &items)
	return Page[Webhook]{Items: items}, response, err
}
func (s *WebhookService) Update(ctx context.Context, id string, input UpdateWebhookInput) (Webhook, Response, error) {
	var result Webhook
	response, err := s.client.do(ctx, "PATCH", "/webhooks/"+url.PathEscape(id), nil, input, &result)
	return result, response, err
}
func (s *WebhookService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/webhooks/"+url.PathEscape(id), nil, nil, nil)
}
func (q WebhookListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	return values
}

type WebhookEvent struct {
	APIVersion string                 `json:"apiVersion"`
	Event      WebhookEventDescriptor `json:"event"`
	Object     WebhookEventObject     `json:"object"`
	Data       json.RawMessage        `json:"data"`
}
type WebhookEventDescriptor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}
type WebhookEventObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func ParseWebhookEvent(body []byte) (WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("monime: decode webhook event: %w", err)
	}
	return event, nil
}
