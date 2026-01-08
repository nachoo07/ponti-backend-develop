package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
)

type Crop struct {
	ID   int64
	Name string

	shareddomain.Base
}
