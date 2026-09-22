package cardnumber

// ValidationErrorCode is a stable application-level code for invalid card
// input.
type ValidationErrorCode string

const (
	ValidationErrorCodeRequired        ValidationErrorCode = "card_required"
	ValidationErrorCodeNonNumeric      ValidationErrorCode = "card_non_numeric"
	ValidationErrorCodeInvalidLength   ValidationErrorCode = "card_invalid_length"
	ValidationErrorCodeInvalidChecksum ValidationErrorCode = "card_invalid_checksum"
)

// ValidationError describes a user-correctable card-number error.
type ValidationError struct {
	Code    ValidationErrorCode
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}
