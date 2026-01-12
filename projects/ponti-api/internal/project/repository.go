package project

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	gorm "gorm.io/gorm"

	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	casmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	cusmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	domainField "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	invmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	lotmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	manmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	base "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
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

func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var projectID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userID, err := convertStringToID(ctx)
		if err != nil {
			return err
		}
		p.CreatedBy = &userID
		p.UpdatedBy = &userID

		customer := &cusmod.Customer{
			ID:   p.Customer.ID,
			Name: p.Customer.Name,
			Base: base.Base{
				CreatedBy: p.CreatedBy,
				UpdatedBy: p.UpdatedBy,
			},
		}
		custID, err := ensureCustomer(tx, customer)
		if err != nil {
			return err
		}
		p.Customer.ID = custID

		campaign := &casmod.Campaign{
			ID:   p.Campaign.ID,
			Name: p.Campaign.Name,
			Base: base.Base{
				CreatedBy: p.CreatedBy,
				UpdatedBy: p.UpdatedBy,
			},
		}
		campID, err := ensureCampaign(tx, campaign)
		if err != nil {
			return err
		}
		p.Campaign.ID = campID

		for i := range p.Managers {
			manager := &manmod.Manager{
				ID:   p.Managers[i].ID,
				Name: p.Managers[i].Name,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			mgrID, err := ensureManager(tx, manager)
			if err != nil {
				return err
			}
			p.Managers[i].ID = mgrID
		}

		for i := range p.Investors {
			investor := &invmod.Investor{
				ID:   p.Investors[i].ID,
				Name: p.Investors[i].Name,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			invID, err := ensureInvestor(tx, investor)
			if err != nil {
				return err
			}
			p.Investors[i].ID = invID
		}

		for aci := range p.AdminCostInvestors {
			AdminCostInv := &invmod.Investor{
				ID:   p.AdminCostInvestors[aci].ID,
				Name: p.AdminCostInvestors[aci].Name,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			aciID, err := ensureInvestor(tx, AdminCostInv)
			if err != nil {
				return err
			}
			p.AdminCostInvestors[aci].ID = aciID
		}

		for key := range p.Fields {
			p.Fields[key].ID = 0
			for key2, lot := range p.Fields[key].Lots {
				currentCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   lot.CurrentCrop.ID,
					Name: lot.CurrentCrop.Name,
				})
				if err != nil {
					return err
				}
				p.Fields[key].Lots[key2].CurrentCrop.ID = currentCropID
				previousCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   lot.PreviousCrop.ID,
					Name: lot.PreviousCrop.Name,
				})
				if err != nil {
					return err
				}
				p.Fields[key].Lots[key2].PreviousCrop.ID = previousCropID
				p.Fields[key].Lots[key2].ID = 0
			}
			for fi := range p.Fields[key].Investors {
				id, err := ensureInvestor(tx, &invmod.Investor{
					ID:   p.Fields[key].Investors[fi].ID,
					Name: p.Fields[key].Investors[fi].Name,
					Base: base.Base{
						CreatedBy: p.CreatedBy,
						UpdatedBy: p.UpdatedBy,
					},
				})
				if err != nil {
					return err
				}
				p.Fields[key].Investors[fi].ID = id
			}
		}

		projectModel := models.FromDomain(p)
		if err := tx.Create(&projectModel).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = projectModel.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return projectID, nil
}

// --- LIST ---
func (r *Repository) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []domain.ListedProject
	var total int64

	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{})

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count projects", err)
	}

	if err := db0.
		Select("id, name").
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects", err)
	}

	return projects, total, nil
}

