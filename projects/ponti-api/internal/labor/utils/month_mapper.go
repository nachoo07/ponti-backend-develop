package utils

import (
	"strings"
)

// MonthNameToNumber mapea nombres de meses en inglés a números de dos dígitos
// Acepta tanto nombres completos como abreviados, case-insensitive
func MonthNameToNumber(monthName string) string {
	month := strings.ToLower(strings.TrimSpace(monthName))

	monthMap := map[string]string{
		// Nombres completos
		"january":   "01",
		"february":  "02",
		"march":     "03",
		"april":     "04",
		"may":       "05",
		"june":      "06",
		"july":      "07",
		"august":    "08",
		"september": "09",
		"october":   "10",
		"november":  "11",
		"december":  "12",

		// Abreviaciones comunes
		"jan": "01",
		"feb": "02",
		"mar": "03",
		"apr": "04",
		"jun": "06",
		"jul": "07",
		"aug": "08",
		"sep": "09",
		"oct": "10",
		"nov": "11",
		"dec": "12",
	}

	// Si ya es un número de dos dígitos, devolverlo tal como está
	if len(month) == 2 && month >= "01" && month <= "12" {
		return month
	}

	// Si es un número de un dígito, agregar el cero al inicio
	if len(month) == 1 && month >= "1" && month <= "9" {
		return "0" + month
	}

	// Buscar en el mapa de nombres
	if mapped, exists := monthMap[month]; exists {
		return mapped
	}

	// Si no se encuentra, devolver el valor original (para compatibilidad)
	return monthName
}

// IsValidMonth verifica si el mes es válido (01-12)
func IsValidMonth(month string) bool {
	return len(month) == 2 && month >= "01" && month <= "12"
}
