package models

import (
	"time"

	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// Project es el modelo GORM para proyectos.
type Project struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name       string    `gorm:"size:100;not null;column:name"`
	CustomerID int64     `gorm:"not null;index;column:customer_id;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;column:updated_at"`

	Managers  []Manager  `gorm:"many2many:project_managers;association_autocreate:false;association_autoupdate:false;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Investors []Investor `gorm:"many2many:project_investors;association_autocreate:false;association_autoupdate:false;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Fields    []Field    `gorm:"many2many:project_fields;association_autocreate:false;association_autoupdate:false;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Manager sólo expone el ID para la tabla pivote project_managers.
type Manager struct {
	ID int64 `gorm:"primaryKey;column:id;autoIncrement:false"`
}

// Investor sólo expone el ID para la tabla pivote project_investors.
type Investor struct {
	ID int64 `gorm:"primaryKey;column:id;autoIncrement:false"`
}

// Field es el modelo GORM para campos de un proyecto (tabla 'fields').
type Field struct {
	ID int64 `gorm:"primaryKey;column:id;autoIncrement:false"`
}

// FromDomain convierte el dominio a modelo GORM, guardando sólo los IDs para asociaciones.
func FromDomain(d *domain.Project) *Project {
	m := &Project{
		ID:         d.ID,
		Name:       d.Name,
		CustomerID: d.Customer.ID,
	}
	for _, mgr := range d.Managers {
		m.Managers = append(m.Managers, Manager{ID: mgr.ID})
	}
	for _, inv := range d.Investors {
		m.Investors = append(m.Investors, Investor{ID: inv.ID})
	}
	for _, fld := range d.Fields {
		m.Fields = append(m.Fields, Field{ID: fld.ID})
	}
	return m
}

// ToDomain convierte el modelo GORM a dominio, cargando sólo los IDs.
func (m *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		ID:   m.ID,
		Name: m.Name,
		Customer: customerdom.Customer{
			ID: m.CustomerID,
		},
	}
	for _, mgr := range m.Managers {
		d.Managers = append(d.Managers, managerdom.Manager{ID: mgr.ID})
	}
	for _, inv := range m.Investors {
		d.Investors = append(d.Investors, investordom.Investor{ID: inv.ID})
	}
	for _, fld := range m.Fields {
		d.Fields = append(d.Fields, fielddom.Field{ID: fld.ID})
	}
	return d
}