func (r *Repository) GetProjects(ctx context.Context, name string, customerID int64, campaignID int64, page, perPage int) ([]domain.Project, decimal.Decimal, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []models.Project
	var total int64

	baseClient := r.db.Client().WithContext(ctx).Model(&models.Project{})
	sumClient := r.db.Client().WithContext(ctx).Model(&models.Project{})
	if name != "" {
		baseClient = baseClient.Where("projects.name = ?", name)
		sumClient = sumClient.Where("projects.name = ?", name)
	}
	if customerID > 0 {
		baseClient = baseClient.Where("customer_id = ?", customerID)
		sumClient = sumClient.Where("customer_id = ?", customerID)
	}
	if campaignID > 0 {
		baseClient = baseClient.Where("campaign_id = ?", campaignID)
		sumClient = sumClient.Where("campaign_id = ?", campaignID)
	}

	if err := baseClient.
		Count(&total).Error; err != nil {
		return nil, decimal.Zero, 0, types.NewError(types.ErrInternal, "failed to count projects", err)
	}

	var totalHectares decimal.Decimal

	if err := sumClient.
		Joins("JOIN fields ON fields.project_id = projects.id").
		Joins("JOIN lots ON lots.field_id = fields.id").
		Select("COALESCE(SUM(lots.hectares), 0)").
		Scan(&totalHectares).Error; err != nil {
		return nil, decimal.Zero, 0, types.NewError(types.ErrInternal, "failed to calculate total hectares", err)
	}

	if err := baseClient.
		Preload("Customer").
		Preload("Campaign").
		Preload("Managers").
		Preload("Investors.Investor").
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&projects).Error; err != nil {
		return nil, decimal.Zero, 0, types.NewError(types.ErrInternal, "failed to list projects", err)
	}

	var projectList []domain.Project
	for _, p := range projects {
		projectList = append(projectList, *p.ToDomain())
	}

	return projectList, totalHectares, total, nil
}

func (r *Repository) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []domain.ListedProject
	var total int64

	base := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{})

	if customerID > 0 {
		base = base.Where("customer_id = ?", customerID)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count projects by customer", err)
	}

	if err := base.
		Select("name").
		Group("name").
		Order("name ASC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects by customer", err)
	}

	return projects, total, nil
}

// --- GET ---
func (r *Repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", id), nil)
	}

	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Customer").
		Preload("Campaign").
		Preload("Managers").
		Preload("Investors.Investor").
		Preload("AdminCostInvestors.Investor").
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Fields.LeaseType").
		Preload("Fields.Lots", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Fields.FieldInvestors.Investor").
		Preload("Fields.Lots.PreviousCrop").
		Preload("Fields.Lots.CurrentCrop").
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, fmt.Sprintf("failed to get project %d", id), err)
	}

	return m.ToDomain(), nil
}

func (r *Repository) GetProjectByNameAndCampaignID(ctx context.Context, name string, campaignID int64) (*domain.Project, error) {
	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Where("name = ?", name).
		Where("campaign_id = ?", campaignID).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, types.NewError(types.ErrInternal, fmt.Sprintf("failed to get project %s", name), err)
	}

	return m.ToDomain(), nil
}

func (r *Repository) GetFieldsByProjectID(ctx context.Context, projectID int64) ([]domainField.Field, error) {
	var fields []fieldmod.Field
	err := r.db.Client().WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&fields).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, fmt.Sprintf("failed to get fields for project %d", projectID), err)
	}

	var fieldList []domainField.Field
	for _, f := range fields {
		fieldList = append(fieldList, *f.ToDomain())
	}

	return fieldList, nil
}

