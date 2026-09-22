package service

import (
	"net/netip"

	"github.com/google/uuid"
)

// CardToIBANRequest is the application input received from the HTTP layer.
type CardToIBANRequest struct {
	Card     string
	ClientIP *netip.Addr
}

// CardToIBANResponse is the safe application response returned to the HTTP
// layer.
type CardToIBANResponse struct {
	RequestID          uuid.UUID
	TrackID            uuid.UUID
	BankName           string
	EnglishBankName    string
	IBAN               string
	Deposit            string
	DepositDescription string
	DepositOwners      string
	DepositStatus      string
}
