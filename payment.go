package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PaymentStatus is the processing state of a payment.
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
)

// Payment is a customer payment transaction.
type Payment struct {
	ID                            string          `json:"id"`
	Status                        PaymentStatus   `json:"status"`
	Amount                        Amount          `json:"amount"`
	Channel                       *Channel        `json:"channel,omitempty"`
	Name                          *string         `json:"name,omitempty"`
	Reference                     *string         `json:"reference,omitempty"`
	OrderNumber                   *string         `json:"orderNumber,omitempty"`
	FinancialAccountID            *string         `json:"financialAccountId,omitempty"`
	FinancialTransactionReference *string         `json:"financialTransactionReference,omitempty"`
	Fees                          []Fee           `json:"fees,omitempty"`
	CreateTime                    time.Time       `json:"createTime"`
	UpdateTime                    *time.Time      `json:"updateTime,omitempty"`
	OwnershipGraph                *OwnershipGraph `json:"ownershipGraph,omitempty"`
	Metadata                      Metadata        `json:"metadata,omitempty"`
}

// UpdatePaymentInput contains mutable payment fields.
type UpdatePaymentInput struct {
	Name     Field[string]   `json:"name,omitzero"`
	Metadata Field[Metadata] `json:"metadata,omitzero"`
}

// PaymentListParams filters a payment list request.
type PaymentListParams struct {
	Limit                         int
	After                         string
	OrderNumber                   string
	FinancialAccountID            string
	FinancialTransactionReference string
}

// GetPayment returns a payment by ID.
func (c *Client) GetPayment(ctx context.Context, id string) (*Payment, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var payment Payment
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/payments/" + url.PathEscape(id),
		output: &payment,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &payment, &response, nil
}

// ListPayments returns payments matching params.
func (c *Client) ListPayments(
	ctx context.Context,
	params PaymentListParams,
) (*Page[Payment], *Response, error) {
	page, response, err := doList[Payment](c, ctx, "/payments", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdatePayment updates mutable fields on a payment.
func (c *Client) UpdatePayment(
	ctx context.Context,
	id string,
	input UpdatePaymentInput,
) (*Payment, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !input.Name.IsSet() && !input.Metadata.IsSet() {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var payment Payment
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/payments/" + url.PathEscape(id),
		input:  input,
		output: &payment,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &payment, &response, nil
}

func (p PaymentListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	if p.OrderNumber != "" {
		values.Set("orderNumber", p.OrderNumber)
	}
	if p.FinancialAccountID != "" {
		values.Set("financialAccountId", p.FinancialAccountID)
	}
	if p.FinancialTransactionReference != "" {
		values.Set("financialTransactionReference", p.FinancialTransactionReference)
	}
	return values
}