// --- UPDATE ---
func (r *Repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	if d.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", d.ID), nil)
	}

	var existing models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors.Investor").
		Preload("AdminCostInvestors.Investor").
		Preload("Fields").
		Preload("Fields.FieldInvestors.Investor").
		Preload("Fields.Lots").
		Where("id = ? AND updated_at = ?", d.ID, d.UpdatedAt).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return types.NewError(types.ErrNotFound, "registro desactualizado o no encontrado", nil)
	}
	if err != nil {
		return types.NewError(types.ErrInternal, "error buscando proyecto", err)
	}

	userID, err := convertStringToID(ctx)
	if err != nil {
		return err
	}
	d.UpdatedBy = &userID
	d.CreatedBy = existing.CreatedBy
	m := models.FromDomain(d)
	m.ID = d.ID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"updated_by": d.UpdatedBy,
		}

		if existing.CustomerID != d.Customer.ID {
			customer := &cusmod.Customer{
				ID:   d.Customer.ID,
				Name: d.Customer.Name,
				Base: base.Base{
					CreatedBy: d.UpdatedBy,
					UpdatedBy: d.UpdatedBy,
				},
			}
			custID, err := ensureCustomer(tx, customer)
			if err != nil {
				return err
			}
			d.Customer.ID = custID
			updates["customer_id"] = custID
		}

		if existing.CampaignID != d.Campaign.ID {
			campaign := &casmod.Campaign{
				ID:   d.Campaign.ID,
				Name: d.Campaign.Name,
				Base: base.Base{
					CreatedBy: d.UpdatedBy,
					UpdatedBy: d.UpdatedBy,
				},
			}
			campID, err := ensureCampaign(tx, campaign)
			if err != nil {
				return err
			}
			d.Campaign.ID = campID
			updates["campaign_id"] = campID
		}

		if existing.Name != d.Name {
			updates["name"] = d.Name
		}

		if existing.AdminCost != d.AdminCost {
			updates["admin_cost"] = d.AdminCost
		}

		if len(updates) > 0 {
			result := tx.Model(&models.Project{}).
				Where("id = ? AND updated_at = ?", d.ID, d.UpdatedAt).
				Updates(updates)
			if result.Error != nil {
				return types.NewError(types.ErrInternal, "failed to update project", result.Error)
			}

			if result.RowsAffected == 0 {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", d.ID), nil)
			}
		}

		err := relinkManagers(tx, existing, d)
		if err != nil {
			return err
		}

		err = relinkInvestors(tx, existing, d)
		if err != nil {
			return err
		}

		err = relinkAdminCostInvestors(tx, existing, d)
		if err != nil {
			return err
		}

		err = relinkFieldInvestors(tx, existing, d)
		if err != nil {
			return err
		}

		if err := relinkFieldsAndLots(tx, existing, m.Fields); err != nil {
			return err
		}

		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", id), nil)
	}

	userID, err := convertStringToID(ctx)
	if err != nil {
		return err
	}
	deletedBy := userID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Project{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check project existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", id), nil)
		}

		// clear managers
		if err := tx.Exec("UPDATE project_managers SET deleted_at = ?, deleted_by = ? WHERE project_id = ?", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear managers", err)
		}

		// clear investors
		if err := tx.Model(&models.ProjectInvestor{}).Where("project_id = ?", id).Updates(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear investors", err)
		}

		// clear fields
		var fieldIDs []int64
		if err := tx.Model(&fieldmod.Field{}).
			Where("project_id = ?", id).
			Pluck("id", &fieldIDs).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to get field ids", err)
		}

		if len(fieldIDs) > 0 {
			if err := tx.Model(&lotmod.Lot{}).
				Where("field_id IN ?", fieldIDs).
				Updates(map[string]any{
					"deleted_at": time.Now(),
					"deleted_by": deletedBy,
				}).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to soft delete lots", err)
			}
		}

		// clear workorders
		if err := tx.Exec("UPDATE workorders SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete workorders", err)
		}

		// clear supply_movements
		if err := tx.Exec("UPDATE supply_movements SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete supply_movements", err)
		}

		// clear stocks
		if err := tx.Exec("UPDATE stocks SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete stocks", err)
		}

		// clear crop_commercializations
		if err := tx.Exec("UPDATE crop_commercializations SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete commercializations", err)
		}

		// clear project_dollar_values
		if err := tx.Exec("UPDATE project_dollar_values SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete dollar values", err)
		}

		// clear admin_cost_investors
		if err := tx.Exec("UPDATE admin_cost_investors SET deleted_at = ?, deleted_by = ? WHERE project_id = ? AND deleted_at IS NULL", time.Now(), deletedBy, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete admin cost investors", err)
		}

		if err := tx.Model(&fieldmod.Field{}).
			Where("id IN ?", fieldIDs).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft delete fields", err)
		}

		// delete project
		if err := tx.Model(&models.Project{}).Where("id = ?", id).Updates(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete project", err)
		}

		return nil
	})
}

