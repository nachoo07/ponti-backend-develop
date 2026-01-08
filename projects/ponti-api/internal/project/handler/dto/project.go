package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	campdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	leasetypedom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type Project struct {
	ID                 int64               `json:"id,omitempty"`
	ProjectName        string              `json:"name" binding:"required"`
	Customer           Customer            `json:"customer" binding:"required"`
	AdminCost          decimal.Decimal     `json:"admin_cost" binding:"required"`
	PlannedCost        decimal.Decimal     `json:"planned_cost" binding:"required"`
	Campaign           Campaign            `json:"campaign" binding:"required"`
	ProjectManagers    []Manager           `json:"managers" binding:"required,dive,required"`
	Investors          []Investor          `json:"investors" binding:"required,dive,required"`
	AdminCostInvestors []AdminCostInvestor `json:"admin_cost_investors" binding:"required,dive,required"`
	Fields             []Field             `json:"fields" binding:"required,dive,required"`
	UpdatedAt          *time.Time          `json:"updated_at,omitempty"`
	CreatedBy          *int64              `json:"created_by,omitempty"`
	UpdatedBy          *int64              `json:"updated_by,omitempty"`
}

type Customer struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name" binding:"required"`
}

type Campaign struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name" binding:"required"`
}

// Manager DTO
type Manager struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name" binding:"required"`
}

// Investor DTO
type Investor struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage" binding:"required"`
}

type AdminCostInvestor struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage" binding:"required"`
}

type Field struct {
	ID               int64            `json:"id,omitempty"`
	ProjectID        int64            `json:"project_id,omitempty"`
	Name             string           `json:"name" binding:"required"`
	LeaseTypeName    string           `json:"lease_type_name"`
	LeaseTypeID      int64            `json:"lease_type_id" binding:"required"`
	LeaseTypePercent *decimal.Decimal `json:"lease_type_percent"`
	LeaseTypeValue   *decimal.Decimal `json:"lease_type_value"`
	Investors        []Investor       `json:"investors" binding:"required,dive,required"`
	Lots             []Lot            `json:"lots" binding:"required,dive,required"`
}

// MarshalJSON aplica redondeo de 3 decimales a los campos decimales
func (f Field) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID               int64      `json:"id,omitempty"`
		ProjectID        int64      `json:"project_id,omitempty"`
		Name             string     `json:"name"`
		LeaseTypeName    string     `json:"lease_type_name"`
		LeaseTypeID      int64      `json:"lease_type_id"`
		LeaseTypePercent *string    `json:"lease_type_percent"`
		LeaseTypeValue   *string    `json:"lease_type_value"`
		Investors        []Investor `json:"investors"`
		Lots             []Lot      `json:"lots"`
	}{
		ID:            f.ID,
		ProjectID:     f.ProjectID,
		Name:          f.Name,
		LeaseTypeName: f.LeaseTypeName,
		LeaseTypeID:   f.LeaseTypeID,
		Investors:     f.Investors,
		Lots:          f.Lots,
	}

	if f.LeaseTypePercent != nil {
		val := f.LeaseTypePercent.Round(3).String()
		aux.LeaseTypePercent = &val
	}

	if f.LeaseTypeValue != nil {
		val := f.LeaseTypeValue.Round(3).String()
		aux.LeaseTypeValue = &val
	}

	return json.Marshal(aux)
}

type Lot struct {
	ID               int64           `json:"id,omitempty"`
	Name             string          `json:"name" binding:"required"`
	Hectares         decimal.Decimal `json:"hectares" binding:"required"`
	PreviousCropID   int64           `json:"previous_crop_id"`
	CurrentCropID    int64           `json:"current_crop_id"`
	PreviousCropName string          `json:"previous_crop_name"`
	CurrentCropName  string          `json:"current_crop_name"`
	Season           string          `json:"season" binding:"required"`
}

