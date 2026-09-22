package monime

import (
	"context"
	"net/url"
)

// ReceiptStatus is the redemption progress of a receipt.
type ReceiptStatus string

const (
	ReceiptNotRedeemed       ReceiptStatus = "not_redeemed"
	ReceiptPartiallyRedeemed ReceiptStatus = "partially_redeemed"
	ReceiptFullyRedeemed     ReceiptStatus = "fully_redeemed"
)

// Receipt is digital proof of purchase with redeemable entitlements.
type Receipt struct {
	Status       ReceiptStatus        `json:"status"`
	OrderName    *string              `json:"orderName,omitempty"`
	OrderNumber  string               `json:"orderNumber"`
	OrderAmount  *Amount              `json:"orderAmount,omitempty"`
	CreateTime   string               `json:"createTime"`
	UpdateTime   *string              `json:"updateTime,omitempty"`
	Entitlements []ReceiptEntitlement `json:"entitlements,omitempty"`
	Metadata     Metadata             `json:"metadata,omitempty"`
}

type ReceiptEntitlement struct {
	Key       string  `json:"key"`
	Name      *string `json:"name,omitempty"`
	Limit     int64   `json:"limit"`
	Current   int64   `json:"current"`
	Remaining int64   `json:"remaining"`
	Exhausted bool    `json:"exhausted"`
}

type RedeemReceiptInput struct {
	RedeemAll    *bool                    `json:"redeemAll,omitempty"`
	Entitlements []RedeemEntitlementInput `json:"entitlements,omitempty"`
	Metadata     Metadata                 `json:"metadata,omitempty"`
}

type RedeemEntitlementInput struct {
	Key   string `json:"key"`
	Units *int64 `json:"units,omitempty"`
}

type RedeemReceiptResult struct {
	Redeem  bool    `json:"redeem"`
	Receipt Receipt `json:"receipt"`
}

type ReceiptService struct{ client *Client }

func (s *ReceiptService) Get(ctx context.Context, orderNumber string) (Receipt, Response, error) {
	if orderNumber == "" {
		return Receipt{}, Response{}, newValidationError("OrderNumber", "must be non-empty")
	}
	var result Receipt
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/receipts/" + url.PathEscape(orderNumber),
		output: &result,
	})
	return result, response, err
}

func (s *ReceiptService) Redeem(
	ctx context.Context,
	orderNumber string,
	input RedeemReceiptInput,
) (RedeemReceiptResult, Response, error) {
	if orderNumber == "" {
		return RedeemReceiptResult{}, Response{}, newValidationError("OrderNumber", "must be non-empty")
	}
	redeemAll := input.RedeemAll != nil && *input.RedeemAll
	if !redeemAll && len(input.Entitlements) == 0 {
		return RedeemReceiptResult{}, Response{}, newValidationError(
			"RedeemReceiptInput",
			"must redeem all or select entitlements",
		)
	}
	if redeemAll && len(input.Entitlements) > 0 {
		return RedeemReceiptResult{}, Response{}, newValidationError(
			"RedeemReceiptInput",
			"must not combine redeem-all with selected entitlements",
		)
	}
	var result RedeemReceiptResult
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/receipts/" + url.PathEscape(orderNumber) + "/redeem",
		input:  input,
		output: &result,
	})
	return result, response, err
}
