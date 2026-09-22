package cardnumber

import (
	"strings"
	"unicode"
)

// Normalize converts Persian and Arabic digits to ASCII and removes common
// visual separators. Unsupported characters are preserved for validation.
func Normalize(value string) string {
	var normalized strings.Builder
	normalized.Grow(len(value))

	for _, character := range value {
		switch {
		case character >= '\u06F0' && character <= '\u06F9':
			normalized.WriteRune('0' + character - '\u06F0')
		case character >= '\u0660' && character <= '\u0669':
			normalized.WriteRune('0' + character - '\u0660')
		case unicode.IsSpace(character), isHyphen(character):
			continue
		default:
			normalized.WriteRune(character)
		}
	}

	return normalized.String()
}

func isHyphen(character rune) bool {
	switch character {
	case '-', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2212':
		return true
	default:
		return false
	}
}
