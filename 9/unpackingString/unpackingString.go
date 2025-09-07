package unpackingstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

// UnpackString распаковывает строку
func UnpackString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var builder strings.Builder
	runes := []rune(s)
	var prev rune
	escaped := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case escaped:
			builder.WriteRune(r)
			prev = r
			escaped = false

		case r == '\\':
			escaped = true

		case unicode.IsDigit(r):
			if i == 0 || unicode.IsDigit(prev) {
				return "", ErrInvalidString
			}

			count, digits := extractNumber(runes[i:])
			for j := 0; j < count-1; j++ {
				builder.WriteRune(prev)
			}
			i += digits - 1

		default:
			builder.WriteRune(r)
			prev = r
		}
	}

	if escaped {
		return "", ErrInvalidString
	}

	return builder.String(), nil
}

// extractNumber извлекает число из начала среза рун
func extractNumber(runes []rune) (int, int) {
	var digits []rune

	for _, r := range runes {
		if !unicode.IsDigit(r) {
			break
		}
		digits = append(digits, r)
	}

	if len(digits) == 0 {
		return 0, 0
	}

	// игнорируем ошибку, так как digits гарантированно содержат только цифры
	count, _ := strconv.Atoi(string(digits))
	return count, len(digits)
}
