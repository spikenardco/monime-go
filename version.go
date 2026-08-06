package monime

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// APIVersion identifies a versioned Monime API contract.
type APIVersion string

// APIVersionCaph20250823 is the default Monime API contract version.
const APIVersionCaph20250823 APIVersion = "caph.2025-08-23"

func validateAPIVersion(version APIVersion) error {
	if version == "" {
		return errors.New("monime: api version is required")
	}

	if !utf8.ValidString(string(version)) {
		return errors.New("monime: api version must be valid UTF-8")
	}

	value := string(version)
	if len(value) != len("caph.2006-01-02") || !strings.HasPrefix(value, "caph.") ||
		value[9] != '-' || value[12] != '-' ||
		!isASCIIDigits(value[5:9]) || !isASCIIDigits(value[10:12]) || !isASCIIDigits(value[13:15]) {
		return errors.New("monime: api version must have the caph.YYYY-MM-DD form")
	}

	return nil
}

func isASCIIDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}
