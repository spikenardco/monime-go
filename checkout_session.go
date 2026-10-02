package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// CheckoutSessionStatus is the lifecycle state of a checkout session.
type CheckoutSessionStatus string

const (
	CheckoutSessionPending   CheckoutSessionStatus = "pending"
	CheckoutSessionCompleted CheckoutSessionStatus = "completed"
	CheckoutSessionCancelled CheckoutSessionStatus = "cancelled"
	CheckoutSessionExpired   CheckoutSessionStatus = "expired"
)

// CheckoutSession is a hosted payment page for collecting payments.
type CheckoutSession struct {
	ID                 string                `json:"id"`
	Status             CheckoutSessionStatus `json:"status"`
	Name               string                `json:"name"`
	OrderNumber        *string               `json:"orderNumber,omitempty"`
	Reference          *string               `json:"reference,omitempty"`
	Description        *string               `json:"description,omitempty"`
	RedirectURL        string                `json:"redirectUrl"`
	CancelURL          string                `json:"cancelUrl"`
	SuccessURL         string                `json:"successUrl"`
	LineItems          CheckoutLineItems     `json:"lineItems"`
	FinancialAccountID *string               `json:"financialAccountId,omitempty"`
	BrandingOptions    *BrandingOptions      `json:"brandingOptions,omitempty"`
	ExpireTime         time.Time             `json:"expireTime"`
	CreateTime         time.Time             `json:"createTime"`
	OwnershipGraph     *OwnershipGraph       `json:"ownershipGraph,omitempty"`
	Metadata           Metadata              `json:"metadata,omitempty"`
}

// CheckoutLineItems contains items displayed by a checkout session.
type CheckoutLineItems struct {
	Data []CheckoutLineItem `json:"data"`
}

// CheckoutLineItem is an item returned for a checkout session.
type CheckoutLineItem struct {
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       Amount   `json:"price"`
	Quantity    int64    `json:"quantity"`
	Reference   *string  `json:"reference,omitempty"`
	Description *string  `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// CreateCheckoutLineItem is an item accepted when creating a checkout session.
type CreateCheckoutLineItem struct {
	Type        string   `json:"type,omitempty"`
	Name        string   `json:"name"`
	Price       Amount   `json:"price"`
	Quantity    *int64   `json:"quantity,omitempty"`
	Reference   *string  `json:"reference,omitempty"`
	Description *string  `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// BrandingOptions controls checkout presentation.
type BrandingOptions struct {
	PrimaryColor *string `json:"primaryColor,omitempty"`
}

// CardPaymentOptions controls card availability.
type CardPaymentOptions struct {
	Disable *bool `json:"disable,omitempty"`
}

// ProviderPaymentOptions controls a provider-based payment method.
type ProviderPaymentOptions struct {
	Disable           *bool    `json:"disable,omitempty"`
	EnabledProviders  []string `json:"enabledProviders,omitempty"`
	DisabledProviders []string `json:"disabledProviders,omitempty"`
}

// CheckoutPaymentOptions controls payment methods shown at checkout.
type CheckoutPaymentOptions struct {
	Card   *CardPaymentOptions     `json:"card,omitempty"`
	Bank   *ProviderPaymentOptions `json:"bank,omitempty"`
	Momo   *ProviderPaymentOptions `json:"momo,omitempty"`
	Wallet *ProviderPaymentOptions `json:"wallet,omitempty"`
}

// CreateCheckoutSessionInput contains fields accepted when creating a checkout session.
type CreateCheckoutSessionInput struct {
	Name               string                   `json:"name"`
	LineItems          []CreateCheckoutLineItem `json:"lineItems"`
	Description        string                   `json:"description,omitempty"`
	CancelURL          string                   `json:"cancelUrl,omitempty"`
	SuccessURL         string                   `json:"successUrl,omitempty"`
	CallbackState      string                   `json:"callbackState,omitempty"`
	Reference          string                   `json:"reference,omitempty"`
	FinancialAccountID string                   `json:"financialAccountId,omitempty"`
	PaymentOptions     *CheckoutPaymentOptions  `json:"paymentOptions,omitempty"`
	BrandingOptions    *BrandingOptions         `json:"brandingOptions,omitempty"`
	Metadata           Metadata                 `json:"metadata,omitempty"`
}

// CheckoutSessionListParams controls checkout-session pagination.
type CheckoutSessionListParams struct {
	Limit int
	After string
}

// CreateCheckoutSession creates a hosted checkout session.
func (c *Client) CreateCheckoutSession(
	ctx context.Context,
	idempotencyKey string,
	input CreateCheckoutSessionInput,
) (*CheckoutSession, *Response, error) {
	if input.Name == "" {
		return nil, nil, newValidationError("Name", "must be non-empty")
	}
	if len(input.LineItems) == 0 {
		return nil, nil, newValidationError("LineItems", "must contain at least one item")
	}
	var session CheckoutSession
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/checkout-sessions",
		input:          input,
		output:         &session,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &session, &response, nil
}

// GetCheckoutSession returns a checkout session by ID.
func (c *Client) GetCheckoutSession(ctx context.Context, id string) (*CheckoutSession, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var session CheckoutSession
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/checkout-sessions/" + url.PathEscape(id),
		output: &session,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &session, &response, nil
}

// ListCheckoutSessions returns checkout sessions.
func (c *Client) ListCheckoutSessions(
	ctx context.Context,
	params CheckoutSessionListParams,
) (*Page[CheckoutSession], *Response, error) {
	page, response, err := doList[CheckoutSession](c, ctx, "/checkout-sessions", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// DeleteCheckoutSession deletes a checkout session that has not started.
func (c *Client) DeleteCheckoutSession(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/checkout-sessions/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func (p CheckoutSessionListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	return values
}
