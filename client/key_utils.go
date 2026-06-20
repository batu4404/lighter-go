package client

import (
	"fmt"
	"strings"
)

// AreKeysEqual reports whether two hex API keys are equal, ignoring an optional
// "0x" prefix on either. Mirrors lighter-python's are_keys_equal.
func AreKeysEqual(key1, key2 string) bool {
	return strings.TrimPrefix(key1, "0x") == strings.TrimPrefix(key2, "0x")
}

// ValidateAPIPrivateKeys validates a set of API private keys keyed by api key
// index, normalizing each by stripping an optional "0x" prefix in place.
// Returns an error if the set is empty or any key is blank. Mirrors
// lighter-python's validate_api_private_keys.
func ValidateAPIPrivateKeys(privateKeys map[int]string) error {
	if len(privateKeys) == 0 {
		return fmt.Errorf("no API keys provided")
	}
	for idx, pk := range privateKeys {
		trimmed := strings.TrimPrefix(pk, "0x")
		if trimmed == "" {
			return fmt.Errorf("api key index %d has an empty private key", idx)
		}
		privateKeys[idx] = trimmed
	}
	return nil
}
