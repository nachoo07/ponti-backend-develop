package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Claims con exactamente los mismos valores
	claims := jwt.MapClaims{
		"cuil": "20345678901",
		"exp":  int64(2547504000),
		"iat":  int64(1701204497),
	}

	// Crear el token usando el mismo método
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar con la misma clave
	secretKey := "ce5abdb2-9b00-431c-a213-8c815cb97226"
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error al firmar: %v\n", err)
		return
	}

	// Verificar el token inmediatamente
	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	fmt.Printf("\nToken generado (usar este):\n%s\n", tokenString)
}
