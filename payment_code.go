package monime

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PaymentCodeMode is the usage mode of a payment code.
type PaymentCodeMode string

const (
	PaymentCodeModeOneTime   PaymentCodeMode = "one_time"
	PaymentCodeModeRecurrent PaymentCodeMode = "recurrent"
)

// PaymentCodeStatus is the lifecycle state of a payment code.
type PaymentCodeStatus string

const (
	PaymentCodeStatusPending    PaymentCodeStatus = "pending"
	PaymentCodeStatusCancelled  PaymentCodeStatus = "cancelled"
	PaymentCodeStatusProcessing PaymentCodeStatus = "processing"
	PaymentCodeStatusExpired    PaymentCodeStatus = "expired"
	PaymentCodeStatusCompleted  PaymentCodeStatus = "completed"
)

// PaymentCode is a token that generates a USSD dial string for collections.
type PaymentCode struct {
	ID                     string                  `json:"id"`
	Mode                   PaymentCodeMode         `json:"mode"`
	Status                 PaymentCodeStatus       `json:"status"`
	Name                   *string                 `json:"name,omitempty"`
	Amount                 *Amount                 `json:"amount,omitempty"`
	Enable                 bool                    `json:"enable"`
	ExpireTime             time.Time               `json:"expireTime"`
	Customer               *PaymentCodeCustomer    `json:"customer,omitempty"`
	USSDCode               string                  `json:"ussdCode"`
	Reference              *string                 `json:"reference,omitempty"`
	AuthorizedProviders    []string                `json:"authorizedProviders,omitempty"`
	AuthorizedPhoneNumber  *string                 `json:"authorizedPhoneNumber,omitempty"`
	RecurrentPaymentTarget *RecurrentPaymentTarget `json:"recurrentPaymentTarget,omitempty"`
	FinancialAccountID     *string                 `json:"financialAccountId,omitempty"`
	ProcessedPaymentData   *ProcessedPaymentData   `json:"processedPaymentData,omitempty"`
	CreateTime             time.Time               `json:"createTime"`
	UpdateTime             *time.Time              `json:"updateTime,omitempty"`
	OwnershipGraph         *OwnershipGraph         `json:"ownershipGraph,omitempty"`
	Metadata               Metadata                `json:"metadata,omitempty"`
}

// PaymentCodeCustomer identifies the customer shown in payment prompts.
type PaymentCodeCustomer struct {
	Name *string `json:"name,omitempty"`
}

// RecurrentPaymentTarget describes when a recurrent code is complete.
type RecurrentPaymentTarget struct {
	ExpectedPaymentCount *int64  `json:"expectedPaymentCount,omitempty"`
	ExpectedPaymentTotal *Amount `json:"expectedPaymentTotal,omitempty"`
}

// ProcessedPaymentData describes a payment processed from a payment code.
type ProcessedPaymentData struct {
	Amount                        Amount      `json:"amount"`
	OrderID                       string      `json:"orderId"`
	PaymentID                     string      `json:"paymentId"`
	OrderNumber                   string      `json:"orderNumber"`
	ChannelData                   ChannelData `json:"channelData"`
	FinancialTransactionReference string      `json:"financialTransactionReference"`
	Metadata                      Metadata    `json:"metadata,omitempty"`
}

// ChannelData identifies the provider transaction behind a processed payment.
type ChannelData struct {
	ProviderID string `json:"providerId"`
	AccountID  string `json:"accountId"`
	Reference  string `json:"reference"`
}

// CreatePaymentCodeInput contains fields accepted when creating a payment code.
type CreatePaymentCodeInput struct {
	Mode                   PaymentCodeMode         `json:"mode,omitempty"`
	Name                   string                  `json:"name"`
	Enable                 *bool                   `json:"enable,omitempty"`
	Amount                 *Amount                 `json:"amount,omitempty"`
	Duration               string                  `json:"duration,omitempty"`
	Customer               *PaymentCodeCustomer    `json:"customer,omitempty"`
	Reference              *string                 `json:"reference,omitempty"`
	AuthorizedProviders    []string                `json:"authorizedProviders,omitempty"`
	AuthorizedPhoneNumber  *string                 `json:"authorizedPhoneNumber,omitempty"`
	RecurrentPaymentTarget *RecurrentPaymentTarget `json:"recurrentPaymentTarget,omitempty"`
	FinancialAccountID     *string                 `json:"financialAccountId,omitempty"`
	Metadata               Metadata                `json:"metadata,omitempty"`
}

