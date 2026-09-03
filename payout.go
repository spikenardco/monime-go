package monime

import (
	"context"
	"net/url"
	"strconv"
)

type Payout struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Amount      Amount            `json:"amount"`
	Destination PayoutDestination `json:"destination"`
	Metadata    map[string]string `json:"metadata"`
}
type PayoutDestination struct {
	Type        string `json:"type"`
	ProviderID  string `json:"providerId"`
	PhoneNumber string `json:"phoneNumber"`
	AccountID   string `json:"accountId"`
}
type CreatePayoutInput struct {
	Amount                   Amount            `json:"amount"`
	Destination              PayoutDestination `json:"destination"`
	SourceFinancialAccountID string            `json:"sourceFinancialAccountId,omitempty"`
	Metadata                 map[string]string `json:"metadata,omitempty"`
}
type UpdatePayoutInput struct {
	Metadata Field[map[string]string] `json:"metadata,omitzero"`
}
type PayoutListQuery struct {
	Limit         int
	After, Status string
}
type PayoutService struct{ client *Client }

func (s *PayoutService) Create(ctx context.Context, i CreatePayoutInput) (Payout, Response, error) {
	var r Payout
	x, e := s.client.do(ctx, "POST", "/payouts", nil, i, &r)
	return r, x, e
}
func (s *PayoutService) Get(ctx context.Context, id string) (Payout, Response, error) {
	var r Payout
	x, e := s.client.do(ctx, "GET", "/payouts/"+url.PathEscape(id), nil, nil, &r)
	return r, x, e
}
func (s *PayoutService) List(ctx context.Context, q PayoutListQuery) (Page[Payout], Response, error) {
	var r []Payout
	x, e := s.client.do(ctx, "GET", "/payouts", q.values(), nil, &r)
	return Page[Payout]{Items: r}, x, e
}
func (s *PayoutService) Update(ctx context.Context, id string, i UpdatePayoutInput) (Payout, Response, error) {
	var r Payout
	x, e := s.client.do(ctx, "PATCH", "/payouts/"+url.PathEscape(id), nil, i, &r)
	return r, x, e
}
func (s *PayoutService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/payouts/"+url.PathEscape(id), nil, nil, nil)
}
func (q PayoutListQuery) values() url.Values {
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