func (r *Project) ToDomain() *domain.Project {
	d := &domain.Project{
		Name: strings.TrimSpace(r.ProjectName),
		Customer: customerdom.Customer{
			ID:   r.Customer.ID,
			Name: r.Customer.Name,
		},
		Campaign: campdom.Campaign{
			ID:   r.Campaign.ID,
			Name: r.Campaign.Name,
		},
		AdminCost:   r.AdminCost,
		PlannedCost: r.PlannedCost,
	}

	for _, mgr := range r.ProjectManagers {
		d.Managers = append(d.Managers,
			managerdom.Manager{ID: mgr.ID, Name: mgr.Name},
		)
	}

	for _, inv := range r.Investors {
		d.Investors = append(d.Investors,
			investordom.Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage},
		)
	}

	for _, aci := range r.AdminCostInvestors {
		d.AdminCostInvestors = append(d.AdminCostInvestors,
			investordom.Investor{ID: aci.ID, Name: aci.Name, Percentage: aci.Percentage},
		)
	}

	for _, f := range r.Fields {
		fld := fielddom.Field{
			ID:               f.ID,
			Name:             f.Name,
			LeaseType:        &leasetypedom.LeaseType{ID: f.LeaseTypeID},
			LeaseTypePercent: f.LeaseTypePercent,
			LeaseTypeValue:   f.LeaseTypeValue,
		}
		for _, fi := range f.Investors {
			fld.Investors = append(fld.Investors, investordom.Investor{
				ID:         fi.ID,
				Name:       fi.Name,
				Percentage: fi.Percentage,
			})
		}
		for _, lt := range f.Lots {
			fld.Lots = append(fld.Lots, lotdom.Lot{
				ID:       lt.ID,
				Name:     lt.Name,
				Hectares: lt.Hectares,
				PreviousCrop: cropdom.Crop{
					ID:   lt.PreviousCropID,
					Name: lt.PreviousCropName,
				},
				CurrentCrop: cropdom.Crop{
					ID:   lt.CurrentCropID,
					Name: lt.CurrentCropName,
				},
				Season: lt.Season,
			})
		}

		d.Fields = append(d.Fields, fld)
	}

	return d
}

func FromDomain(d *domain.Project) *Project {
	r := &Project{
		ID:          d.ID,
		ProjectName: d.Name,
		Customer:    Customer{ID: d.Customer.ID, Name: d.Customer.Name},
		Campaign:    Campaign{ID: d.Campaign.ID, Name: d.Campaign.Name},
		AdminCost:   d.AdminCost,
		PlannedCost: d.PlannedCost,
		CreatedBy:   d.CreatedBy,
		UpdatedBy:   d.UpdatedBy,
		UpdatedAt:   &d.UpdatedAt,
	}

	for _, mgr := range d.Managers {
		r.ProjectManagers = append(r.ProjectManagers,
			Manager{ID: mgr.ID, Name: mgr.Name},
		)
	}

	for _, inv := range d.Investors {
		r.Investors = append(r.Investors,
			Investor{ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage},
		)
	}

	for _, aci := range d.AdminCostInvestors {
		r.AdminCostInvestors = append(r.AdminCostInvestors, AdminCostInvestor{
			ID: aci.ID, Name: aci.Name, Percentage: aci.Percentage,
		})
	}

	for _, fld := range d.Fields {
		dtoF := Field{
			ID:               fld.ID,
			Name:             fld.Name,
			LeaseTypeID:      fld.LeaseType.ID,
			LeaseTypeName:    fld.LeaseType.Name,
			LeaseTypePercent: fld.LeaseTypePercent,
			LeaseTypeValue:   fld.LeaseTypeValue,
		}

		for _, lt := range fld.Lots {
			dtoF.Lots = append(dtoF.Lots, Lot{
				ID:               lt.ID,
				Name:             lt.Name,
				Hectares:         lt.Hectares,
				PreviousCropID:   lt.PreviousCrop.ID,
				PreviousCropName: lt.PreviousCrop.Name,
				CurrentCropID:    lt.CurrentCrop.ID,
				CurrentCropName:  lt.CurrentCrop.Name,
				Season:           lt.Season,
			})
		}

		for _, fi := range fld.Investors {
			dtoF.Investors = append(dtoF.Investors, Investor{
				ID:         fi.ID,
				Name:       fi.Name,
				Percentage: fi.Percentage,
			})
		}
		r.Fields = append(r.Fields, dtoF)
	}

	return r
}

func FieldsFromDomain(d fielddom.Field) Field {
	r := Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseType.ID,
		ProjectID:   d.ProjectID,
	}

	for _, inv := range d.Investors {
		r.Investors = append(r.Investors, Investor{
			ID: inv.ID, Name: inv.Name, Percentage: inv.Percentage,
		})
	}
	for _, ld := range d.Lots {
		r.Lots = append(r.Lots, Lot{
			ID:               ld.ID,
			Name:             ld.Name,
			Hectares:         ld.Hectares,
			PreviousCropID:   ld.PreviousCrop.ID,
			PreviousCropName: ld.PreviousCrop.Name,
			CurrentCropID:    ld.CurrentCrop.ID,
			CurrentCropName:  ld.CurrentCrop.Name,
			Season:           ld.Season,
		})
	}
	return r
}
