package cardnumber_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ascii digits",
			input:    "6037997512345670",
			expected: "6037997512345670",
		},
		{
			name:     "spaces and hyphens",
			input:    "6037-9975 1234-5670",
			expected: "6037997512345670",
		},
		{
			name:     "Persian digits",
			input:    "۶۰۳۷ ۹۹۷۵ ۱۲۳۴ ۵۶۷۰",
			expected: "6037997512345670",
		},
		{
			name:     "Arabic digits",
			input:    "٦٠٣٧٩٩٧٥١٢٣٤٥٦٧٠",
			expected: "6037997512345670",
		},
		{
			name:     "unsupported characters remain visible to validation",
			input:    "6037x997512345670",
			expected: "6037x997512345670",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, cardnumber.Normalize(test.input))
		})
	}
}

func TestNormalizeAndValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		expectedCard string
		expectedCode cardnumber.ValidationErrorCode
	}{
		{
			name:         "valid normalized card",
			input:        "۶۰۳۷-۹۹۷۵-۱۲۳۴-۵۶۷۰",
			expectedCard: "6037997512345670",
		},
		{
			name:         "empty",
			input:        "   ",
			expectedCard: "",
			expectedCode: cardnumber.ValidationErrorCodeRequired,
		},
		{
			name:         "non-numeric",
			input:        "60379975x2345670",
			expectedCard: "60379975x2345670",
			expectedCode: cardnumber.ValidationErrorCodeNonNumeric,
		},
		{
			name:         "invalid length",
			input:        "603799751234567",
			expectedCard: "603799751234567",
			expectedCode: cardnumber.ValidationErrorCodeInvalidLength,
		},
		{
			name:         "invalid checksum",
			input:        "6037997512345678",
			expectedCard: "6037997512345678",
			expectedCode: cardnumber.ValidationErrorCodeInvalidChecksum,
		},
		{
			name:         "repeated digits",
			input:        "0000000000000000",
			expectedCard: "0000000000000000",
			expectedCode: cardnumber.ValidationErrorCodeInvalidChecksum,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			normalized, err := cardnumber.NormalizeAndValidate(test.input)
			assert.Equal(t, test.expectedCard, normalized)

			if test.expectedCode == "" {
				require.NoError(t, err)
				return
			}

			var validationError *cardnumber.ValidationError
			require.ErrorAs(t, err, &validationError)
			assert.Equal(t, test.expectedCode, validationError.Code)
		})
	}
}

func TestMask(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "603799******5670", cardnumber.Mask("6037997512345670"))
	assert.Equal(t, "empty", cardnumber.Mask(""))
	assert.Equal(t, "invalid", cardnumber.Mask("6037x997512345670"))
	assert.Equal(t, "invalid", cardnumber.Mask("1234"))
}

func TestFingerprint(t *testing.T) {
	t.Parallel()

	key := []byte("0123456789abcdef0123456789abcdef")

	first, err := cardnumber.Fingerprint("6037997512345670", key)
	require.NoError(t, err)

	second, err := cardnumber.Fingerprint("6037997512345670", key)
	require.NoError(t, err)

	different, err := cardnumber.Fingerprint("6037997512345688", key)
	require.NoError(t, err)

	assert.Len(t, first, 64)
	assert.Equal(t, first, second)
	assert.NotEqual(t, first, different)
	assert.NotContains(t, first, "6037997512345670")

	_, err = cardnumber.Fingerprint("6037997512345670", []byte("short"))
	assert.True(t, errors.Is(err, cardnumber.ErrFingerprintKeyTooShort))
}
