package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/invoice/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Invoice struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	WorkOrderID int64     `gorm:"not null;uniqueIndex;column:work_order_id"`
	Number      string    `gorm:"not null;column:number"`
	Company     string    `gorm:"varchar(100);not null;column:company"`
	Date        time.Time `gorm:"not null;column:date"`
	Status      string    `gorm:"varchar(100); not null; column:status"`

	sharedmodels.Base
}

func (Invoice) TableName() string {
	return "invoices"
}

func FromDomain(i *domain.Invoice) *Invoice {
	return &Invoice{
		ID:          i.ID,
		WorkOrderID: i.WorkOrderID,
		Number:      i.Number,
		Company:     i.Company,
		Date:        i.Date,
		Status:      i.Status,
		Base: sharedmodels.Base{
			CreatedBy: i.CreatedBy,
			UpdatedBy: i.UpdatedBy,
		},
	}
}

func (im *Invoice) ToDomain() *domain.Invoice {
	return &domain.Invoice{
		ID:          im.ID,
		WorkOrderID: im.WorkOrderID,
		Number:      im.Number,
		Company:     im.Company,
		Date:        im.Date,
		Status:      im.Status,
		Base: shareddomain.Base{
			CreatedAt: im.CreatedAt,
			CreatedBy: im.CreatedBy,
			UpdatedAt: im.UpdatedAt,
			UpdatedBy: im.UpdatedBy,
		},
	}
}
