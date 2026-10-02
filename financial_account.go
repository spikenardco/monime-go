package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
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
	CreateTime  time.Time                `json:"createTime"`
	UpdateTime  *time.Time               `json:"updateTime,omitempty"`
	Metadata    Metadata                 `json:"metadata,omitempty"`
}

// FinancialAccountBalance contains the account's available funds.
type FinancialAccountBalance struct {
	Available Amount `json:"available"`
}

// CreateFinancialAccountInput contains fields accepted when creating an account.
type CreateFinancialAccountInput struct {
	Name        string   `json:"name"`
	Currency    Currency `json:"currency"`
	Reference   *string  `json:"reference,omitempty"`
	Description *string  `json:"description,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
}

// UpdateFinancialAccountInput contains mutable account fields.
type UpdateFinancialAccountInput struct {
	Name        Field[string]   `json:"name,omitzero"`
	Reference   Field[string]   `json:"reference,omitzero"`
	Description Field[string]   `json:"description,omitzero"`
	Metadata    Field[Metadata] `json:"metadata,omitzero"`
}

// FinancialAccountGetParams controls optional account response fields.
type FinancialAccountGetParams struct {
	WithBalance bool
}

// FinancialAccountListParams filters an account list request.
type FinancialAccountListParams struct {
	Limit       int
	After       string
	UVAN        string
	Reference   string
	WithBalance bool
}

// CreateFinancialAccount creates a financial account.
func (c *Client) CreateFinancialAccount(
	ctx context.Context,
	idempotencyKey string,
	input CreateFinancialAccountInput,
) (*FinancialAccount, *Response, error) {
	if input.Name == "" {
		return nil, nil, newValidationError("Name", "must be non-empty")
	}
	if input.Currency == "" {
		return nil, nil, newValidationError("Currency", "must be non-empty")
	}
	var account FinancialAccount
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/financial-accounts",
		input:          input,
		output:         &account,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &account, &response, nil
}

// GetFinancialAccount returns a financial account by ID.
func (c *Client) GetFinancialAccount(
	ctx context.Context,
	id string,
	params FinancialAccountGetParams,
) (*FinancialAccount, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var account FinancialAccount
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/financial-accounts/" + url.PathEscape(id),
		query:  params.values(),
		output: &account,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &account, &response, nil
}

// ListFinancialAccounts returns financial accounts matching params.
func (c *Client) ListFinancialAccounts(
	ctx context.Context,
	params FinancialAccountListParams,
) (*Page[FinancialAccount], *Response, error) {
	page, response, err := doList[FinancialAccount](c, ctx, "/financial-accounts", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdateFinancialAccount updates mutable fields on a financial account.
func (c *Client) UpdateFinancialAccount(
	ctx context.Context,
	id string,
	input UpdateFinancialAccountInput,
) (*FinancialAccount, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !input.Name.IsSet() && !input.Reference.IsSet() && !input.Description.IsSet() && !input.Metadata.IsSet() {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var account FinancialAccount
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/financial-accounts/" + url.PathEscape(id),
		input:  input,
		output: &account,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &account, &response, nil
}

func (p FinancialAccountGetParams) values() url.Values {
	values := url.Values{}
	if p.WithBalance {
		values.Set("withBalance", "true")
	}
	return values
}

func (p FinancialAccountListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	if p.UVAN != "" {
		values.Set("uvan", p.UVAN)
	}
	if p.Reference != "" {
		values.Set("reference", p.Reference)
	}
	if p.WithBalance {
		values.Set("withBalance", "true")
	}
	return values
}
