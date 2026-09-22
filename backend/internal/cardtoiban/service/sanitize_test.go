package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeProviderResponseMasksCardValuesAtEveryKnownLocation(t *testing.T) {
	t.Parallel()

	rawResponse := json.RawMessage(`{
		"data": {"card": "6037-9975-1234-5670"},
		"meta": {
			"errors": [
				{
					"property": "Card",
					"attemptedValue": "۶۰۳۷ ۹۹۷۵ ۱۲۳۴ ۵۶۷۰"
				},
				{
					"property": "requestHandlingType",
					"attemptedValue": "standard"
				}
			]
		}
	}`)

	sanitized := sanitizeProviderResponse(rawResponse)

	assert.NotContains(t, string(sanitized), "6037-9975-1234-5670")
	assert.NotContains(t, string(sanitized), "۶۰۳۷ ۹۹۷۵ ۱۲۳۴ ۵۶۷۰")
	assert.JSONEq(t, `{
		"data": {"card": "603799******5670"},
		"meta": {
			"errors": [
				{
					"property": "Card",
					"attemptedValue": "603799******5670"
				},
				{
					"property": "requestHandlingType",
					"attemptedValue": "standard"
				}
			]
		}
	}`, string(sanitized))
}

func TestSanitizeProviderResponseDropsMalformedJSON(t *testing.T) {
	t.Parallel()

	assert.Nil(t, sanitizeProviderResponse(json.RawMessage(`{"data":`)))
}
