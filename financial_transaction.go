package monime

import (
	"context"
	"net/url"
	"strconv"
)

type FinancialTransaction struct {
	ID               string                      `json:"id"`
	Type             string                      `json:"type"`
	Amount           Amount                      `json:"amount"`
	Timestamp        string                      `json:"timestamp"`
	Reference        string                      `json:"reference"`
	FinancialAccount FinancialTransactionAccount `json:"financialAccount"`
	Metadata         map[string]string           `json:"metadata"`
}
type FinancialTransactionAccount struct {
	ID      string `json:"id"`
	Balance struct {
		After Amount `json:"after"`
	} `json:"balance"`
}
type FinancialTransactionListQuery struct {
	Limit              int
	After              string
	FinancialAccountID string
	Type               string
	Reference          string
}
type FinancialTransactionService struct{ client *Client }

func (s *FinancialTransactionService) Get(ctx context.Context, id string) (FinancialTransaction, Response, error) {
	var result FinancialTransaction
	response, err := s.client.do(ctx, "GET", "/financial-transactions/"+url.PathEscape(id), nil, nil, &result)
	return result, response, err
}
func (s *FinancialTransactionService) List(ctx context.Context, query FinancialTransactionListQuery) (Page[FinancialTransaction], Response, error) {
	var items []FinancialTransaction
	response, err := s.client.do(ctx, "GET", "/financial-transactions", query.values(), nil, &items)
	return Page[FinancialTransaction]{Items: items}, response, err
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
		values.Set("type", q.Type)
	}
	if q.Reference != "" {
		values.Set("reference", q.Reference)
	}
	return values
}
