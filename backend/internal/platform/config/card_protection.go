package config

import (
	"errors"
	"os"
)

const minimumCardFingerprintKeyLength = 32

// CardProtection contains secrets used to safely correlate card requests.
type CardProtection struct {
	FingerprintKey []byte
}

// LoadCardProtection loads card protection settings from the environment.
func LoadCardProtection() (CardProtection, error) {
	if err := loadEnvironmentFile(); err != nil {
		return CardProtection{}, err
	}

	configuration := CardProtection{
		FingerprintKey: []byte(os.Getenv("CARD_FINGERPRINT_KEY")),
	}
	if err := configuration.Validate(); err != nil {
		return CardProtection{}, err
	}

	return configuration, nil
}

// Validate checks that the HMAC key has enough entropy-bearing capacity.
func (c CardProtection) Validate() error {
	if len(c.FingerprintKey) < minimumCardFingerprintKeyLength {
		return errors.New("CARD_FINGERPRINT_KEY must contain at least 32 bytes")
	}

	return nil
}
