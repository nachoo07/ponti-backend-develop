package domain

import (
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Field struct {
	ID          int64
	Name        string
	LeaseTypeID int64
	Lots        []lotdom.Lot
}
