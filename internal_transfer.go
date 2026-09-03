package monime

import (
	"context"
	"net/url"
	"strconv"
)

// InternalTransfer is a transfer between financial accounts in a Space.
type InternalTransfer struct {
	ID                          string            `json:"id"`
	Status                      string            `json:"status"`
	Amount                      Amount            `json:"amount"`
	SourceFinancialAccount      TransferAccount   `json:"sourceFinancialAccount"`
	DestinationFinancialAccount TransferAccount   `json:"destinationFinancialAccount"`
	Description                 string            `json:"description"`
	Metadata                    map[string]string `json:"metadata"`
}

type TransferAccount struct {
	ID string `json:"id"`
}

type CreateInternalTransferInput struct {
	Amount                        Amount            `json:"amount"`
	SourceFinancialAccountID      string            `json:"sourceFinancialAccountId"`
	DestinationFinancialAccountID string            `json:"destinationFinancialAccountId"`
	Description                   string            `json:"description,omitempty"`
	Metadata                      map[string]string `json:"metadata,omitempty"`
}

type UpdateInternalTransferInput struct {
	Description Field[string]            `json:"description,omitzero"`
	Metadata    Field[map[string]string] `json:"metadata,omitzero"`
}

type InternalTransferListQuery struct {
	Limit                         int
	After                         string
	Status                        string
	SourceFinancialAccountID      string
	DestinationFinancialAccountID string
}

type InternalTransferService struct{ client *Client }

func (s *InternalTransferService) Create(ctx context.Context, input CreateInternalTransferInput) (InternalTransfer, Response, error) {
	if input.SourceFinancialAccountID == input.DestinationFinancialAccountID && input.SourceFinancialAccountID != "" {
		return InternalTransfer{}, Response{}, newConfigValidationError("DestinationFinancialAccountID", "must differ from SourceFinancialAccountID")
	}
	var result InternalTransfer
	response, err := s.client.do(ctx, "POST", "/internal-transfers", nil, input, &result)
	return result, response, err
}
func (s *InternalTransferService) Get(ctx context.Context, id string) (InternalTransfer, Response, error) {
	var result InternalTransfer
	response, err := s.client.do(ctx, "GET", "/internal-transfers/"+url.PathEscape(id), nil, nil, &result)
	return result, response, err
}
func (s *InternalTransferService) List(ctx context.Context, query InternalTransferListQuery) (Page[InternalTransfer], Response, error) {
	var items []InternalTransfer
	response, err := s.client.do(ctx, "GET", "/internal-transfers", query.values(), nil, &items)
	return Page[InternalTransfer]{Items: items}, response, err
}
func (s *InternalTransferService) Update(ctx context.Context, id string, input UpdateInternalTransferInput) (InternalTransfer, Response, error) {
	var result InternalTransfer
	response, err := s.client.do(ctx, "PATCH", "/internal-transfers/"+url.PathEscape(id), nil, input, &result)
	return result, response, err
}
func (s *InternalTransferService) Delete(ctx context.Context, id string) (Response, error) {
	return s.client.do(ctx, "DELETE", "/internal-transfers/"+url.PathEscape(id), nil, nil, nil)
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
		values.Set("status", q.Status)
	}
	return values
}
