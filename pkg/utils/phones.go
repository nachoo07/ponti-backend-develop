package pkgutils

import (
	"fmt"
	"strings"
	"unicode"
)

func ValidatePhone(phone string, digitsLen int) error {
	// Eliminar cualquier carácter no numérico
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phone)

	if len(digits) < digitsLen {
		return fmt.Errorf("phone number must have at least %d digits", digitsLen)
	}

	return nil
}
