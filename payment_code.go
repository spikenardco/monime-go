package monime

import (
	"context"
	"net/url"
	"strconv"
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
	ExpireTime             string                  `json:"expireTime"`
	Customer               *PaymentCodeCustomer    `json:"customer,omitempty"`
	USSDCode               string                  `json:"ussdCode"`
	Reference              *string                 `json:"reference,omitempty"`
	AuthorizedProviders    []string                `json:"authorizedProviders,omitempty"`
	AuthorizedPhoneNumber  *string                 `json:"authorizedPhoneNumber,omitempty"`
	RecurrentPaymentTarget *RecurrentPaymentTarget `json:"recurrentPaymentTarget,omitempty"`
	FinancialAccountID     *string                 `json:"financialAccountId,omitempty"`
	ProcessedPaymentData   *ProcessedPaymentData   `json:"processedPaymentData,omitempty"`
	CreateTime             string                  `json:"createTime"`
	UpdateTime             *string                 `json:"updateTime,omitempty"`
	OwnershipGraph         *OwnershipGraph         `json:"ownershipGraph,omitempty"`
	Metadata               Metadata                `json:"metadata,omitempty"`
}

type PaymentCodeCustomer struct {
	Name *string `json:"name,omitempty"`
}

type RecurrentPaymentTarget struct {
	ExpectedPaymentCount *int64  `json:"expectedPaymentCount,omitempty"`
	ExpectedPaymentTotal *Amount `json:"expectedPaymentTotal,omitempty"`
}

type ProcessedPaymentData struct {
	Amount                        Amount      `json:"amount"`
	OrderID                       string      `json:"orderId"`
	PaymentID                     string      `json:"paymentId"`
	OrderNumber                   string      `json:"orderNumber"`
	ChannelData                   ChannelData `json:"channelData"`
	FinancialTransactionReference string      `json:"financialTransactionReference"`
	Metadata                      Metadata    `json:"metadata,omitempty"`
}

type ChannelData struct {
	ProviderID string `json:"providerId"`
	AccountID  string `json:"accountId"`
	Reference  string `json:"reference"`
}

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

type PaymentCodeListQuery struct {
	Limit    int
	After    string
	Mode     PaymentCodeMode
	USSDCode string
	Status   PaymentCodeStatus
}

type PaymentCodeService struct{ client *Client }

func (s *PaymentCodeService) Create(ctx context.Context, input CreatePaymentCodeInput) (PaymentCode, Response, error) {
	if input.Name == "" {
		return PaymentCode{}, Response{}, newValidationError("Name", "must be non-empty")
	}
	if input.Mode == PaymentCodeModeOneTime && input.RecurrentPaymentTarget != nil {
		return PaymentCode{}, Response{}, newValidationError("RecurrentPaymentTarget", "must be omitted for one-time codes")
	}
	if input.Mode == PaymentCodeModeRecurrent && input.RecurrentPaymentTarget == nil {
		return PaymentCode{}, Response{}, newValidationError("RecurrentPaymentTarget", "is required for recurrent codes")
	}
	var result PaymentCode
	response, err := s.client.do(ctx, operation{
		method: "POST",
		path:   "/payment-codes",
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *PaymentCodeService) Get(ctx context.Context, id string) (PaymentCode, Response, error) {
	if id == "" {
		return PaymentCode{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result PaymentCode
	response, err := s.client.do(ctx, operation{
		method: "GET",
		path:   "/payment-codes/" + url.PathEscape(id),
		output: &result,
	})
	return result, response, err
}

func (s *PaymentCodeService) List(
	ctx context.Context,
	query PaymentCodeListQuery,
) (Page[PaymentCode], Response, error) {
	return doList[PaymentCode](s.client, ctx, "/payment-codes", query.values())
}

func (s *PaymentCodeService) Update(
	ctx context.Context,
	id string,
	input UpdatePaymentCodeInput,
) (PaymentCode, Response, error) {
	if id == "" {
		return PaymentCode{}, Response{}, newValidationError("ID", "must be non-empty")
	}
	var result PaymentCode
	response, err := s.client.do(ctx, operation{
		method: "PATCH",
		path:   "/payment-codes/" + url.PathEscape(id),
		input:  input,
		output: &result,
	})
	return result, response, err
}

func (s *PaymentCodeService) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, newValidationError("ID", "must be non-empty")
	}
	return s.client.do(ctx, operation{
		method: "DELETE",
		path:   "/payment-codes/" + url.PathEscape(id),
	})
}

func (q PaymentCodeListQuery) values() url.Values {
	values := url.Values{}
	if q.Limit != 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.After != "" {
		values.Set("after", q.After)
	}
	if q.Mode != "" {
		values.Set("mode", string(q.Mode))
	}
	if q.USSDCode != "" {
		values.Set("ussdCode", q.USSDCode)
	}
	if q.Status != "" {
		values.Set("status", string(q.Status))
	}
	return values
}
