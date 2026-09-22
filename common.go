package monime

// Metadata stores caller-defined key-value pairs attached to a resource.
// The API accepts up to 64 string properties.
type Metadata map[string]string

// OwnershipGraphOwner is one link in an ownership audit-trail chain.
type OwnershipGraphOwner struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Metadata Metadata             `json:"metadata,omitempty"`
	Owner    *OwnershipGraphOwner `json:"owner,omitempty"`
}

// OwnershipGraph traces the chain of objects that led to a resource.
type OwnershipGraph struct {
	Owner OwnershipGraphOwner `json:"owner"`
}

// Fee is a charge applied while processing a payment or payout.
type Fee struct {
	Code     string   `json:"code"`
	Amount   Amount   `json:"amount"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// ChannelType identifies the payment method used for a payment.
type ChannelType string

const (
	ChannelTypeBank   ChannelType = "bank"
	ChannelTypeCard   ChannelType = "card"
	ChannelTypeMomo   ChannelType = "momo"
	ChannelTypeWallet ChannelType = "wallet"
)

// Channel describes the payment method used for a payment.
//
// Fields outside Type depend on the channel type. Bank and mobile-money
// channels use Provider and Reference; bank uses AccountNumber; mobile-money
// uses PhoneNumber; wallet uses WalletID; card uses Scheme and Last4.
type Channel struct {
	Type          ChannelType `json:"type"`
	Provider      *string     `json:"provider,omitempty"`
	Reference     *string     `json:"reference,omitempty"`
	AccountNumber *string     `json:"accountNumber,omitempty"`
	PhoneNumber   *string     `json:"phoneNumber,omitempty"`
	WalletID      *string     `json:"walletId,omitempty"`
	Fingerprint   *string     `json:"fingerprint,omitempty"`
	Scheme        *string     `json:"scheme,omitempty"`
	Last4         *string     `json:"last4,omitempty"`
	Metadata      Metadata    `json:"metadata,omitempty"`
}
