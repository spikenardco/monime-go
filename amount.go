package monime

// Currency identifies the currency of an Amount.
type Currency string

const (
	// CurrencySLE is the Sierra Leonean leone currency code.
	CurrencySLE Currency = "SLE"
	// CurrencyUSD is the United States dollar currency code.
	CurrencyUSD Currency = "USD"
)

// Amount represents a monetary value in integer minor units.
type Amount struct {
	Currency Currency `json:"currency"`
	Value    int64    `json:"value"`
}