// --- RESTORE ---
// RestoreProject restaura un proyecto eliminado junto con todas sus entidades relacionadas
func (r *Repository) RestoreProject(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", id), nil)
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verificar que el proyecto esté eliminado
		var project models.Project
		if err := tx.Unscoped().Where("id = ?", id).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", id), err)
			}
			return types.NewError(types.ErrInternal, "failed to check project", err)
		}

		if !project.DeletedAt.Valid {
			return types.NewError(types.ErrValidation, "project is not deleted, cannot restore", nil)
		}

		// Restaurar project (usar Unscoped para actualizar registros eliminados)
		if err := tx.Unscoped().Model(&models.Project{}).Where("id = ?", id).Updates(map[string]any{
			"deleted_at": nil,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore project", err)
		}

		// Restaurar managers
		if err := tx.Exec("UPDATE project_managers SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore managers", err)
		}

		// Restaurar investors
		if err := tx.Exec("UPDATE project_investors SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore investors", err)
		}

		// Restaurar fields (obtener IDs primero)
		var fieldIDs []int64
		if err := tx.Unscoped().Model(&fieldmod.Field{}).
			Where("project_id = ? AND deleted_at IS NOT NULL", id).
			Pluck("id", &fieldIDs).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to get field ids", err)
		}

		// Restaurar fields
		if err := tx.Exec("UPDATE fields SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore fields", err)
		}

		// Restaurar lots (solo los que pertenecen a los fields de este proyecto)
		if len(fieldIDs) > 0 {
			if err := tx.Exec("UPDATE lots SET deleted_at = NULL, updated_at = ? WHERE field_id IN ? AND deleted_at IS NOT NULL", time.Now(), fieldIDs).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to restore lots", err)
			}
		}

		// Restaurar workorders
		if err := tx.Exec("UPDATE workorders SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore workorders", err)
		}

		// Restaurar supply_movements
		if err := tx.Exec("UPDATE supply_movements SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore supply_movements", err)
		}

		// Restaurar stocks
		if err := tx.Exec("UPDATE stocks SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore stocks", err)
		}

		// Restaurar crop_commercializations
		if err := tx.Exec("UPDATE crop_commercializations SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore commercializations", err)
		}

		// Restaurar project_dollar_values
		if err := tx.Exec("UPDATE project_dollar_values SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore dollar values", err)
		}

		// Restaurar admin_cost_investors
		if err := tx.Exec("UPDATE admin_cost_investors SET deleted_at = NULL, updated_at = ? WHERE project_id = ? AND deleted_at IS NOT NULL", time.Now(), id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore admin cost investors", err)
		}

		return nil
	})
}

// --- HELPERS ---

func ensureCustomer(tx *gorm.DB, c *cusmod.Customer) (int64, error) {
	if c.ID != 0 {
		var existing cusmod.Customer
		if err := tx.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check customer: %w", err)
		}
	}
	var existing cusmod.Customer
	if err := tx.Where("name = ?", c.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check customer: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create customer: %w", err)
	}
	return c.ID, nil
}

func ensureCampaign(tx *gorm.DB, c *casmod.Campaign) (int64, error) {
	if c.ID != 0 {
		var existing casmod.Campaign
		if err := tx.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check campaign: %w", err)
		}
	}
	var existing casmod.Campaign
	if err := tx.Where("name = ?", c.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check campaign: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create campaign: %w", err)
	}
	return c.ID, nil
}

func ensureManager(tx *gorm.DB, m *manmod.Manager) (int64, error) {
	if m.ID != 0 {
		var existing manmod.Manager
		if err := tx.First(&existing, m.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check manager: %w", err)
		}
	}
	var existing manmod.Manager
	if err := tx.Where("name = ?", m.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check manager: %w", err)
	}

	if err := tx.Create(m).Error; err != nil {
		return 0, fmt.Errorf("failed to create manager: %w", err)
	}
	return m.ID, nil
}

func ensureInvestor(tx *gorm.DB, i *invmod.Investor) (int64, error) {
	if i.ID != 0 {
		var existing invmod.Investor
		if err := tx.First(&existing, i.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
	}
	var existing invmod.Investor
	if err := tx.Where("name = ?", i.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check investor: %w", err)
	}

	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create investor: %w", err)
	}
	return i.ID, nil
}

