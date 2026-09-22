package monime

import (
	"context"
	"net/url"
	"strconv"
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
	CreateTime                    string          `json:"createTime"`
	UpdateTime                    *string         `json:"updateTime,omitempty"`
	OwnershipGraph                *OwnershipGraph `json:"ownershipGraph,omitempty"`
	Metadata                      Metadata        `json:"metadata,omitempty"`
}

type UpdatePaymentInput struct {
	Name     Field[string]   `json:"name,omitzero"`
	Metadata Field[Metadata] `json:"metadata,omitzero"`
}

type PaymentListQuery struct {
	Limit                         int
	After                         string
	OrderNumber                   string
	FinancialAccountID            string
	FinancialTransactionReference string
}

type PaymentService struct{ client *Client }

func (s *PaymentService) Get(ctx context.Context, id string) (Payment, Response, error) {
	if id == "" {
		return Payment{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Payment
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/payments/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *PaymentService) List(ctx context.Context, query PaymentListQuery) (Page[Payment], Response, error) {
	return doList[Payment](s.client, ctx, "/payments", query.values())
}

func (s *PaymentService) Update(ctx context.Context, id string, input UpdatePaymentInput) (Payment, Response, error) {
	if id == "" {
		return Payment{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Payment
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/payments/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (q PaymentListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.OrderNumber != "" {
		values.Set("orderNumber", q.OrderNumber)
	}
	if q.FinancialAccountID != "" {
		values.Set("financialAccountId", q.FinancialAccountID)
	}
	if q.FinancialTransactionReference != "" {
		values.Set("financialTransactionReference", q.FinancialTransactionReference)
	}
	return values
}
