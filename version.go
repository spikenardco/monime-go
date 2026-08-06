package monime

import (
	"errors"
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

	return nil
}