func ensureCrop(tx *gorm.DB, c *cropmod.Crop) (int64, error) {
	if c.ID != 0 {
		var existing cropmod.Crop
		if err := tx.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check crop: %w", err)
		}
	}

	if c.Name == "" {
		return 0, fmt.Errorf("crop name is required")
	}

	var existing cropmod.Crop
	if err := tx.Where("name = ?", c.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check crop: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create crop: %w", err)
	}
	return c.ID, nil
}

func convertStringToID(ctx context.Context) (int64, error) {
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

func relinkManagers(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingManagerIDs := make(map[int64]struct{})
	for _, m := range existing.Managers {
		existingManagerIDs[m.ID] = struct{}{}
	}

	newManagerIDs := make(map[int64]struct{})
	for i, m := range d.Managers {
		if m.ID != 0 {
			newManagerIDs[m.ID] = struct{}{}
		} else {
			manager := &manmod.Manager{
				Name: m.Name,
				Base: base.Base{
					CreatedBy: d.UpdatedBy,
					UpdatedBy: d.UpdatedBy,
				},
			}
			mgrID, err := ensureManager(tx, manager)
			if err != nil {
				return err
			}
			newManagerIDs[mgrID] = struct{}{}
			d.Managers[i].ID = mgrID
		}
	}

	for _, m := range d.Managers {
		if _, exists := existingManagerIDs[m.ID]; !exists {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id, created_by, updated_by) VALUES (?, ?, ?, ?)",
				d.ID, m.ID, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add manager", err)
			}
		}
	}

	for _, m := range existing.Managers {
		if _, exists := newManagerIDs[m.ID]; !exists {
			if err := tx.Exec(
				"DELETE FROM project_managers WHERE project_id = ? AND manager_id = ?",
				d.ID, m.ID,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove manager", err)
			}
		}
	}

	return nil
}

func relinkInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingInvestorIDs := make(map[int64]struct{})
	for _, i := range existing.Investors {
		existingInvestorIDs[i.InvestorID] = struct{}{}
	}

	newInvestorIDs := make(map[int64]struct{})
	for k, i := range d.Investors {
		if i.ID != 0 {
			newInvestorIDs[i.ID] = struct{}{}
		} else {
			invID, err := ensureInvestor(tx, &invmod.Investor{
				Name: i.Name,
				Base: base.Base{
					CreatedBy: d.UpdatedBy,
					UpdatedBy: d.UpdatedBy,
				},
			})
			if err != nil {
				return err
			}
			newInvestorIDs[invID] = struct{}{}
			d.Investors[k].ID = invID
		}
	}

	for _, i := range d.Investors {
		if _, exists := existingInvestorIDs[i.ID]; !exists {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by) VALUES (?, ?, ?, ?, ?)",
				d.ID, i.ID, i.Percentage, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add investor", err)
			}
		}
	}

	for _, i := range existing.Investors {
		if _, exists := newInvestorIDs[i.InvestorID]; !exists {
			if err := tx.Exec(
				"DELETE FROM project_investors WHERE project_id = ? AND investor_id = ?",
				d.ID, i.InvestorID,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove investor", err)
			}
		}
	}

	return nil
}

