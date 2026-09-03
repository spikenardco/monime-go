package monime

import (
	"context"
	"net/url"
	"strconv"
)

type CheckoutSession struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Name        string            `json:"name"`
	OrderNumber string            `json:"orderNumber"`
	RedirectURL string            `json:"redirectUrl"`
	CancelURL   string            `json:"cancelUrl"`
	SuccessURL  string            `json:"successUrl"`
	ExpireTime  string            `json:"expireTime"`
	Metadata    map[string]string `json:"metadata"`
}
type CheckoutLineItem struct {
	Name      string `json:"name"`
	Price     Amount `json:"price"`
	Quantity  int64  `json:"quantity,omitempty"`
	Reference string `json:"reference,omitempty"`
}
type CreateCheckoutSessionInput struct {
	Name               string             `json:"name"`
	LineItems          []CheckoutLineItem `json:"lineItems"`
	Description        string             `json:"description,omitempty"`
	CancelURL          string             `json:"cancelUrl,omitempty"`
	SuccessURL         string             `json:"successUrl,omitempty"`
	Reference          string             `json:"reference,omitempty"`
	FinancialAccountID string             `json:"financialAccountId,omitempty"`
	Metadata           map[string]string  `json:"metadata,omitempty"`
}
type CheckoutSessionListQuery struct {
	Limit         int
	After, Status string
}
type CheckoutSessionService struct{ client *Client }

func (s *CheckoutSessionService) Create(ctx context.Context, i CreateCheckoutSessionInput) (CheckoutSession, Response, error) {
	var r CheckoutSession
	x, e := s.client.do(ctx, "POST", "/checkout-sessions", nil, i, &r)
	return r, x, e
}
func (s *CheckoutSessionService) Get(ctx context.Context, id string) (CheckoutSession, Response, error) {
	var r CheckoutSession
	x, e := s.client.do(ctx, "GET", "/checkout-sessions/"+url.PathEscape(id), nil, nil, &r)
	return r, x, e
}
func (s *CheckoutSessionService) List(ctx context.Context, q CheckoutSessionListQuery) (Page[CheckoutSession], Response, error) {
	var r []CheckoutSession
	x, e := s.client.do(ctx, "GET", "/checkout-sessions", q.values(), nil, &r)
	return Page[CheckoutSession]{Items: r}, x, e
}
func (s *CheckoutSessionService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/checkout-sessions/"+url.PathEscape(id), nil, nil, nil)
}
func (q CheckoutSessionListQuery) values() url.Values {
	v := url.Values{}
	if q.Limit != 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		v.Set("after", q.After)
	}
	if q.Status != "" {
		v.Set("status", q.Status)
	}
	return v
}
