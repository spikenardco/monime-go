package monime

import (
	"encoding/json"
	"testing"
)

func TestValidateAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		amount        Amount
		allowNegative bool
		wantErr       bool
	}{
		{
			name: "known currency",
			amount: Amount{
				Currency: CurrencySLE,
				Value:    100,
			},
		},
		{
			name: "zero value",
			amount: Amount{
				Currency: CurrencyUSD,
			},
		},
		{
			name: "negative value is disallowed",
			amount: Amount{
				Currency: CurrencySLE,
				Value:    -1,
			},
			wantErr: true,
		},
		{
			name: "negative value is allowed",
			amount: Amount{
				Currency: CurrencySLE,
				Value:    -1,
			},
			allowNegative: true,
		},
		{
			name: "empty currency",
			amount: Amount{
				Value: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid UTF-8 currency",
			amount: Amount{
				Currency: Currency("\xff"),
				Value:    100,
			},
			wantErr: true,
		},
		{
			name: "unsupported currency",
			amount: Amount{
				Currency: "XYZ",
				Value:    100,
			},
			wantErr: true,
		},
		{
			name: "currency with trailing space",
			amount: Amount{
				Currency: "SLE ",
				Value:    100,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateAmount(test.amount, test.allowNegative)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateAmount(%+v, %t) error = %v, want error = %t", test.amount, test.allowNegative, err, test.wantErr)
			}
		})
	}
}

func TestAmount_UnmarshalJSONPreservesUnknownCurrency(t *testing.T) {
	t.Parallel()

	var amount Amount
	if err := json.Unmarshal([]byte(`{"currency":"XYZ","value":100}`), &amount); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if amount.Currency != "XYZ" {
		t.Errorf("Currency = %q, want %q", amount.Currency, "XYZ")
	}
}
