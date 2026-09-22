package monime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// WebhookVerificationType identifies the signature algorithm of a webhook.
type WebhookVerificationType string

const (
	WebhookVerificationHS256 WebhookVerificationType = "HS256"
	WebhookVerificationES256 WebhookVerificationType = "ES256"
)

// Webhook is an event-notification configuration.
type Webhook struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	URL                string                    `json:"url"`
	Enabled            bool                      `json:"enabled"`
	Events             []string                  `json:"events"`
	APIRelease         string                    `json:"apiRelease"`
	VerificationMethod WebhookVerificationMethod `json:"verificationMethod"`
	Headers            map[string]string         `json:"headers,omitempty"`
	AlertEmails        []string                  `json:"alertEmails,omitempty"`
	CreateTime         string                    `json:"createTime"`
	UpdateTime         *string                   `json:"updateTime,omitempty"`
	Metadata           Metadata                  `json:"metadata,omitempty"`
}

type WebhookVerificationMethod struct {
	Type      WebhookVerificationType `json:"type"`
	Secret    *string                 `json:"secret,omitempty"`
	PublicKey *string                 `json:"publicKey,omitempty"`
}

type CreateWebhookInput struct {
	Name               string                     `json:"name"`
	URL                string                     `json:"url"`
	APIRelease         string                     `json:"apiRelease"`
	Events             []string                   `json:"events"`
	Enabled            *bool                      `json:"enabled,omitempty"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty"`
	Metadata           Metadata                   `json:"metadata,omitempty"`
}

type UpdateWebhookInput struct {
	Name        Field[string]            `json:"name,omitzero"`
	URL         Field[string]            `json:"url,omitzero"`
	Enabled     Field[bool]              `json:"enabled,omitzero"`
	APIRelease  Field[string]            `json:"apiRelease,omitzero"`
	Events      Field[[]string]          `json:"events,omitzero"`
	Headers     Field[map[string]string] `json:"headers,omitzero"`
	AlertEmails Field[[]string]          `json:"alertEmails,omitzero"`
	Metadata    Field[Metadata]          `json:"metadata,omitzero"`
}

type WebhookListQuery struct {
	Limit int
	After string
}

type WebhookService struct{ client *Client }

func (s *WebhookService) Create(ctx context.Context, input CreateWebhookInput) (Webhook, Response, error) {
	if input.Name == "" {
		return Webhook{}, Response{}, newValidationError("Name", "must be non-empty")
	}
	if input.URL == "" {
		return Webhook{}, Response{}, newValidationError("URL", "must be non-empty")
	}
	if len(input.Events) == 0 {
		return Webhook{}, Response{}, newValidationError("Events", "must contain at least one event")
	}
	var result Webhook
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/webhooks",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *WebhookService) Get(ctx context.Context, id string) (Webhook, Response, error) {
	if id == "" {
		return Webhook{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Webhook
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/webhooks/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *WebhookService) List(ctx context.Context, query WebhookListQuery) (Page[Webhook], Response, error) {
	return doList[Webhook](s.client, ctx, "/webhooks", query.values())
}

func (s *WebhookService) Update(ctx context.Context, id string, input UpdateWebhookInput) (Webhook, Response, error) {
	if id == "" {
		return Webhook{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Webhook
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/webhooks/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *WebhookService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/webhooks/" + url.PathEscape(id),
	})
}

// Verify authenticates a webhook delivery using the client's configured
// HS256 secret. The raw body must be the exact bytes received from Monime.
func (s *WebhookService) Verify(rawBody []byte, signatureHeader string) (WebhookEvent, error) {
	if s.client.webhookSecret == "" {
		return WebhookEvent{}, fmt.Errorf("monime: configure WebhookSecret before verifying webhook signatures")
	}
	return VerifyWebhookSignature(rawBody, signatureHeader, s.client.webhookSecret)
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

// WebhookEvent is the envelope Monime sends for a webhook delivery.
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

// ParseWebhookEvent decodes a webhook envelope without authenticating it.
// Use Verify or VerifyWebhookSignature to authenticate first.
func ParseWebhookEvent(body []byte) (WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("monime: decode webhook event: %w", err)
	}
	if event.Event.ID == "" || event.Event.Name == "" || event.Object.ID == "" || event.Object.Type == "" {
		return WebhookEvent{}, fmt.Errorf("monime: invalid webhook event envelope")
	}
	return event, nil
}
