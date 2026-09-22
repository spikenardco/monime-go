package monime

import (
	"context"
	"net/url"
	"strconv"
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
	OrderNumber        string                `json:"orderNumber"`
	Reference          *string               `json:"reference,omitempty"`
	Description        *string               `json:"description,omitempty"`
	RedirectURL        string                `json:"redirectUrl"`
	CancelURL          *string               `json:"cancelUrl,omitempty"`
	SuccessURL         *string               `json:"successUrl,omitempty"`
	LineItems          CheckoutLineItems     `json:"lineItems"`
	FinancialAccountID *string               `json:"financialAccountId,omitempty"`
	BrandingOptions    *BrandingOptions      `json:"brandingOptions,omitempty"`
	ExpireTime         string                `json:"expireTime"`
	CreateTime         string                `json:"createTime"`
	OwnershipGraph     *OwnershipGraph       `json:"ownershipGraph,omitempty"`
	Metadata           Metadata              `json:"metadata,omitempty"`
}

type CheckoutLineItems struct {
	Data []CheckoutLineItem `json:"data"`
}

type CheckoutLineItem struct {
	Type        string   `json:"type,omitempty"`
	Name        string   `json:"name"`
	Price       Amount   `json:"price"`
	Quantity    *int64   `json:"quantity,omitempty"`
	Reference   string   `json:"reference,omitempty"`
	Description string   `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}

type BrandingOptions struct {
	PrimaryColor *string `json:"primaryColor,omitempty"`
}

type PaymentMethodOptions struct {
	Disable           *bool    `json:"disable,omitempty"`
	EnabledProviders  []string `json:"enabledProviders,omitempty"`
	DisabledProviders []string `json:"disabledProviders,omitempty"`
}

type CheckoutPaymentOptions struct {
	Card   *PaymentMethodOptions `json:"card,omitempty"`
	Bank   *PaymentMethodOptions `json:"bank,omitempty"`
	Momo   *PaymentMethodOptions `json:"momo,omitempty"`
	Wallet *PaymentMethodOptions `json:"wallet,omitempty"`
}

type CreateCheckoutSessionInput struct {
	Name               string                  `json:"name"`
	LineItems          []CheckoutLineItem      `json:"lineItems"`
	Description        string                  `json:"description,omitempty"`
	CancelURL          string                  `json:"cancelUrl,omitempty"`
	SuccessURL         string                  `json:"successUrl,omitempty"`
	CallbackState      string                  `json:"callbackState,omitempty"`
	Reference          string                  `json:"reference,omitempty"`
	FinancialAccountID string                  `json:"financialAccountId,omitempty"`
	PaymentOptions     *CheckoutPaymentOptions `json:"paymentOptions,omitempty"`
	BrandingOptions    *BrandingOptions        `json:"brandingOptions,omitempty"`
	Metadata           Metadata                `json:"metadata,omitempty"`
}

type CheckoutSessionListQuery struct {
	Limit  int
	After  string
	Status CheckoutSessionStatus
}

type CheckoutSessionService struct{ client *Client }

func (s *CheckoutSessionService) Create(
	ctx context.Context,
	input CreateCheckoutSessionInput,
) (CheckoutSession, Response, error) {
	if input.Name == "" {
		return CheckoutSession{}, Response{}, newValidationError("Name", "must be non-empty")
	}
	if len(input.LineItems) == 0 {
		return CheckoutSession{}, Response{}, newValidationError("LineItems", "must contain at least one item")
	}
	var result CheckoutSession
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/checkout-sessions",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *CheckoutSessionService) Get(ctx context.Context, id string) (CheckoutSession, Response, error) {
	if id == "" {
		return CheckoutSession{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result CheckoutSession
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/checkout-sessions/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *CheckoutSessionService) List(
	ctx context.Context,
	query CheckoutSessionListQuery,
) (Page[CheckoutSession], Response, error) {
	return doList[CheckoutSession](s.client, ctx, "/checkout-sessions", query.values())
}

func (s *CheckoutSessionService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/checkout-sessions/" + url.PathEscape(id),
	})
}

func (q CheckoutSessionListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.Status != "" {
		values.Set("status", string(q.Status))
	}
	return values
}
