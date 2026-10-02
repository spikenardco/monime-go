package monime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PayoutStatus is the processing state of a payout.
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "pending"
	PayoutStatusProcessing PayoutStatus = "processing"
	PayoutStatusCompleted  PayoutStatus = "completed"
	PayoutStatusFailed     PayoutStatus = "failed"
)

// PayoutDestinationType identifies the recipient account kind.
type PayoutDestinationType string

const (
	PayoutDestinationBank   PayoutDestinationType = "bank"
	PayoutDestinationMomo   PayoutDestinationType = "momo"
	PayoutDestinationWallet PayoutDestinationType = "wallet"
)

// Payout is a disbursement to an external financial provider.
type Payout struct {
	ID             string               `json:"id"`
	Status         PayoutStatus         `json:"status"`
	Amount         Amount               `json:"amount"`
	Source         *PayoutSource        `json:"source,omitempty"`
	Destination    PayoutDestination    `json:"destination"`
	Fees           []Fee                `json:"fees,omitempty"`
	FailureDetail  *PayoutFailureDetail `json:"failureDetail,omitempty"`
	CreateTime     time.Time            `json:"createTime"`
	UpdateTime     time.Time            `json:"updateTime"`
	OwnershipGraph *OwnershipGraph      `json:"ownershipGraph,omitempty"`
	Metadata       Metadata             `json:"metadata,omitempty"`
}

// PayoutSource identifies the account debited by a payout.
type PayoutSource struct {
	FinancialAccountID   string  `json:"financialAccountId"`
	TransactionReference *string `json:"transactionReference,omitempty"`
}

// CreatePayoutSource selects the financial account debited by a payout.
type CreatePayoutSource struct {
	FinancialAccountID string `json:"financialAccountId"`
}

// PayoutDestination is a bank, mobile-money, or wallet recipient.
// Construct values with BankDestination, MobileMoneyDestination, or WalletDestination.
type PayoutDestination struct {
	typeName             PayoutDestinationType
	providerID           string
	accountNumber        string
	phoneNumber          string
	walletID             string
	transactionReference *string
}

// BankDestination returns a bank-account payout destination.
func BankDestination(providerID, accountNumber string) PayoutDestination {
	return PayoutDestination{
		typeName:      PayoutDestinationBank,
		providerID:    providerID,
		accountNumber: accountNumber,
	}
}

// MobileMoneyDestination returns a mobile-money payout destination.
func MobileMoneyDestination(providerID, phoneNumber string) PayoutDestination {
	return PayoutDestination{
		typeName:    PayoutDestinationMomo,
		providerID:  providerID,
		phoneNumber: phoneNumber,
	}
}

// WalletDestination returns a digital-wallet payout destination.
func WalletDestination(providerID, walletID string) PayoutDestination {
	return PayoutDestination{
		typeName:   PayoutDestinationWallet,
		providerID: providerID,
		walletID:   walletID,
	}
}

// Type returns the destination discriminator.
func (d PayoutDestination) Type() PayoutDestinationType { return d.typeName }

// ProviderID returns the Monime provider ID.
func (d PayoutDestination) ProviderID() string { return d.providerID }

// AccountNumber returns the bank account number, or an empty string for other variants.
func (d PayoutDestination) AccountNumber() string { return d.accountNumber }

// PhoneNumber returns the mobile-money number, or an empty string for other variants.
func (d PayoutDestination) PhoneNumber() string { return d.phoneNumber }

// WalletID returns the wallet ID, or an empty string for other variants.
func (d PayoutDestination) WalletID() string { return d.walletID }

// TransactionReference returns the provider transaction reference when present.
func (d PayoutDestination) TransactionReference() *string { return d.transactionReference }

// MarshalJSON emits only fields valid for the destination variant.
func (d PayoutDestination) MarshalJSON() ([]byte, error) {
	switch d.typeName {
	case PayoutDestinationBank:
		return json.Marshal(struct {
			Type          PayoutDestinationType `json:"type"`
			ProviderID    string                `json:"providerId"`
			AccountNumber string                `json:"accountNumber"`
		}{
			Type:          d.typeName,
			ProviderID:    d.providerID,
			AccountNumber: d.accountNumber,
		})
	case PayoutDestinationMomo:
		return json.Marshal(struct {
			Type        PayoutDestinationType `json:"type"`
			ProviderID  string                `json:"providerId"`
			PhoneNumber string                `json:"phoneNumber"`
		}{
			Type:        d.typeName,
			ProviderID:  d.providerID,
			PhoneNumber: d.phoneNumber,
		})
	case PayoutDestinationWallet:
		return json.Marshal(struct {
			Type       PayoutDestinationType `json:"type"`
			ProviderID string                `json:"providerId"`
			WalletID   string                `json:"walletId"`
		}{
			Type:       d.typeName,
			ProviderID: d.providerID,
			WalletID:   d.walletID,
		})
	default:
		return nil, errors.New("monime: invalid payout destination")
	}
}

