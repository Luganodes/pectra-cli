package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateValidatorPubkeys checks if the provided validator public keys are valid.
// A valid public key is a 48-byte hex string, optionally prefixed with "0x".
func ValidateValidatorPubkeys(pubkeys []string) error {
	// Regular expression to match a hexadecimal string.
	hexRegex := regexp.MustCompile("^[0-9a-fA-F]+$")

	for _, pubkey := range pubkeys {
		// Remove "0x" prefix if present
		processedPubkey := strings.TrimPrefix(pubkey, "0x")

		// Check length (48 bytes = 96 hex characters)
		if len(processedPubkey) != 96 {
			return fmt.Errorf("invalid public key length for %s: expected 96 hex characters, got %d", pubkey, len(processedPubkey))
		}
		// Check for valid hexadecimal characters
		if !hexRegex.MatchString(processedPubkey) {
			return fmt.Errorf("public key %s contains non-hexadecimal characters", pubkey)
		}
	}
	return nil
}
