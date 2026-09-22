package handler

import "github.com/google/uuid"

type convertRequest struct {
	Card string `json:"card"`
}

type convertResponse struct {
	Data convertResponseData `json:"data"`
}

type convertResponseData struct {
	RequestID          uuid.UUID `json:"requestId"`
	TrackID            uuid.UUID `json:"trackId"`
	BankName           string    `json:"bankName"`
	EnglishBankName    string    `json:"englishBankName"`
	IBAN               string    `json:"iban"`
	Deposit            string    `json:"deposit"`
	DepositDescription string    `json:"depositDescription"`
	DepositOwners      string    `json:"depositOwners"`
	DepositStatus      string    `json:"depositStatus"`
}

type errorResponse struct {
	Error errorResponseData `json:"error"`
}

type errorResponseData struct {
	Code           string     `json:"code"`
	Message        string     `json:"message"`
	RequestID      *uuid.UUID `json:"requestId,omitempty"`
	ValidationCode string     `json:"validationCode,omitempty"`
}
