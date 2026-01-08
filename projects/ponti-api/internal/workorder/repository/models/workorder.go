package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	labormod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	lotmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	projectmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	supplymod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

// Workorder GORM model con todas las relaciones
type Workorder struct {
	ID            int64              `gorm:"primaryKey;column:id"`
	Number        string             `gorm:"column:number;uniqueIndex"`
	ProjectID     int64              `gorm:"not null"`
	Project       projectmod.Project `gorm:"foreignKey:ProjectID"`
	FieldID       int64              `gorm:"not null"`
	Field         fieldmod.Field     `gorm:"foreignKey:FieldID"`
	LotID         int64              `gorm:"not null"`
	Lot           lotmod.Lot         `gorm:"foreignKey:LotID"`
	CropID        int64              `gorm:"not null"`
	Crop          cropmod.Crop       `gorm:"foreignKey:CropID"`
	LaborID       int64              `gorm:"not null"`
	Labor         labormod.Labor     `gorm:"foreignKey:LaborID"`
	Contractor    string             `gorm:"size:100"`
	Observations  string             `gorm:"size:1000"`
	Date          time.Time          `gorm:"type:date;not null"`
	InvestorID    int64              `gorm:"not null"`
	EffectiveArea decimal.Decimal    `gorm:"not null"`
	DeletedAt     gorm.DeletedAt     `gorm:"index"`
	Items         []WorkorderItem    `gorm:"foreignKey:WorkorderID;references:ID;constraint:OnDelete:CASCADE"`

	sharedmodels.Base
}

// WorkorderItem GORM model
type WorkorderItem struct {
	ID          int64            `gorm:"primaryKey;autoIncrement"`
	WorkorderID int64            `gorm:"column:workorder_id;index"`
	SupplyID    int64            `gorm:"not null"`
	Supply      supplymod.Supply `gorm:"foreignKey:SupplyID"`
	TotalUsed   decimal.Decimal  `gorm:"not null"`
	FinalDose   decimal.Decimal  `gorm:"not null"`
}

func FromDomain(o *domain.Workorder) *Workorder {
	w := &Workorder{
		Number:        o.Number,
		ProjectID:     o.ProjectID,
		FieldID:       o.FieldID,
		LotID:         o.LotID,
		CropID:        o.CropID,
		LaborID:       o.LaborID,
		Contractor:    o.Contractor,
		Observations:  o.Observations,
		Date:          o.Date,
		InvestorID:    o.InvestorID,
		EffectiveArea: o.EffectiveArea,
	}

	// Solo establecer ID si es mayor que 0 (para updates)
	if o.ID > 0 {
		w.ID = o.ID
	}

	if len(o.Items) > 0 {
		items := make([]WorkorderItem, len(o.Items))
		for i, it := range o.Items {
			items[i] = WorkorderItem{
				SupplyID:  it.SupplyID,
				TotalUsed: it.TotalUsed,
				FinalDose: it.FinalDose,
			}
		}
		w.Items = items
	}
	return w
}

// ToDomain convierte GORM → domain
func (m *Workorder) ToDomain() *domain.Workorder {
	var items []domain.WorkorderItem
	if len(m.Items) > 0 {
		items = make([]domain.WorkorderItem, len(m.Items))
		for i, it := range m.Items {
			items[i] = domain.WorkorderItem{
				SupplyID:  it.SupplyID,
				TotalUsed: it.TotalUsed,
				FinalDose: it.FinalDose,
			}
		}
	}
	return &domain.Workorder{
		ID:            m.ID,
		Number:        m.Number,
		ProjectID:     m.ProjectID,
		FieldID:       m.FieldID,
		LotID:         m.LotID,
		CropID:        m.CropID,
		LaborID:       m.LaborID,
		Contractor:    m.Contractor,
		Observations:  m.Observations,
		Date:          m.Date,
		InvestorID:    m.InvestorID,
		EffectiveArea: m.EffectiveArea,
		Items:         items,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}
