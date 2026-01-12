package pkgutils

import (
	"errors"
	"log"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashea una contraseña utilizando bcrypt. Permite especificar el coste (rounds) o usar el valor por defecto.
// El coste determina el número de iteraciones del algoritmo y afecta tanto la seguridad como el rendimiento.
// Un valor común de coste es 10-12. A mayor coste, más seguro pero más lento.
func HashPassword(password string, cost int) (string, error) {
	// Si no se especifica un coste válido, usamos el coste por defecto de bcrypt
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}

	// Generar el hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Printf("Error al hashear la contraseña: %v", err)
		return "", errors.New("error al hashear la contraseña")
	}

	// Retornar el hash en formato string
	return string(hashedPassword), nil
}

// VerifyPassword verifica si una contraseña en texto plano coincide con su hash almacenado.
// Retorna true si coinciden, y false si no.
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// Comparar el hash almacenado con la contraseña ingresada
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			// Si las contraseñas no coinciden, retornamos false sin error
			return false, nil
		}
		// Si ocurre otro tipo de error, lo registramos
		log.Printf("Error al verificar la contraseña: %v", err)
		return false, errors.New("error al verificar la contraseña")
	}

	// Si coinciden, retornamos true
	return true, nil
}

// ValidatePasswordComplexity valida que la contraseña cumpla con los criterios de complejidad
func ValidatePasswordComplexity(password string) error {
	var hasMinLen, hasUpper, hasLower, hasNumber, hasSpecial bool
	const minLen = 8 // Longitud mínima de la contraseña

	if len(password) >= minLen {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Definir los criterios de complejidad
	if !hasMinLen {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
	}
	if !hasUpper {
		return errors.New("la contraseña debe tener al menos una letra mayúscula")
	}
	if !hasLower {
		return errors.New("la contraseña debe tener al menos una letra minúscula")
	}
	if !hasNumber {
		return errors.New("la contraseña debe tener al menos un número")
	}
	if !hasSpecial {
		return errors.New("la contraseña debe tener al menos un carácter especial")
	}

	return nil
}
