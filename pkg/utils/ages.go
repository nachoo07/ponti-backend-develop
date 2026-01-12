package pkgutils

import "fmt"

func ValidateAge(age int, minAge, maxAge int) error {
	if age < minAge {
		return fmt.Errorf("age must be between %d and %d", minAge, maxAge)
	}

	if age > maxAge {
		return fmt.Errorf("age must be between %d and %d", minAge, maxAge)
	}

	return nil
}
