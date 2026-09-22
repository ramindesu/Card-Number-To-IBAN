package cardnumber

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const minimumFingerprintKeyLength = 32

var ErrFingerprintKeyTooShort = errors.New("card fingerprint key must be at least 32 bytes")

// Mask returns a safe audit value without exposing the full card number.
func Mask(normalizedCard string) string {
	if normalizedCard == "" {
		return "empty"
	}

	if len(normalizedCard) != cardNumberLength || !containsOnlyASCIIDigits(normalizedCard) {
		return "invalid"
	}

	return normalizedCard[:6] + "******" + normalizedCard[12:]
}

// Fingerprint returns a deterministic keyed fingerprint suitable for safely
// correlating repeated card input in audit records.
func Fingerprint(normalizedCard string, key []byte) (string, error) {
	if len(key) < minimumFingerprintKeyLength {
		return "", ErrFingerprintKeyTooShort
	}

	hash := hmac.New(sha256.New, key)
	_, _ = hash.Write([]byte(normalizedCard))

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func containsOnlyASCIIDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}
