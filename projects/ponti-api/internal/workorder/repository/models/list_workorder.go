package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkorderListElement mapea directamente la vista SQL
type WorkorderListElement struct {
	ID                int64           `gorm:"column:id;primaryKey"`
	Number            string          `gorm:"column:number;primaryKey"`
	ProjectName       string          `gorm:"column:project_name"`
	FieldName         string          `gorm:"column:field_name"`
	LotName           string          `gorm:"column:lot_name"`
	Date              time.Time       `gorm:"column:date"`
	CropName          string          `gorm:"column:crop_name"`
	LaborName         string          `gorm:"column:labor_name"`
	LaborCategoryName string          `gorm:"column:labor_category_name"`
	TypeName          string          `gorm:"column:type_name"`
	Contractor        string          `gorm:"column:contractor"`
	SurfaceHa         decimal.Decimal `gorm:"column:surface_ha"`
	SupplyName        string          `gorm:"column:supply_name"`
	Consumption       decimal.Decimal `gorm:"column:consumption"`
	CategoryName      string          `gorm:"column:category_name"`
	Dose              decimal.Decimal `gorm:"column:dose_per_ha"`
	CostPerHa         decimal.Decimal `gorm:"column:supply_cost_per_ha"`
	UnitPrice         decimal.Decimal `gorm:"column:unit_price"`
	TotalCost         decimal.Decimal `gorm:"column:supply_total_cost"`
}

// TableName apunta a la vista
func (WorkorderListElement) TableName() string {
	return "v3_workorder_list"
}
