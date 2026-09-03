package monime

import (
	"context"
	"net/url"
	"strconv"
)

type PaymentCode struct {
	ID                    string            `json:"id"`
	Mode                  string            `json:"mode"`
	Status                string            `json:"status"`
	Name                  string            `json:"name"`
	Amount                Amount            `json:"amount"`
	Enable                bool              `json:"enable"`
	ExpireTime            string            `json:"expireTime"`
	USSDCode              string            `json:"ussdCode"`
	Reference             string            `json:"reference"`
	AuthorizedProviders   []string          `json:"authorizedProviders"`
	AuthorizedPhoneNumber string            `json:"authorizedPhoneNumber"`
	FinancialAccountID    string            `json:"financialAccountId"`
	Metadata              map[string]string `json:"metadata"`
}
type RecurrentPaymentTarget struct {
	ExpectedPaymentCount *int64  `json:"expectedPaymentCount,omitempty"`
	ExpectedPaymentTotal *Amount `json:"expectedPaymentTotal,omitempty"`
}
type CreatePaymentCodeInput struct {
	Mode                   string                  `json:"mode,omitempty"`
	Name                   string                  `json:"name"`
	Enable                 *bool                   `json:"enable,omitempty"`
	Amount                 *Amount                 `json:"amount,omitempty"`
	Duration               string                  `json:"duration,omitempty"`
	Reference              *string                 `json:"reference,omitempty"`
	AuthorizedProviders    []string                `json:"authorizedProviders,omitempty"`
	AuthorizedPhoneNumber  *string                 `json:"authorizedPhoneNumber,omitempty"`
	RecurrentPaymentTarget *RecurrentPaymentTarget `json:"recurrentPaymentTarget,omitempty"`
	FinancialAccountID     *string                 `json:"financialAccountId,omitempty"`
	Metadata               map[string]string       `json:"metadata,omitempty"`
}
type UpdatePaymentCodeInput struct {
	Name      Field[string]            `json:"name,omitzero"`
	Enable    Field[bool]              `json:"enable,omitzero"`
	Reference Field[string]            `json:"reference,omitzero"`
	Metadata  Field[map[string]string] `json:"metadata,omitzero"`
}
type PaymentCodeListQuery struct {
	Limit                         int
	After, Mode, USSDCode, Status string
}
type PaymentCodeService struct{ client *Client }

func (s *PaymentCodeService) Create(ctx context.Context, input CreatePaymentCodeInput) (PaymentCode, Response, error) {
	var r PaymentCode
	res, e := s.client.do(ctx, "POST", "/payment-codes", nil, input, &r)
	return r, res, e
}
func (s *PaymentCodeService) Get(ctx context.Context, id string) (PaymentCode, Response, error) {
	var r PaymentCode
	res, e := s.client.do(ctx, "GET", "/payment-codes/"+url.PathEscape(id), nil, nil, &r)
	return r, res, e
}
func (s *PaymentCodeService) List(ctx context.Context, q PaymentCodeListQuery) (Page[PaymentCode], Response, error) {
	var r []PaymentCode
	res, e := s.client.do(ctx, "GET", "/payment-codes", q.values(), nil, &r)
	return Page[PaymentCode]{Items: r}, res, e
}
func (s *PaymentCodeService) Update(ctx context.Context, id string, input UpdatePaymentCodeInput) (PaymentCode, Response, error) {
	var r PaymentCode
	res, e := s.client.do(ctx, "PATCH", "/payment-codes/"+url.PathEscape(id), nil, input, &r)
	return r, res, e
}
func (s *PaymentCodeService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/payment-codes/"+url.PathEscape(id), nil, nil, nil)
}
func (q PaymentCodeListQuery) values() url.Values {
	v := url.Values{}
	if q.Limit != 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		v.Set("after", q.After)
	}
	if q.Mode != "" {
		v.Set("mode", q.Mode)
	}
	if q.USSDCode != "" {
		v.Set("ussdCode", q.USSDCode)
	}
	if q.Status != "" {
		v.Set("status", q.Status)
	}
	return v
}
