package service

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
)

func sanitizeProviderResponse(rawResponse json.RawMessage) json.RawMessage {
	if len(rawResponse) == 0 {
		return nil
	}

	var response map[string]any
	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil
	}

	sanitizeProviderValue(response)

	sanitizedResponse, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return sanitizedResponse
}

func sanitizeProviderValue(value any) {
	switch typedValue := value.(type) {
	case map[string]any:
		property, hasProperty := typedValue["property"].(string)
		propertyIsCard := hasProperty && strings.EqualFold(property, "card")

		for key, nestedValue := range typedValue {
			text, isText := nestedValue.(string)
			if isText && (strings.EqualFold(key, "card") ||
				(propertyIsCard && key == "attemptedValue")) {
				typedValue[key] = cardnumber.Mask(cardnumber.Normalize(text))
				continue
			}

			sanitizeProviderValue(nestedValue)
		}

	case []any:
		for _, nestedValue := range typedValue {
			sanitizeProviderValue(nestedValue)
		}
	}
}
