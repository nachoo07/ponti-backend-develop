package pkgutils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// isNumeric verifica si una cadena es numérica
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// normalizeString convierte una cadena a minúsculas y elimina acentos y caracteres especiales
func NormalizeString(input string) string {
	// Convertir a minúsculas
	input = strings.ToLower(input)

	// Eliminar acentos y caracteres especiales usando el paquete `transform`
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, input)

	// Eliminar cualquier carácter que no sea una letra de la 'a' a la 'z'
	clean := make([]rune, 0, len(result))
	for _, r := range result {
		if r >= 'a' && r <= 'z' {
			clean = append(clean, r)
		}
	}

	return string(clean)
}

// Elimina los espacios en blanco al inicio y final de la cadena
// Elimina todas las etiquetas HTML/XML de la cadena usando una expresión regular
func BasicInputSanitizer(input string) string {
	input = strings.TrimSpace(input)
	input = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(input, "")
	return input
}
