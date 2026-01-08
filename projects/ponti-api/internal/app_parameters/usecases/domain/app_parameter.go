package domain

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type AppParameter struct {
	ID          int64
	Key         string
	Value       string
	Type        string
	Category    string
	Description string

	shareddomain.Base
}

// GetValueAsDecimal convierte el valor a decimal.Decimal
func (ap *AppParameter) GetValueAsDecimal() (decimal.Decimal, error) {
	if ap.Type != "decimal" {
		return decimal.Zero, fmt.Errorf("parameter %s is not of type decimal", ap.Key)
	}
	val, err := strconv.ParseFloat(ap.Value, 64)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(val), nil
}

// GetValueAsInteger convierte el valor a int64
func (ap *AppParameter) GetValueAsInteger() (int64, error) {
	if ap.Type != "integer" {
		return 0, fmt.Errorf("parameter %s is not of type integer", ap.Key)
	}
	return strconv.ParseInt(ap.Value, 10, 64)
}

// GetValueAsBoolean convierte el valor a bool
func (ap *AppParameter) GetValueAsBoolean() (bool, error) {
	if ap.Type != "boolean" {
		return false, fmt.Errorf("parameter %s is not of type boolean", ap.Key)
	}
	return strconv.ParseBool(ap.Value)
}
