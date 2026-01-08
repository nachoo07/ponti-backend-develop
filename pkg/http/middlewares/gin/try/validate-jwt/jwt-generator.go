package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Generar claims
	claims := jwt.MapClaims{
		"cuil": "20345678901",
		"exp":  int64(2547504000),
		"iat":  int64(1701204497),
	}

	// Crear token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar token
	secretKey := "ce5abdb2-9b00-431c-a213-8c815cb97226"
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error al firmar: %v\n", err)
		return
	}

	// Verificar token
	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	fmt.Printf("\nToken generado (usar este):\n%s\n", tokenString)
}
