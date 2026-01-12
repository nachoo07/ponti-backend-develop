package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type ClassType struct {
	ID   int64  // unique id
	Name string // class type name (e.g., "Agroquímicos", "Fertilizantes", etc.)

	shareddomain.Base
}
