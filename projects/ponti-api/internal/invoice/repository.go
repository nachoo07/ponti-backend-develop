package invoice

import (
	"context"
	"errors"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/usecases/domain"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByWorkOrderID(ctx context.Context, WorkOrderId int64) (*domain.Invoice, error) {
	if WorkOrderId == 0 {
		return nil, types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	var row models.Invoice
	if err := r.db.Client().WithContext(ctx).Where("work_order_id = ?", WorkOrderId).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			types.NewError(types.ErrNotFound, "There is no invoice for this workorder", err)
		}
		return nil, types.NewError(types.ErrInternal, "Failed to find invoice", err)

	}
	out := row.ToDomain()

	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.Invoice) (int64, error) {
	if item.WorkOrderID == 0 {
		return 0, types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(&m).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "fail to create invoice", err)
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.Invoice) error {
	if item.WorkOrderID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	err := r.db.Client().WithContext(ctx).
		Where("work_order_id = ?", item.WorkOrderID).
		Model(models.Invoice{}).
		Updates(map[string]any{
			"number":     item.Number,
			"company":    item.Company,
			"date":       item.Date,
			"status":     item.Status,
			"updated_at": time.Now(),
			"updated_by": item.UpdatedBy,
		}).Error

	if err != nil {
		return types.NewError(types.ErrInternal, "Failed to update invoice", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, WorkOrderId int64) error {
	if WorkOrderId == 0 {
		return types.NewError(types.ErrInvalidID, "invalid WorkOrderID", nil)
	}

	if err := r.db.Client().WithContext(ctx).Where("work_order_id = ?", WorkOrderId).Delete(&models.Invoice{}).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to delete invoice", err)
	}
	return nil
}