func relinkFieldsAndLots(tx *gorm.DB, existing models.Project, fields []fieldmod.Field) error {
	// Mapear fields existentes y nuevos por ID
	existingFieldMap := make(map[int64]fieldmod.Field)
	for _, f := range existing.Fields {
		existingFieldMap[f.ID] = f
	}
	newFieldMap := make(map[int64]fieldmod.Field)
	for i, f := range fields {
		if f.ID != 0 {
			newFieldMap[f.ID] = f
		} else {
			f.ProjectID = existing.ID
			if err := tx.Create(&f).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add field", err)
			}
			newFieldMap[f.ID] = f
			fields[i] = f
		}
	}

	for _, f := range fields {
		if ef, exists := existingFieldMap[f.ID]; exists {
			updates := map[string]any{}
			if ef.Name != f.Name {
				updates["name"] = f.Name
			}
			if ef.LeaseTypeID != f.LeaseTypeID {
				updates["lease_type_id"] = f.LeaseTypeID
			}
			if *ef.LeaseTypePercent != *f.LeaseTypePercent {
				updates["lease_type_percent"] = *f.LeaseTypePercent
			}
			if *ef.LeaseTypeValue != *f.LeaseTypeValue {
				updates["lease_type_value"] = *f.LeaseTypeValue
			}
			if len(updates) > 0 {
				updates["updated_by"] = f.UpdatedBy
				if err := tx.Model(&fieldmod.Field{}).
					Where("id = ?", f.ID).
					Updates(updates).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to update field", err)
				}
			}
			if err := relinkLots(tx, ef, f); err != nil {
				return err
			}
		}
	}

	for _, ef := range existing.Fields {
		if _, exists := newFieldMap[ef.ID]; !exists {
			if err := tx.Exec("DELETE FROM lots WHERE field_id = ?", ef.ID).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove lots", err)
			}
			if err := tx.Exec("DELETE FROM fields WHERE id = ?", ef.ID).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove field", err)
			}
		}
	}
	return nil
}

func relinkLots(tx *gorm.DB, existingField, newField fieldmod.Field) error {
	existingLotsByID := make(map[int64]lotmod.Lot)
	for _, l := range existingField.Lots {
		existingLotsByID[l.ID] = l
	}

	newLotIDs := make(map[int64]struct{})
	for i, l := range newField.Lots {
		if l.ID != 0 {
			newLotIDs[l.ID] = struct{}{}
		} else {
			currentCropID, err := ensureCrop(tx, &cropmod.Crop{
				ID:   l.CurrentCrop.ID,
				Name: l.CurrentCrop.Name,
			})
			if err != nil {
				return err
			}
			l.CurrentCropID = currentCropID
			previousCropID, err := ensureCrop(tx, &cropmod.Crop{
				ID:   l.PreviousCrop.ID,
				Name: l.PreviousCrop.Name,
			})
			if err != nil {
				return err
			}
			l.PreviousCropID = previousCropID
			lot := lotmod.Lot{
				FieldID:        newField.ID,
				Name:           l.Name,
				Hectares:       l.Hectares,
				PreviousCropID: l.PreviousCropID,
				CurrentCropID:  l.CurrentCropID,
				Season:         l.Season,
				Base: base.Base{
					CreatedBy: l.CreatedBy,
					UpdatedBy: l.CreatedBy,
				},
			}
			if err := tx.Create(&lot).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add lot", err)
			}
			newLotIDs[lot.ID] = struct{}{}
			newField.Lots[i] = lot
		}
	}

	for _, l := range newField.Lots {
		if el, exists := existingLotsByID[l.ID]; exists {
			updates := map[string]any{}
			if l.Name != el.Name {
				updates["name"] = l.Name
			}
			if l.Hectares != el.Hectares {
				updates["hectares"] = l.Hectares
			}
			if l.PreviousCropID != el.PreviousCropID {
				previousCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   l.PreviousCrop.ID,
					Name: l.PreviousCrop.Name,
				})
				if err != nil {
					return err
				}
				l.PreviousCropID = previousCropID
				updates["previous_crop_id"] = l.PreviousCropID
			}
			if l.CurrentCropID != el.CurrentCropID {
				currentCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   l.CurrentCropID,
					Name: l.CurrentCrop.Name,
				})
				if err != nil {
					return err
				}
				l.CurrentCropID = currentCropID
				updates["current_crop_id"] = l.CurrentCropID
			}
			if l.Season != el.Season {
				updates["season"] = l.Season
			}
			if len(updates) > 0 {
				updates["updated_by"] = newField.UpdatedBy
				if err := tx.Model(&lotmod.Lot{}).
					Where("id = ?", l.ID).
					Updates(updates).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to update lot", err)
				}
			}
		}
	}

	for _, l := range existingField.Lots {
		if _, exists := newLotIDs[l.ID]; !exists {
			if err := tx.Exec("DELETE FROM lots WHERE id = ?", l.ID).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove lot", err)
			}
		}
	}
	return nil
}

func relinkAdminCostInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingAdCostInvIDs := make(map[int64]struct{})
	for _, aci := range existing.AdminCostInvestors {
		existingAdCostInvIDs[aci.InvestorID] = struct{}{}
	}

	newAdCostInvIDs := make(map[int64]struct{})
	for k, aci := range d.AdminCostInvestors {
		if aci.ID != 0 {
			newAdCostInvIDs[aci.ID] = struct{}{}
		} else {
			aciID, err := ensureInvestor(tx, &invmod.Investor{
				Name: aci.Name,
				Base: base.Base{
					CreatedBy: d.UpdatedBy,
					UpdatedBy: d.UpdatedBy,
				},
			})
			if err != nil {
				return err
			}
			newAdCostInvIDs[aciID] = struct{}{}
			d.AdminCostInvestors[k].ID = aciID
		}
	}

	for _, aci := range d.AdminCostInvestors {
		if _, exists := existingAdCostInvIDs[aci.ID]; !exists {
			if err := tx.Exec(
				"INSERT INTO admin_cost_investors (project_id, investor_id, percentage, created_by, updated_by) VALUES (?, ?, ?, ?, ?)",
				d.ID, aci.ID, aci.Percentage, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add admin cost investor", err)
			}
		} else {
			if err := tx.Exec(
				"UPDATE admin_cost_investors SET percentage = ?, updated_by = ? WHERE project_id = ? AND investor_id = ?",
				aci.Percentage, d.UpdatedBy, d.ID, aci.ID,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to update admin cost investor", err)
			}
		}
	}

	for _, aci := range existing.AdminCostInvestors {
		if _, exists := newAdCostInvIDs[aci.InvestorID]; !exists {
			if err := tx.Exec(
				"DELETE FROM admin_cost_investors WHERE project_id = ? AND investor_id = ?",
				d.ID, aci.InvestorID,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to remove investor", err)
			}
		}
	}

	return nil
}

func relinkFieldInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	// helper para matchear el field del payload contra el field existente (por ID o por nombre si es nuevo)
	findDomainField := func(fid int64, fname string) *domainField.Field {
		for i := range d.Fields {
			if d.Fields[i].ID > 0 && d.Fields[i].ID == fid {
				return &d.Fields[i]
			}
			if d.Fields[i].ID == 0 && d.Fields[i].Name == fname {
				return &d.Fields[i]
			}
		}
		return nil
	}

	for _, ef := range existing.Fields {
		df := findDomainField(ef.ID, ef.Name)
		if df == nil {
			continue
		}

		existingFieldInvIDs := make(map[int64]struct{}, len(ef.FieldInvestors))
		for _, fi := range ef.FieldInvestors {
			existingFieldInvIDs[fi.InvestorID] = struct{}{}
		}

		newIDs := make(map[int64]struct{}, len(df.Investors))
		for i := range df.Investors {
			inv := &df.Investors[i]
			if inv.ID == 0 {
				id, err := ensureInvestor(tx, &invmod.Investor{
					Name: inv.Name,
					Base: base.Base{CreatedBy: d.UpdatedBy, UpdatedBy: d.UpdatedBy},
				})
				if err != nil {
					return err
				}
				inv.ID = id
			}
			newIDs[inv.ID] = struct{}{}

			if _, existed := existingFieldInvIDs[inv.ID]; !existed {
				if err := tx.Exec(
					`INSERT INTO field_investors (field_id, investor_id, percentage, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?)`,
					ef.ID, inv.ID, inv.Percentage, d.UpdatedBy, d.UpdatedBy,
				).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to add field investor", err)
				}
			} else {
				if err := tx.Exec(
					"UPDATE field_investors SET percentage = ?, updated_by = ? WHERE field_id = ? AND investor_id = ?",
					inv.Percentage, d.UpdatedBy, ef.ID, inv.ID,
				).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to update admin cost investor", err)
				}
			}
		}

		for invID := range existingFieldInvIDs {
			if _, exists := newIDs[invID]; !exists {
				if err := tx.Exec(
					`DELETE FROM field_investors WHERE field_id = ? AND investor_id = ?`,
					ef.ID, invID,
				).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to remove field investor", err)
				}
			}
		}
	}
	return nil
}
