package monime

import (
	"context"
	"net/url"
	"strconv"
)

// PayoutStatus is the processing state of a payout.
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "pending"
	PayoutStatusProcessing PayoutStatus = "processing"
	PayoutStatusCompleted  PayoutStatus = "completed"
	PayoutStatusFailed     PayoutStatus = "failed"
)

// PayoutDestinationType identifies the recipient account kind.
type PayoutDestinationType string

const (
	PayoutDestinationBank   PayoutDestinationType = "bank"
	PayoutDestinationMomo   PayoutDestinationType = "momo"
	PayoutDestinationWallet PayoutDestinationType = "wallet"
)

// Payout is a disbursement from a financial account to an external recipient.
type Payout struct {
	ID             string               `json:"id"`
	Status         PayoutStatus         `json:"status"`
	Amount         Amount               `json:"amount"`
	Source         *PayoutSource        `json:"source,omitempty"`
	Destination    PayoutDestination    `json:"destination"`
	Fees           []Fee                `json:"fees,omitempty"`
	FailureDetail  *PayoutFailureDetail `json:"failureDetail,omitempty"`
	CreateTime     string               `json:"createTime"`
	UpdateTime     *string              `json:"updateTime,omitempty"`
	OwnershipGraph *OwnershipGraph      `json:"ownershipGraph,omitempty"`
	Metadata       Metadata             `json:"metadata,omitempty"`
}

type PayoutSource struct {
	FinancialAccountID   string  `json:"financialAccountId"`
	TransactionReference *string `json:"transactionReference,omitempty"`
}

type PayoutDestination struct {
	Type          PayoutDestinationType `json:"type"`
	ProviderID    string                `json:"providerId"`
	AccountNumber string                `json:"accountNumber,omitempty"`
	PhoneNumber   string                `json:"phoneNumber,omitempty"`
	WalletID      string                `json:"walletId,omitempty"`
}

type PayoutFailureDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreatePayoutInput struct {
	Amount      Amount            `json:"amount"`
	Destination PayoutDestination `json:"destination"`
	Source      *PayoutSource     `json:"source,omitempty"`
	Metadata    Metadata          `json:"metadata,omitempty"`
}

type UpdatePayoutInput struct {
	Metadata Field[Metadata] `json:"metadata,omitzero"`
}

type PayoutListQuery struct {
	Limit                           int
	After                           string
	Status                          PayoutStatus
	SourceFinancialAccountID        string
	SourceTransactionReference      string
	DestinationTransactionReference string
}

type PayoutService struct{ client *Client }

func (s *PayoutService) Create(ctx context.Context, input CreatePayoutInput) (Payout, Response, error) {
	if input.Destination.Type == "" {
		return Payout{}, Response{}, newValidationError("Destination.Type", "must be non-empty")
	}
	if input.Destination.ProviderID == "" {
		return Payout{}, Response{}, newValidationError("Destination.ProviderID", "must be non-empty")
	}
	var result Payout
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/payouts",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *PayoutService) Get(ctx context.Context, id string) (Payout, Response, error) {
	if id == "" {
		return Payout{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Payout
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/payouts/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *PayoutService) List(ctx context.Context, query PayoutListQuery) (Page[Payout], Response, error) {
	return doList[Payout](s.client, ctx, "/payouts", query.values())
}

func (s *PayoutService) Update(ctx context.Context, id string, input UpdatePayoutInput) (Payout, Response, error) {
	if id == "" {
		return Payout{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result Payout
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/payouts/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *PayoutService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/payouts/" + url.PathEscape(id),
	})
}

func (q PayoutListQuery) values() url.Values {
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
	if q.SourceFinancialAccountID != "" {
		values.Set("sourceFinancialAccountId", q.SourceFinancialAccountID)
	}
	if q.SourceTransactionReference != "" {
		values.Set("sourceTransactionReference", q.SourceTransactionReference)
	}
	if q.DestinationTransactionReference != "" {
		values.Set("destinationTransactionReference", q.DestinationTransactionReference)
	}
	return values
}
