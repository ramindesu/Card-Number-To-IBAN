package zarinhub

import "github.com/google/uuid"

type tokenRequest struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type tokenData struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresTn"`
	RefreshToken string `json:"refreshToken"`
}

type cardToIBANRequest struct {
	Card                string    `json:"card"`
	TrackID             uuid.UUID `json:"trackId"`
	RequestHandlingType string    `json:"requestHandlingType"`
}

type cardToIBANData struct {
	BankName           string `json:"bankName"`
	Card               string `json:"card"`
	Deposit            string `json:"deposit"`
	DepositDescription string `json:"depositDescription"`
	DepositOwners      string `json:"depositOwners"`
	DepositStatus      string `json:"depositStatus"`
	EnglishBankName    string `json:"englishBankName"`
	IBAN               string `json:"iban"`
}

type responseEnvelope[T any] struct {
	Data T            `json:"data"`
	Meta responseMeta `json:"meta"`
}

type responseMeta struct {
	TrackID      string                    `json:"trackId"`
	Status       string                    `json:"status"`
	IsSuccess    bool                      `json:"isSuccess"`
	Code         int                       `json:"code"`
	Message      *string                   `json:"message"`
	ErrorMessage *string                   `json:"errorMessage"`
	ErrorType    *string                   `json:"errorType"`
	Errors       []responseValidationError `json:"errors"`
}

type responseValidationError struct {
	Property       *string `json:"property"`
	Message        string  `json:"message"`
	AttemptedValue any     `json:"attemptedValue"`
}
