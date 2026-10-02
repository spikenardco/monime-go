package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FinancialTransactionType is the ledger direction of a transaction.
type FinancialTransactionType string

const (
	FinancialTransactionCredit FinancialTransactionType = "credit"
	FinancialTransactionDebit  FinancialTransactionType = "debit"
)

// FinancialTransaction is an immutable ledger entry for a fund movement.
type FinancialTransaction struct {
	ID                  string                      `json:"id"`
	Type                FinancialTransactionType    `json:"type"`
	Amount              Amount                      `json:"amount"`
	Timestamp           time.Time                   `json:"timestamp"`
	Reference           *string                     `json:"reference,omitempty"`
	FinancialAccount    FinancialTransactionAccount `json:"financialAccount"`
	OriginatingReversal *OriginatingReversal        `json:"originatingReversal,omitempty"`
	OriginatingFee      *OriginatingFee             `json:"originatingFee,omitempty"`
	OwnershipGraph      *OwnershipGraph             `json:"ownershipGraph,omitempty"`
	Metadata            Metadata                    `json:"metadata,omitempty"`
}

// FinancialTransactionAccount identifies the affected account and optional balance snapshot.
type FinancialTransactionAccount struct {
	ID      string                       `json:"id"`
	Balance *FinancialTransactionBalance `json:"balance,omitempty"`
}

// FinancialTransactionBalance is the account balance after a transaction.
type FinancialTransactionBalance struct {
	After Amount `json:"after"`
}

// OriginatingReversal links a reversal to the transaction it reversed.
type OriginatingReversal struct {
	OriginTxnID  string `json:"originTxnId"`
	OriginTxnRef string `json:"originTxnRef"`
}

// OriginatingFee identifies the fee that produced a transaction.
type OriginatingFee struct {
	Code string `json:"code"`
}

// FinancialTransactionListParams filters a transaction list request.
type FinancialTransactionListParams struct {
	Limit              int
	After              string
	FinancialAccountID string
	Type               FinancialTransactionType
	Reference          string
}

// GetFinancialTransaction returns a financial transaction by ID.
func (c *Client) GetFinancialTransaction(
	ctx context.Context,
	id string,
) (*FinancialTransaction, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var transaction FinancialTransaction
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/financial-transactions/" + url.PathEscape(id),
		output: &transaction,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &transaction, &response, nil
}

// ListFinancialTransactions returns transactions matching params.
func (c *Client) ListFinancialTransactions(
	ctx context.Context,
	params FinancialTransactionListParams,
) (*Page[FinancialTransaction], *Response, error) {
	page, response, err := doList[FinancialTransaction](c, ctx, "/financial-transactions", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

func (p FinancialTransactionListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	if p.FinancialAccountID != "" {
		values.Set("financialAccountId", p.FinancialAccountID)
	}
	if p.Type != "" {
		values.Set("type", string(p.Type))
	}
	if p.Reference != "" {
		values.Set("reference", p.Reference)
	}
	return values
}
