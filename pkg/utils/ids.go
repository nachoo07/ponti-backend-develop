package pkgutils

import (
	"errors"
	"strconv"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

func ValidateStringID(idParam string) (uint, error) {
	// Intentar convertir el parámetro a entero
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		// Si la conversión falla o el ID no es positivo, crear un error de dominio ErrInvalidID
		return 0, pkgtypes.NewInvalidIDError("Invalid ID parameter", err)
	}
	return uint(id), nil
}

func ValidateNumericID(id int64) error {
	if id <= 0 {
		return errors.New("id must be greater than 0")
	}
	return nil
}
