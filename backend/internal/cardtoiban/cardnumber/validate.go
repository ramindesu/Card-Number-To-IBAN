package cardnumber

const cardNumberLength = 16

// Validate checks that a normalized card number is non-empty, contains exactly
// 16 ASCII digits, and passes the Luhn checksum.
func Validate(card string) error {
	if card == "" {
		return &ValidationError{
			Code:    ValidationErrorCodeRequired,
			Message: "شماره کارت الزامی است",
		}
	}

	for _, character := range card {
		if character < '0' || character > '9' {
			return &ValidationError{
				Code:    ValidationErrorCodeNonNumeric,
				Message: "شماره کارت باید فقط شامل رقم باشد",
			}
		}
	}

	if len(card) != cardNumberLength {
		return &ValidationError{
			Code:    ValidationErrorCodeInvalidLength,
			Message: "شماره کارت باید ۱۶ رقم باشد",
		}
	}

	if hasOneRepeatedDigit(card) || !passesLuhn(card) {
		return &ValidationError{
			Code:    ValidationErrorCodeInvalidChecksum,
			Message: "شماره کارت معتبر نیست",
		}
	}

	return nil
}

// NormalizeAndValidate returns a canonical ASCII card number when the input is
// valid.
func NormalizeAndValidate(value string) (string, error) {
	normalized := Normalize(value)
	if err := Validate(normalized); err != nil {
		return normalized, err
	}

	return normalized, nil
}

func hasOneRepeatedDigit(card string) bool {
	firstDigit := card[0]
	for index := 1; index < len(card); index++ {
		if card[index] != firstDigit {
			return false
		}
	}

	return true
}

func passesLuhn(card string) bool {
	checksum := 0

	for index := 0; index < len(card); index++ {
		digit := int(card[index] - '0')
		if index%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		checksum += digit
	}

	return checksum%10 == 0
}
