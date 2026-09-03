package monime

import (
	"context"
	"net/url"
	"strconv"
)

type FinancialAccount struct {
	ID          string                  `json:"id"`
	UVAN        string                  `json:"uvan"`
	Name        string                  `json:"name"`
	Currency    Currency                `json:"currency"`
	Reference   string                  `json:"reference"`
	Description string                  `json:"description"`
	Balance     FinancialAccountBalance `json:"balance"`
	CreateTime  string                  `json:"createTime"`
	UpdateTime  string                  `json:"updateTime"`
	Metadata    map[string]string       `json:"metadata"`
}
type FinancialAccountBalance struct {
	Available Amount `json:"available"`
}
type CreateFinancialAccountInput struct {
	Name        string            `json:"name"`
	Currency    Currency          `json:"currency"`
	Reference   *string           `json:"reference,omitempty"`
	Description *string           `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
type UpdateFinancialAccountInput struct {
	Name        Field[string]            `json:"name,omitzero"`
	Reference   Field[string]            `json:"reference,omitzero"`
	Description Field[string]            `json:"description,omitzero"`
	Metadata    Field[map[string]string] `json:"metadata,omitzero"`
}
type FinancialAccountListQuery struct {
	Limit     int
	After     string
	Currency  Currency
	Reference string
}
type FinancialAccountService struct{ client *Client }

func (s *FinancialAccountService) Create(ctx context.Context, input CreateFinancialAccountInput) (FinancialAccount, Response, error) {
	var result FinancialAccount
	response, err := s.client.do(ctx, "POST", "/financial-accounts", nil, input, &result)
	return result, response, err
}
func (s *FinancialAccountService) Get(ctx context.Context, id string) (FinancialAccount, Response, error) {
	var result FinancialAccount
	response, err := s.client.do(ctx, "GET", "/financial-accounts/"+url.PathEscape(id), nil, nil, &result)
	return result, response, err
}
func (s *FinancialAccountService) List(ctx context.Context, query FinancialAccountListQuery) (Page[FinancialAccount], Response, error) {
	var items []FinancialAccount
	response, err := s.client.do(ctx, "GET", "/financial-accounts", query.values(), nil, &items)
	return Page[FinancialAccount]{Items: items}, response, err
}
func (s *FinancialAccountService) Update(ctx context.Context, id string, input UpdateFinancialAccountInput) (FinancialAccount, Response, error) {
	var result FinancialAccount
	response, err := s.client.do(ctx, "PATCH", "/financial-accounts/"+url.PathEscape(id), nil, input, &result)
	return result, response, err
}
func (q FinancialAccountListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.Currency != "" {
		values.Set("currency", string(q.Currency))
	}
	if q.Reference != "" {
		values.Set("reference", q.Reference)
	}
	return values
}
