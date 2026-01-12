package pkgutils

import (
	"fmt"
	"time"
)

func ValidateBirthDate(birthDate time.Time, age int) error {
	calculatedAge := time.Now().Year() - birthDate.Year()

	// Ajustar la edad si aún no ha pasado el cumpleaños este año
	if time.Now().YearDay() < birthDate.YearDay() {
		calculatedAge--
	}

	if calculatedAge != age {
		return fmt.Errorf("birth date does not match provided age")
	}

	// Verificar que la fecha no sea futura
	if birthDate.After(time.Now()) {
		return fmt.Errorf("birth date cannot be in the future")
	}

	return nil
}
