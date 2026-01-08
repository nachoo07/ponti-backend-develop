package shareddomain

import (
	"time"
)

type Base struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *int64
	UpdatedBy *int64
}
