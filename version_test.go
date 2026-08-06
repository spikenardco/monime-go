package monime

import "testing"

func TestValidateAPIVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version APIVersion
		wantErr bool
	}{
		{
			name:    "pinned version",
			version: APIVersionCaph20250823,
		},
		{
			name:    "future version",
			version: "caph.2026-01-01",
		},
		{
			name:    "empty version",
			wantErr: true,
		},
		{
			name:    "invalid UTF-8",
			version: APIVersion("\xff"),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateAPIVersion(test.version)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateAPIVersion(%q) error = %v, want error = %t", test.version, err, test.wantErr)
			}
		})
	}
}
