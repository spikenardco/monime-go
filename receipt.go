package monime

import (
	"context"
	"net/url"
)

type Receipt struct {
	Status       string               `json:"status"`
	OrderName    string               `json:"orderName"`
	OrderNumber  string               `json:"orderNumber"`
	OrderAmount  Amount               `json:"orderAmount"`
	Entitlements []ReceiptEntitlement `json:"entitlements"`
	Metadata     map[string]string    `json:"metadata"`
}
type ReceiptEntitlement struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Units int64  `json:"units"`
}
type RedeemReceiptInput struct {
	Entitlements []ReceiptEntitlementRedemption `json:"entitlements"`
}
type ReceiptEntitlementRedemption struct {
	ID    string `json:"id"`
	Units int64  `json:"units"`
}
type ReceiptService struct{ client *Client }

func (s *ReceiptService) Get(ctx context.Context, orderNumber string) (Receipt, Response, error) {
	var result Receipt
	response, err := s.client.do(ctx, "GET", "/receipts/"+url.PathEscape(orderNumber), nil, nil, &result)
	return result, response, err
}
func (s *ReceiptService) Redeem(ctx context.Context, orderNumber string, input RedeemReceiptInput) (Receipt, Response, error) {
	var result Receipt
	response, err := s.client.do(ctx, "POST", "/receipts/"+url.PathEscape(orderNumber)+"/redeem", nil, input, &result)
	return result, response, err
}