// UpdatePaymentCodeInput contains mutable payment-code fields.
type UpdatePaymentCodeInput struct {
	Name                   Field[string]                 `json:"name,omitzero"`
	Amount                 Field[Amount]                 `json:"amount,omitzero"`
	Duration               Field[string]                 `json:"duration,omitzero"`
	Enable                 Field[bool]                   `json:"enable,omitzero"`
	Customer               Field[PaymentCodeCustomer]    `json:"customer,omitzero"`
	Reference              Field[string]                 `json:"reference,omitzero"`
	AuthorizedProviders    Field[[]string]               `json:"authorizedProviders,omitzero"`
	AuthorizedPhoneNumber  Field[string]                 `json:"authorizedPhoneNumber,omitzero"`
	RecurrentPaymentTarget Field[RecurrentPaymentTarget] `json:"recurrentPaymentTarget,omitzero"`
	FinancialAccountID     Field[string]                 `json:"financialAccountId,omitzero"`
	Metadata               Field[Metadata]               `json:"metadata,omitzero"`
}

// PaymentCodeListParams filters a payment-code list request.
type PaymentCodeListParams struct {
	Limit    int
	After    string
	Mode     PaymentCodeMode
	USSDCode string
	Status   PaymentCodeStatus
}

// CreatePaymentCode creates a payment code.
func (c *Client) CreatePaymentCode(
	ctx context.Context,
	idempotencyKey string,
	input CreatePaymentCodeInput,
) (*PaymentCode, *Response, error) {
	if input.Name == "" {
		return nil, nil, newValidationError("Name", "must be non-empty")
	}
	if input.Mode == PaymentCodeModeOneTime && input.RecurrentPaymentTarget != nil {
		return nil, nil, newValidationError("RecurrentPaymentTarget", "must be omitted for one-time codes")
	}
	var code PaymentCode
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/payment-codes",
		input:          input,
		output:         &code,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &code, &response, nil
}

// GetPaymentCode returns a payment code by ID.
func (c *Client) GetPaymentCode(ctx context.Context, id string) (*PaymentCode, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var code PaymentCode
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/payment-codes/" + url.PathEscape(id),
		output: &code,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &code, &response, nil
}

// ListPaymentCodes returns payment codes matching params.
func (c *Client) ListPaymentCodes(
	ctx context.Context,
	params PaymentCodeListParams,
) (*Page[PaymentCode], *Response, error) {
	page, response, err := doList[PaymentCode](c, ctx, "/payment-codes", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdatePaymentCode updates mutable fields on a payment code.
func (c *Client) UpdatePaymentCode(
	ctx context.Context,
	id string,
	input UpdatePaymentCodeInput,
) (*PaymentCode, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !paymentCodeUpdateSet(input) {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var code PaymentCode
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/payment-codes/" + url.PathEscape(id),
		input:  input,
		output: &code,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &code, &response, nil
}

// DeletePaymentCode deletes an unused payment code.
func (c *Client) DeletePaymentCode(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/payment-codes/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func paymentCodeUpdateSet(input UpdatePaymentCodeInput) bool {
	return input.Name.IsSet() || input.Amount.IsSet() || input.Duration.IsSet() || input.Enable.IsSet() ||
		input.Customer.IsSet() || input.Reference.IsSet() || input.AuthorizedProviders.IsSet() ||
		input.AuthorizedPhoneNumber.IsSet() || input.RecurrentPaymentTarget.IsSet() ||
		input.FinancialAccountID.IsSet() || input.Metadata.IsSet()
}

func (p PaymentCodeListParams) values() url.Values {
	values := url.Values{}
	if p.Limit != 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.After != "" {
		values.Set("after", p.After)
	}
	if p.Mode != "" {
		values.Set("mode", string(p.Mode))
	}
	if p.USSDCode != "" {
		values.Set("ussd_code", p.USSDCode)
	}
	if p.Status != "" {
		values.Set("status", string(p.Status))
	}
	return values
}
