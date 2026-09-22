package service

import (
	"bytes"
	"encoding/json"

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

	if data, ok := response["data"].(map[string]any); ok {
		if card, ok := data["card"].(string); ok {
			data["card"] = cardnumber.Mask(cardnumber.Normalize(card))
		}
	}

	sanitizedResponse, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return sanitizedResponse
}
