package pkgutils

import (
	"fmt"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the email format is valid
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 254 { // RFC 5321
		return fmt.Errorf("email exceeds maximum length of 254 characters")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}
