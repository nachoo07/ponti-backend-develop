package sharedmodels

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	"gorm.io/gorm"
)

type Base struct {
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *int64         `gorm:"column:created_by"`
	UpdatedBy *int64         `gorm:"column:updated_by"`
	DeletedBy *int64         `gorm:"column:deleted_by"`
}

func ConvertStringToID(ctx context.Context) (int64, error) {
	userID := ctx.Value(pkgmwr.ContextUserID)
	if s, ok := userID.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, fmt.Errorf("failed to parse user ID: %w", err)
		}
	}
	return 0, fmt.Errorf("user ID is not a string")
}

// IncrementVersion incrementa la versión del modelo
// func (b *Base) IncrementVersion() {
// 	b.Version++
// }

// // GetVersion retorna la versión actual
// func (b *Base) GetVersion() int64 {
// 	return b.Version
// }
