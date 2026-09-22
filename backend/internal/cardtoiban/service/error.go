package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/cardnumber"
)

// CardToIBANErrorCode is a stable application-level error code.
type CardToIBANErrorCode string

const (
	CardToIBANErrorCodeValidation          CardToIBANErrorCode = "validation_error"
	CardToIBANErrorCodeUpstreamRejected    CardToIBANErrorCode = "upstream_rejected"
	CardToIBANErrorCodeUpstreamAuthFailed  CardToIBANErrorCode = "upstream_auth_failed"
	CardToIBANErrorCodeUpstreamTimeout     CardToIBANErrorCode = "upstream_timeout"
	CardToIBANErrorCodeUpstreamUnavailable CardToIBANErrorCode = "upstream_unavailable"
	CardToIBANErrorCodeInternal            CardToIBANErrorCode = "internal_error"
)

// CardToIBANError is safe for the HTTP layer to map to a JSON error response.
type CardToIBANError struct {
	Code           CardToIBANErrorCode
	Message        string
	HTTPStatus     int
	RequestID      uuid.UUID
	ValidationCode cardnumber.ValidationErrorCode
	Cause          error
}

// Error implements the error interface.
func (e *CardToIBANError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the internal cause without including it in the safe message.
func (e *CardToIBANError) Unwrap() error {
	return e.Cause
}
