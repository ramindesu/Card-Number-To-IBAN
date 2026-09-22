package client

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/model"
)

// CardToIBANRequest is the provider-independent request passed to the external
// Card-to-IBAN client.
type CardToIBANRequest struct {
	Card         string
	TrackID      uuid.UUID
	HandlingType model.CardToIBANRequestHandlingType
}

// CardToIBANData contains the successful, non-sensitive result returned by the
// provider.
type CardToIBANData struct {
	BankName           string
	EnglishBankName    string
	IBAN               string
	Deposit            string
	DepositDescription string
	DepositOwners      string
	DepositStatus      string
}

// CardToIBANResult contains a successful provider response and its raw JSON.
type CardToIBANResult struct {
	HTTPStatus  int
	MetaCode    int
	Data        CardToIBANData
	RawResponse json.RawMessage
}

// CardToIBANClient converts a normalized card number by calling an external
// provider.
type CardToIBANClient interface {
	Convert(context.Context, CardToIBANRequest) (CardToIBANResult, error)
}
