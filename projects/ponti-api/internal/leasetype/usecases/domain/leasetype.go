package domain

import shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"

type LeaseType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`

	shareddomain.Base
}
