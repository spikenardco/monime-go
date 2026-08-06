package monime

import (
	"errors"
	"unicode/utf8"
)

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

func validateAmount(amount Amount, allowNegative bool) error {
	if amount.Currency == "" {
		return errors.New("monime: amount currency is required")
	}

	if !utf8.ValidString(string(amount.Currency)) {
		return errors.New("monime: amount currency must be valid UTF-8")
	}

	if !allowNegative && amount.Value < 0 {
		return errors.New("monime: amount value must not be negative")
	}

	return nil
}