// UnmarshalJSON decodes a destination returned by the API.
func (d *PayoutDestination) UnmarshalJSON(data []byte) error {
	var wire struct {
		Type                 PayoutDestinationType `json:"type"`
		ProviderID           string                `json:"providerId"`
		AccountNumber        string                `json:"accountNumber"`
		PhoneNumber          string                `json:"phoneNumber"`
		WalletID             string                `json:"walletId"`
		TransactionReference *string               `json:"transactionReference"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	*d = PayoutDestination{
		typeName:             wire.Type,
		providerID:           wire.ProviderID,
		accountNumber:        wire.AccountNumber,
		phoneNumber:          wire.PhoneNumber,
		walletID:             wire.WalletID,
		transactionReference: wire.TransactionReference,
	}
	return nil
}

// PayoutFailureDetail describes a failed payout.
type PayoutFailureDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreatePayoutInput contains fields accepted when creating a payout.
type CreatePayoutInput struct {
	Amount      Amount              `json:"amount"`
	Destination PayoutDestination   `json:"destination"`
	Source      *CreatePayoutSource `json:"source,omitempty"`
	Metadata    Metadata            `json:"metadata,omitempty"`
}

// UpdatePayoutInput contains mutable payout fields.
type UpdatePayoutInput struct {
	Metadata Field[Metadata] `json:"metadata,omitzero"`
}

// PayoutListParams filters a payout list request.
type PayoutListParams struct {
	Limit                           int
	After                           string
	Status                          PayoutStatus
	SourceFinancialAccountID        string
	SourceTransactionReference      string
	DestinationTransactionReference string
}

// CreatePayout creates a payout.
func (c *Client) CreatePayout(
	ctx context.Context,
	idempotencyKey string,
	input CreatePayoutInput,
) (*Payout, *Response, error) {
	if input.Amount.Value <= 0 {
		return nil, nil, newValidationError("Amount.Value", "must be positive")
	}
	if err := validatePayoutDestination(input.Destination); err != nil {
		return nil, nil, err
	}
	if input.Source != nil && input.Source.FinancialAccountID == "" {
		return nil, nil, newValidationError("Source.FinancialAccountID", "must be non-empty")
	}
	var payout Payout
	response, err := c.do(ctx, operation{
		method:         http.MethodPost,
		path:           "/payouts",
		input:          input,
		output:         &payout,
		idempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &payout, &response, nil
}

// GetPayout returns a payout by ID.
func (c *Client) GetPayout(ctx context.Context, id string) (*Payout, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	var payout Payout
	response, err := c.do(ctx, operation{
		method: http.MethodGet,
		path:   "/payouts/" + url.PathEscape(id),
		output: &payout,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &payout, &response, nil
}

// ListPayouts returns payouts matching params.
func (c *Client) ListPayouts(ctx context.Context, params PayoutListParams) (*Page[Payout], *Response, error) {
	page, response, err := doList[Payout](c, ctx, "/payouts", params.values())
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &page, &response, nil
}

// UpdatePayout updates mutable fields on a payout.
func (c *Client) UpdatePayout(
	ctx context.Context,
	id string,
	input UpdatePayoutInput,
) (*Payout, *Response, error) {
	if id == "" {
		return nil, nil, newValidationError("ID", "must be non-empty")
	}
	if !input.Metadata.IsSet() {
		return nil, nil, newValidationError("Input", "must contain at least one field")
	}
	var payout Payout
	response, err := c.do(ctx, operation{
		method: http.MethodPatch,
		path:   "/payouts/" + url.PathEscape(id),
		input:  input,
		output: &payout,
	})
	if err != nil {
		return nil, responseOrNil(response), err
	}
	return &payout, &response, nil
}

// DeletePayout deletes a payout that has not begun processing.
func (c *Client) DeletePayout(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, newValidationError("ID", "must be non-empty")
	}
	response, err := c.do(ctx, operation{
		method:  http.MethodDelete,
		path:    "/payouts/" + url.PathEscape(id),
		emptyOK: true,
	})
	return responseOrNil(response), err
}

func validatePayoutDestination(destination PayoutDestination) error {
	if destination.providerID == "" {
		return newValidationError("Destination.ProviderID", "must be non-empty")
	}
	switch destination.typeName {
	case PayoutDestinationBank:
		if destination.accountNumber == "" {
			return newValidationError("Destination.AccountNumber", "must be non-empty")
		}
	case PayoutDestinationMomo:
		if destination.phoneNumber == "" {
			return newValidationError("Destination.PhoneNumber", "must be non-empty")
		}
	case PayoutDestinationWallet:
		if destination.walletID == "" {
			return newValidationError("Destination.WalletID", "must be non-empty")
		}
	default:
		return newValidationError("Destination", "must be created with a destination constructor")
	}
	return nil
}

func (p PayoutListParams) values() url.Values {
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
	if p.SourceTransactionReference != "" {
		values.Set("sourceTransactionReference", p.SourceTransactionReference)
	}
	if p.DestinationTransactionReference != "" {
		values.Set("destinationTransactionReference", p.DestinationTransactionReference)
	}
	return values
}
