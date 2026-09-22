package monime

import (
	"context"
	"net/url"
	"strconv"
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
	Timestamp           string                      `json:"timestamp"`
	Reference           *string                     `json:"reference,omitempty"`
	FinancialAccount    FinancialTransactionAccount `json:"financialAccount"`
	OriginatingReversal *OriginatingReversal        `json:"originatingReversal,omitempty"`
	OriginatingFee      *OriginatingFee             `json:"originatingFee,omitempty"`
	OwnershipGraph      *OwnershipGraph             `json:"ownershipGraph,omitempty"`
	Metadata            Metadata                    `json:"metadata,omitempty"`
}

type FinancialTransactionAccount struct {
	ID      string `json:"id"`
	Balance struct {
		After Amount `json:"after"`
	} `json:"balance"`
}

type OriginatingReversal struct {
	OriginTxnID  string `json:"originTxnId"`
	OriginTxnRef string `json:"originTxnRef"`
}

type OriginatingFee struct {
	Code string `json:"code"`
}

type FinancialTransactionListQuery struct {
	Limit              int
	After              string
	FinancialAccountID string
	Type               FinancialTransactionType
	Reference          string
}

type FinancialTransactionService struct{ client *Client }

func (s *FinancialTransactionService) Get(ctx context.Context, id string) (FinancialTransaction, Response, error) {
	if id == "" {
		return FinancialTransaction{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result FinancialTransaction
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/financial-transactions/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *FinancialTransactionService) List(
	ctx context.Context,
	query FinancialTransactionListQuery,
) (Page[FinancialTransaction], Response, error) {
	return doList[FinancialTransaction](s.client, ctx, "/financial-transactions", query.values())
}

func (q FinancialTransactionListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.FinancialAccountID != "" {
		values.Set("financialAccountId", q.FinancialAccountID)
	}
	if q.Type != "" {
		values.Set("type", string(q.Type))
	}
	if q.Reference != "" {
		values.Set("reference", q.Reference)
	}
	return values
}
