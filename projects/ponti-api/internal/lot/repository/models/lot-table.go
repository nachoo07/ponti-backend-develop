// Package models contiene los modelos de persistencia para lot.
package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

type LotTable struct {
	ID             int64           `gorm:"column:id"`
	ProjectID      int64           `gorm:"column:project_id"`
	FieldID        int64           `gorm:"column:field_id"`
	ProjectName    string          `gorm:"column:project_name"`
	FieldName      string          `gorm:"column:field_name"`
	LotName        string          `gorm:"column:lot_name"`
	PreviousCrop   string          `gorm:"column:previous_crop"`
	PreviousCropID int64           `gorm:"column:previous_crop_id"`
	CurrentCrop    string          `gorm:"column:current_crop"`
	CurrentCropID  int64           `gorm:"column:current_crop_id"`
	Variety        string          `gorm:"column:variety"`
	SowedArea      decimal.Decimal `gorm:"column:sowed_area_ha"`
	Hectares       decimal.Decimal `gorm:"column:hectares"`
	Season         string          `gorm:"column:season"`
	Tons           decimal.Decimal `gorm:"column:tons"`
	UpdatedAt      *time.Time      `gorm:"column:updated_at"`

	AdminCost            decimal.Decimal `gorm:"column:admin_cost_per_ha"`
	HarvestedArea        decimal.Decimal `gorm:"column:harvested_area"`
	HarvestDate          *time.Time      `gorm:"column:harvest_date"`
	CostUsdPerHa         decimal.Decimal `gorm:"column:cost_usd_per_ha"`
	YieldTnPerHa         decimal.Decimal `gorm:"column:yield_tn_per_ha"`
	IncomeNetPerHa       decimal.Decimal `gorm:"column:income_net_per_ha"`
	RentPerHa            decimal.Decimal `gorm:"column:rent_per_ha"`
	ActiveTotalPerHa     decimal.Decimal `gorm:"column:active_total_per_ha"`
	OperatingResultPerHa decimal.Decimal `gorm:"column:operating_result_per_ha"`
}

type LotDates struct {
	LotID       int64      `gorm:"lot_id"`
	SowingDate  *time.Time `gorm:"sowing_date"`
	HarvestDate *time.Time `gorm:"harvest_date"`
	Sequence    int
	sharedmodels.Base
}

func (m *LotTable) ToDomain(dates []LotDates) domain.LotTable {
	domainDates := make([]domain.LotDates, 0, len(dates))
	for _, date := range dates {
		domainDates = append(domainDates, domain.LotDates{
			SowingDate:  date.SowingDate,
			HarvestDate: date.HarvestDate,
			Sequence:    date.Sequence,
		})
	}

	return domain.LotTable{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		ProjectName:    m.ProjectName,
		FieldID:        m.FieldID,
		FieldName:      m.FieldName,
		LotName:        m.LotName,
		PreviousCrop:   m.PreviousCrop,
		PreviousCropID: m.PreviousCropID,
		CurrentCrop:    m.CurrentCrop,
		CurrentCropID:  m.CurrentCropID,
		Variety:        m.Variety,
		SowedArea:      m.SowedArea,
		Hectares:       m.Hectares,
		Season:         m.Season,
		Tons:           m.Tons,
		Dates:          domainDates,
		UpdatedAt:      m.UpdatedAt,

		AdminCost:            m.AdminCost,
		HarvestedArea:        m.HarvestedArea,
		HarvestDate:          m.HarvestDate,
		CostUsdPerHa:         m.CostUsdPerHa,
		YieldTnPerHa:         m.YieldTnPerHa,
		IncomeNetPerHa:       m.IncomeNetPerHa,
		RentPerHa:            m.RentPerHa,
		ActiveTotalPerHa:     m.ActiveTotalPerHa,
		OperatingResultPerHa: m.OperatingResultPerHa,
	}
}
