package monime

import (
	"context"
	"net/url"
	"strconv"
)

// InternalTransferStatus is the processing state of an internal transfer.
type InternalTransferStatus string

const (
	InternalTransferPending    InternalTransferStatus = "pending"
	InternalTransferProcessing InternalTransferStatus = "processing"
	InternalTransferCompleted  InternalTransferStatus = "completed"
	InternalTransferFailed     InternalTransferStatus = "failed"
)

// InternalTransfer moves funds between financial accounts in a Space.
type InternalTransfer struct {
	ID                            string                         `json:"id"`
	Status                        InternalTransferStatus         `json:"status"`
	Amount                        Amount                         `json:"amount"`
	SourceFinancialAccount        FinancialAccountRef            `json:"sourceFinancialAccount"`
	DestinationFinancialAccount   FinancialAccountRef            `json:"destinationFinancialAccount"`
	FinancialTransactionReference *string                        `json:"financialTransactionReference,omitempty"`
	Description                   *string                        `json:"description,omitempty"`
	FailureDetail                 *InternalTransferFailureDetail `json:"failureDetail,omitempty"`
	CreateTime                    string                         `json:"createTime"`
	UpdateTime                    *string                        `json:"updateTime,omitempty"`
	OwnershipGraph                *OwnershipGraph                `json:"ownershipGraph,omitempty"`
	Metadata                      Metadata                       `json:"metadata,omitempty"`
}

type FinancialAccountRef struct {
	ID string `json:"id"`
}

type InternalTransferFailureDetail struct {
	Code    string  `json:"code"`
	Message *string `json:"message,omitempty"`
}

type CreateInternalTransferInput struct {
	Amount                      Amount              `json:"amount"`
	SourceFinancialAccount      FinancialAccountRef `json:"sourceFinancialAccount"`
	DestinationFinancialAccount FinancialAccountRef `json:"destinationFinancialAccount"`
	Description                 string              `json:"description,omitempty"`
	Metadata                    Metadata            `json:"metadata,omitempty"`
}

type UpdateInternalTransferInput struct {
	Description Field[string]   `json:"description,omitzero"`
	Metadata    Field[Metadata] `json:"metadata,omitzero"`
}

type InternalTransferListQuery struct {
	Limit                         int
	After                         string
	Status                        InternalTransferStatus
	SourceFinancialAccountID      string
	DestinationFinancialAccountID string
	FinancialTransactionReference string
}

type InternalTransferService struct{ client *Client }

func (s *InternalTransferService) Create(
	ctx context.Context,
	input CreateInternalTransferInput,
) (InternalTransfer, Response, error) {
	if input.SourceFinancialAccount.ID == "" {
		return InternalTransfer{}, Response{}, newValidationError("SourceFinancialAccount.ID", "must be non-empty")
	}
	if input.DestinationFinancialAccount.ID == "" {
		return InternalTransfer{}, Response{}, newValidationError("DestinationFinancialAccount.ID", "must be non-empty")
	}
	if input.SourceFinancialAccount.ID == input.DestinationFinancialAccount.ID {
		return InternalTransfer{}, Response{}, newValidationError(
			"DestinationFinancialAccount.ID",
			"must differ from SourceFinancialAccount.ID",
		)
	}
	var result InternalTransfer
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/internal-transfers",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *InternalTransferService) Get(ctx context.Context, id string) (InternalTransfer, Response, error) {
	if id == "" {
		return InternalTransfer{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result InternalTransfer
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/internal-transfers/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *InternalTransferService) List(
	ctx context.Context,
	query InternalTransferListQuery,
) (Page[InternalTransfer], Response, error) {
	return doList[InternalTransfer](s.client, ctx, "/internal-transfers", query.values())
}

func (s *InternalTransferService) Update(
	ctx context.Context,
	id string,
	input UpdateInternalTransferInput,
) (InternalTransfer, Response, error) {
	if id == "" {
		return InternalTransfer{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result InternalTransfer
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/internal-transfers/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *InternalTransferService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/internal-transfers/" + url.PathEscape(id),
	})
}

func (q InternalTransferListQuery) values() url.Values {
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
	if q.DestinationFinancialAccountID != "" {
		values.Set("destinationFinancialAccountId", q.DestinationFinancialAccountID)
	}
	if q.FinancialTransactionReference != "" {
		values.Set("financialTransactionReference", q.FinancialTransactionReference)
	}
	return values
}
