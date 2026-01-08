package utils

import (
	"testing"
)

func TestMonthNameToNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Nombres completos
		{"January", "January", "01"},
		{"February", "February", "02"},
		{"March", "March", "03"},
		{"April", "April", "04"},
		{"May", "May", "05"},
		{"June", "June", "06"},
		{"July", "July", "07"},
		{"August", "August", "08"},
		{"September", "September", "09"},
		{"October", "October", "10"},
		{"November", "November", "11"},
		{"December", "December", "12"},

		// Nombres en minúsculas
		{"january", "january", "01"},
		{"june", "june", "06"},
		{"december", "december", "12"},

		// Abreviaciones
		{"Jan", "Jan", "01"},
		{"Feb", "Feb", "02"},
		{"Mar", "Mar", "03"},
		{"Apr", "Apr", "04"},
		{"Jun", "Jun", "06"},
		{"Jul", "Jul", "07"},
		{"Aug", "Aug", "08"},
		{"Sep", "Sep", "09"},
		{"Oct", "Oct", "10"},
		{"Nov", "Nov", "11"},
		{"Dec", "Dec", "12"},

		// Números de dos dígitos
		{"01", "01", "01"},
		{"06", "06", "06"},
		{"12", "12", "12"},

		// Números de un dígito
		{"1", "1", "01"},
		{"6", "6", "06"},
		{"9", "9", "09"},

		// Con espacios
		{" June ", " June ", "06"},
		{" 01 ", " 01 ", "01"},

		// Casos edge
		{"", "", ""},
		{"invalid", "invalid", "invalid"},
		{"13", "13", "13"}, // Número inválido, devuelve tal como está
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonthNameToNumber(tt.input)
			if result != tt.expected {
				t.Errorf("MonthNameToNumber(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid month 01", "01", true},
		{"Valid month 06", "06", true},
		{"Valid month 12", "12", true},
		{"Invalid month 00", "00", false},
		{"Invalid month 13", "13", false},
		{"Invalid month 1", "1", false},
		{"Invalid month 123", "123", false},
		{"Invalid month abc", "abc", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMonth(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidMonth(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
