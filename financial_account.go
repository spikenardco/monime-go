package monime

import (
	"context"
	"net/url"
	"strconv"
)

// FinancialAccount is a wallet that holds and tracks funds.
type FinancialAccount struct {
	ID          string                   `json:"id"`
	UVAN        string                   `json:"uvan"`
	Name        string                   `json:"name"`
	Currency    Currency                 `json:"currency"`
	Reference   *string                  `json:"reference,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Balance     *FinancialAccountBalance `json:"balance,omitempty"`
	CreateTime  string                   `json:"createTime"`
	UpdateTime  *string                  `json:"updateTime,omitempty"`
	Metadata    Metadata                 `json:"metadata,omitempty"`
}

type FinancialAccountBalance struct {
	Available Amount `json:"available"`
}

type CreateFinancialAccountInput struct {
	Name        string   `json:"name"`
	Currency    Currency `json:"currency"`
	Reference   *string  `json:"reference,omitempty"`
	Description *string  `json:"description,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
}

type UpdateFinancialAccountInput struct {
	Name        Field[string]   `json:"name,omitzero"`
	Reference   Field[string]   `json:"reference,omitzero"`
	Description Field[string]   `json:"description,omitzero"`
	Metadata    Field[Metadata] `json:"metadata,omitzero"`
}

type FinancialAccountGetQuery struct {
	WithBalance bool
}

type FinancialAccountListQuery struct {
	Limit       int
	After       string
	UVAN        string
	Reference   string
	WithBalance bool
}

type FinancialAccountService struct{ client *Client }

func (s *FinancialAccountService) Create(
	ctx context.Context,
	input CreateFinancialAccountInput,
) (FinancialAccount, Response, error) {
	if input.Name == "" {
		return FinancialAccount{}, Response{}, newValidationError("Name", "must be non-empty")
	}
	if input.Currency == "" {
		return FinancialAccount{}, Response{}, newValidationError("Currency", "must be non-empty")
	}
	var result FinancialAccount
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/financial-accounts",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *FinancialAccountService) Get(
	ctx context.Context,
	id string,
	query FinancialAccountGetQuery,
) (FinancialAccount, Response, error) {
	if id == "" {
		return FinancialAccount{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result FinancialAccount
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/financial-accounts/" + url.PathEscape(id),
		query:  query.values(),
		output: &result,
	})
	return result, response, err
}

func (s *FinancialAccountService) List(
	ctx context.Context,
	query FinancialAccountListQuery,
) (Page[FinancialAccount], Response, error) {
	return doList[FinancialAccount](s.client, ctx, "/financial-accounts", query.values())
}

func (s *FinancialAccountService) Update(
	ctx context.Context,
	id string,
	input UpdateFinancialAccountInput,
) (FinancialAccount, Response, error) {
	if id == "" {
		return FinancialAccount{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result FinancialAccount
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/financial-accounts/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (q FinancialAccountGetQuery) values() url.Values {
	values := url.Values{}
	if q.WithBalance {
		values.Set("withBalance", "true")
	}
	return values
}

func (q FinancialAccountListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.UVAN != "" {
		values.Set("uvan", q.UVAN)
	}
	if q.Reference != "" {
		values.Set("reference", q.Reference)
	}
	if q.WithBalance {
		values.Set("withBalance", "true")
	}
	return values
}
