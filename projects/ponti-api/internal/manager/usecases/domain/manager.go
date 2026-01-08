package domain

import shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"

type Manager struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string // Sin tag, pero si querés lo podés poner json:"type"
	shareddomain.Base
}
	