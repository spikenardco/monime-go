package monime

import (
	"context"
	"net/http"
	"net/url"
	"time"
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
	OrderAmount  Amount               `json:"orderAmount"`
	CreateTime   time.Time            `json:"createTime"`
	UpdateTime   time.Time            `json:"updateTime"`
	Entitlements []ReceiptEntitlement `json:"entitlements,omitempty"`
	Metadata     Metadata             `json:"metadata,omitempty"`
}

// ReceiptEntitlement is a right granted by a receipt that can be redeemed.
type ReceiptEntitlement struct {
	Key       string  `json:"key"`
	Name      *string `json:"name,omitempty"`
	Limit     int64   `json:"limit"`
	Current   int64   `json:"current"`
	Remaining int64   `json:"remaining"`
	Exhausted bool    `json:"exhausted"`
}

// RedeemReceiptInput selects the receipt entitlements to redeem.
type RedeemReceiptInput struct {
	RedeemAll    *bool                    `json:"redeemAll,omitempty"`
	Entitlements []RedeemEntitlementInput `json:"entitlements,omitempty"`
	Metadata     Metadata                 `json:"metadata,omitempty"`
}

// RedeemEntitlementInput selects an entitlement and number of units.
type RedeemEntitlementInput struct {
	Key   string `json:"key"`
	Units *int64 `json:"units,omitempty"`
}

// RedeemReceiptResult is the outcome of a receipt redemption.
type RedeemReceiptResult struct {
	Redeem  bool    `json:"redeem"`
	Receipt Receipt `json:"receipt"`
}

// GetReceipt returns a receipt by order number.
func (c *Client) GetReceipt(ctx context.Context, orderNumber string) (*Receipt, *Response, error) {
	if orderNumber == "" {
		return nil, nil, newValidationError("OrderNumber", "must be non-empty")
	}
	var receipt Receipt
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/receipts/" + url.PathEscape(orderNumber),
		output: &receipt,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &receipt, &response, nil
}

// RedeemReceipt atomically redeems selected or all entitlements from a receipt.
func (c *Client) RedeemReceipt(
	ctx context.Context,
	idempotencyKey string,
	orderNumber string,
	input RedeemReceiptInput,
) (*RedeemReceiptResult, *Response, error) {
	if orderNumber == "" {
		return nil, nil, newValidationError("OrderNumber", "must be non-empty")
	}
	redeemAll := input.RedeemAll != nil && *input.RedeemAll
	if !redeemAll && len(input.Entitlements) == 0 {
		return nil, nil, newValidationError("Input", "must redeem all or select entitlements")
	}
	if redeemAll && len(input.Entitlements) > 0 {
		return nil, nil, newValidationError("Input", "must not combine redeem-all with selected entitlements")
	}
	var result RedeemReceiptResult
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/receipts/" + url.PathEscape(orderNumber) + "/redeem",
		input:          input,
		output:         &result,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &result, &response, nil
}
