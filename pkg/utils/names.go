package pkgutils

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateName(name string, minNameLength, maxNameLength int) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) < minNameLength || len(name) > maxNameLength {
		return fmt.Errorf("name length must be between %d and %d characters", minNameLength, maxNameLength)
	}

	if strings.Contains(name, "  ") {
		return errors.New("name cannot contain consecutive spaces")
	}

	return nil
}
