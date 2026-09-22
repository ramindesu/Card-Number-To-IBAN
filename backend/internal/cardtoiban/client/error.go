package client

import (
	"encoding/json"
	"fmt"
)

// CardToIBANErrorKind classifies failures without depending on provider error
// text.
type CardToIBANErrorKind string

const (
	CardToIBANErrorKindRejected       CardToIBANErrorKind = "rejected"
	CardToIBANErrorKindAuthentication CardToIBANErrorKind = "authentication"
	CardToIBANErrorKindTimeout        CardToIBANErrorKind = "timeout"
	CardToIBANErrorKindUnavailable    CardToIBANErrorKind = "unavailable"
	CardToIBANErrorKindTransport      CardToIBANErrorKind = "transport"
)

// CardToIBANError carries structured provider failure details required for
// audit and application-level error mapping.
type CardToIBANError struct {
	Kind        CardToIBANErrorKind
	HTTPStatus  *int
	MetaCode    *int
	ErrorType   *string
	RawResponse json.RawMessage
	Cause       error
}

// Error implements the error interface without exposing the raw response.
func (e *CardToIBANError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("card-to-IBAN client %s error: %v", e.Kind, e.Cause)
	}

	return fmt.Sprintf("card-to-IBAN client %s error", e.Kind)
}

// Unwrap makes the underlying transport error available to errors.Is and
// errors.As.
func (e *CardToIBANError) Unwrap() error {
	return e.Cause
}
