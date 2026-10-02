package monime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// WebhookVerificationType identifies the signature algorithm of a webhook.
type WebhookVerificationType string

const (
	WebhookVerificationHS256 WebhookVerificationType = "HS256"
	WebhookVerificationES256 WebhookVerificationType = "ES256"
)

// Webhook is an event-notification configuration.
type Webhook struct {
	ID                 string                     `json:"id"`
	Name               string                     `json:"name"`
	URL                string                     `json:"url"`
	Enabled            bool                       `json:"enabled"`
	Events             []string                   `json:"events,omitempty"`
	APIRelease         string                     `json:"apiRelease"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty"`
	CreateTime         time.Time                  `json:"createTime"`
	UpdateTime         *time.Time                 `json:"updateTime,omitempty"`
	Metadata           Metadata                   `json:"metadata,omitempty"`
}

// WebhookVerificationMethod describes webhook verification configured by the API.
type WebhookVerificationMethod struct {
	Type      WebhookVerificationType `json:"type"`
	Secret    *string                 `json:"secret,omitempty"`
	PublicKey *string                 `json:"publicKey,omitempty"`
}

// WebhookCreateVerification is a valid verification method for webhook creation.
// Construct values with HS256WebhookVerification or ES256WebhookVerification.
type WebhookCreateVerification struct {
	typeName WebhookVerificationType
	secret   string
}

// HS256WebhookVerification returns an HMAC webhook verification method.
func HS256WebhookVerification(secret string) WebhookCreateVerification {
	return WebhookCreateVerification{typeName: WebhookVerificationHS256, secret: secret}
}

// ES256WebhookVerification returns an ECDSA webhook verification method.
// Monime generates the key pair and returns the public key on the Webhook.
func ES256WebhookVerification() WebhookCreateVerification {
	return WebhookCreateVerification{typeName: WebhookVerificationES256}
}

// MarshalJSON emits only fields accepted for the selected verification method.
func (v WebhookCreateVerification) MarshalJSON() ([]byte, error) {
	switch v.typeName {
	case WebhookVerificationHS256:
		return json.Marshal(struct {
			Type   WebhookVerificationType `json:"type"`
			Secret string                  `json:"secret"`
		}{Type: v.typeName, Secret: v.secret})
	case WebhookVerificationES256:
		return json.Marshal(struct {
			Type WebhookVerificationType `json:"type"`
		}{Type: v.typeName})
	default:
		return nil, errors.New("monime: invalid webhook verification method")
	}
}

// CreateWebhookInput contains fields accepted when creating a webhook.
type CreateWebhookInput struct {
	Name               string                     `json:"name"`
	URL                string                     `json:"url"`
	APIRelease         string                     `json:"apiRelease"`
	Events             []string                   `json:"events"`
	Enabled            *bool                      `json:"enabled,omitempty"`
	VerificationMethod *WebhookCreateVerification `json:"verificationMethod,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty"`
	Metadata           Metadata                   `json:"metadata,omitempty"`
}

// UpdateWebhookInput contains mutable webhook fields.
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

// WebhookListParams controls webhook pagination.
type WebhookListParams struct {
	Limit int
	After string
}

// CreateWebhook creates a webhook endpoint configuration.
func (c *Client) CreateWebhook(
	ctx context.Context,
	idempotencyKey string,
	input CreateWebhookInput,
) (*Webhook, *Response, error) {
	if input.Name == "" {
		return nil, nil, newValidationError("Name", "must be non-empty")
	}
	if input.URL == "" {
		return nil, nil, newValidationError("URL", "must be non-empty")
	}
	if input.APIRelease == "" {
		return nil, nil, newValidationError("APIRelease", "must be non-empty")
	}
	if len(input.Events) == 0 {
		return nil, nil, newValidationError("Events", "must contain at least one event")
	}
	if err := validateWebhookVerification(input.VerificationMethod); err != nil {
		return nil, nil, err
	}
	var webhook Webhook
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/webhooks",
		input:          input,
		output:         &webhook,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &webhook, &response, nil
}

// GetWebhook returns a webhook by ID.
func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var webhook Webhook
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/webhooks/" + url.PathEscape(id),
		output: &webhook,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &webhook, &response, nil
}

// ListWebhooks returns a page of webhook configurations.
func (c *Client) ListWebhooks(
	ctx context.Context,
	params WebhookListParams,
) (*Page[Webhook], *Response, error) {
	page, response, err := doList[Webhook](c, ctx, "/webhooks", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdateWebhook updates mutable fields on a webhook.
func (c *Client) UpdateWebhook(
	ctx context.Context,
	id string,
	input UpdateWebhookInput,
) (*Webhook, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !webhookUpdateSet(input) {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var webhook Webhook
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/webhooks/" + url.PathEscape(id),
		input:  input,
		output: &webhook,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &webhook, &response, nil
}

// DeleteWebhook permanently deletes a webhook.
func (c *Client) DeleteWebhook(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/webhooks/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func validateWebhookVerification(method *WebhookCreateVerification) error {
	if method == nil {
		return nil
	}
	switch method.typeName {
	case WebhookVerificationHS256:
		if len(method.secret) < 32 || len(method.secret) > 256 {
			return newValidationError("VerificationMethod.Secret", "must contain 32 to 256 characters")
		}
	case WebhookVerificationES256:
		if method.secret != "" {
			return newValidationError("VerificationMethod", "ES256 must not include a secret")
		}
	default:
		return newValidationError("VerificationMethod", "must be created with a verification constructor")
	}
	return nil
}

func webhookUpdateSet(input UpdateWebhookInput) bool {
	return input.Name.IsSet() || input.URL.IsSet() || input.Enabled.IsSet() ||
		input.APIRelease.IsSet() || input.Events.IsSet() || input.Headers.IsSet() ||
		input.AlertEmails.IsSet() || input.Metadata.IsSet()
}

func (p WebhookListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
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

// WebhookEventDescriptor identifies the event and when it occurred.
type WebhookEventDescriptor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// WebhookEventObject identifies the resource carried by an event.
type WebhookEventObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ParseWebhookEvent decodes a webhook envelope without authenticating it.
// Use VerifyWebhookSignature to authenticate first.
func ParseWebhookEvent(body []byte) (WebhookEvent, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return WebhookEvent{}, fmt.Errorf("monime: decode webhook event: expected JSON object")
	}
	var event WebhookEvent
	if err := json.Unmarshal(trimmed, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("monime: decode webhook event: %w", err)
	}
	if event.Event.ID == "" || event.Event.Name == "" || event.Object.ID == "" || event.Object.Type == "" {
		return WebhookEvent{}, fmt.Errorf("monime: invalid webhook event envelope")
	}
	return event, nil
}
