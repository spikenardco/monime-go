package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	CreateTime                    time.Time                      `json:"createTime"`
	UpdateTime                    *time.Time                     `json:"updateTime,omitempty"`
	OwnershipGraph                *OwnershipGraph                `json:"ownershipGraph,omitempty"`
	Metadata                      Metadata                       `json:"metadata,omitempty"`
}

// FinancialAccountRef identifies a financial account in a transfer.
type FinancialAccountRef struct {
	ID string `json:"id"`
}

// InternalTransferFailureDetail describes a failed transfer.
type InternalTransferFailureDetail struct {
	Code    string  `json:"code"`
	Message *string `json:"message,omitempty"`
}

// CreateInternalTransferInput contains fields accepted when creating a transfer.
type CreateInternalTransferInput struct {
	Amount                      Amount              `json:"amount"`
	SourceFinancialAccount      FinancialAccountRef `json:"sourceFinancialAccount"`
	DestinationFinancialAccount FinancialAccountRef `json:"destinationFinancialAccount"`
	Description                 string              `json:"description,omitempty"`
	Metadata                    Metadata            `json:"metadata,omitempty"`
}

// UpdateInternalTransferInput contains mutable transfer fields.
type UpdateInternalTransferInput struct {
	Description Field[string]   `json:"description,omitzero"`
	Metadata    Field[Metadata] `json:"metadata,omitzero"`
}

// InternalTransferListParams filters an internal-transfer list request.
type InternalTransferListParams struct {
	Limit                         int
	After                         string
	Status                        InternalTransferStatus
	SourceFinancialAccountID      string
	DestinationFinancialAccountID string
	FinancialTransactionReference string
}

// CreateInternalTransfer creates an internal transfer.
func (c *Client) CreateInternalTransfer(
	ctx context.Context,
	idempotencyKey string,
	input CreateInternalTransferInput,
) (*InternalTransfer, *Response, error) {
	if input.Amount.Value <= 0 {
		return nil, nil, newValidationError("Amount.Value", "must be positive")
	}
	if input.SourceFinancialAccount.ID == "" {
		return nil, nil, newValidationError("SourceFinancialAccount.ID", "must be non-empty")
	}
	if input.DestinationFinancialAccount.ID == "" {
		return nil, nil, newValidationError("DestinationFinancialAccount.ID", "must be non-empty")
	}
	if input.SourceFinancialAccount.ID == input.DestinationFinancialAccount.ID {
		return nil, nil, newValidationError(
			"DestinationFinancialAccount.ID",
			"must differ from SourceFinancialAccount.ID",
		)
	}
	var transfer InternalTransfer
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/internal-transfers",
		input:          input,
		output:         &transfer,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &transfer, &response, nil
}

// GetInternalTransfer returns an internal transfer by ID.
func (c *Client) GetInternalTransfer(
	ctx context.Context,
	id string,
) (*InternalTransfer, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var transfer InternalTransfer
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/internal-transfers/" + url.PathEscape(id),
		output: &transfer,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &transfer, &response, nil
}

// ListInternalTransfers returns internal transfers matching params.
func (c *Client) ListInternalTransfers(
	ctx context.Context,
	params InternalTransferListParams,
) (*Page[InternalTransfer], *Response, error) {
	page, response, err := doList[InternalTransfer](c, ctx, "/internal-transfers", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdateInternalTransfer updates mutable fields on a transfer.
func (c *Client) UpdateInternalTransfer(
	ctx context.Context,
	id string,
	input UpdateInternalTransferInput,
) (*InternalTransfer, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !input.Description.IsSet() && !input.Metadata.IsSet() {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var transfer InternalTransfer
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/internal-transfers/" + url.PathEscape(id),
		input:  input,
		output: &transfer,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &transfer, &response, nil
}

// DeleteInternalTransfer deletes a transfer in a deletable state.
func (c *Client) DeleteInternalTransfer(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/internal-transfers/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func (p InternalTransferListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	if p.Status != "" {
		values.Set("status", string(p.Status))
	}
	if p.SourceFinancialAccountID != "" {
		values.Set("sourceFinancialAccountId", p.SourceFinancialAccountID)
	}
	if p.DestinationFinancialAccountID != "" {
		values.Set("destinationFinancialAccountId", p.DestinationFinancialAccountID)
	}
	if p.FinancialTransactionReference != "" {
		values.Set("financialTransactionReference", p.FinancialTransactionReference)
	}
	return values
}
