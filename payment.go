package monime

import (
	"context"
	"net/url"
	"strconv"
)

type Payment struct {
	ID                 string            `json:"id"`
	Status             string            `json:"status"`
	Amount             Amount            `json:"amount"`
	Name               string            `json:"name"`
	Reference          string            `json:"reference"`
	OrderNumber        string            `json:"orderNumber"`
	FinancialAccountID string            `json:"financialAccountId"`
	Metadata           map[string]string `json:"metadata"`
}
type UpdatePaymentInput struct {
	Name      Field[string]            `json:"name,omitzero"`
	Reference Field[string]            `json:"reference,omitzero"`
	Metadata  Field[map[string]string] `json:"metadata,omitzero"`
}
type PaymentListQuery struct {
	Limit         int
	After, Status string
}
type PaymentService struct{ client *Client }

func (s *PaymentService) Get(ctx context.Context, id string) (Payment, Response, error) {
	var r Payment
	res, e := s.client.do(ctx, "GET", "/payments/"+url.PathEscape(id), nil, nil, &r)
	return r, res, e
}
func (s *PaymentService) List(ctx context.Context, q PaymentListQuery) (Page[Payment], Response, error) {
	var r []Payment
	res, e := s.client.do(ctx, "GET", "/payments", q.values(), nil, &r)
	return Page[Payment]{Items: r}, res, e
}
func (s *PaymentService) Update(ctx context.Context, id string, input UpdatePaymentInput) (Payment, Response, error) {
	var r Payment
	res, e := s.client.do(ctx, "PATCH", "/payments/"+url.PathEscape(id), nil, input, &r)
	return r, res, e
}
func (q PaymentListQuery) values() url.Values {
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
