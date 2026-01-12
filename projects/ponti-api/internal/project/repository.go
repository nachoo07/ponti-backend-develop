package project

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{db: db}
}

// CreateProject persists a project and all its associations in a single transaction.
func (r *repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var projectID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// Build GORM model from domain
		m := models.FromDomain(p)

		// 1. Insert project base
		if err := tx.
			Omit("Managers", "Investors", "Fields").
			Create(&m).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = m.ID

		// 2. Associate managers
		for _, mgr := range m.Managers {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id) SELECT ?, id FROM managers WHERE id = ?",
				m.ID, mgr.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate manager %d: %w", mgr.ID, err)
			}
		}

		// 3. Associate investors
		for _, inv := range m.Investors {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id) SELECT ?, id FROM investors WHERE id = ?",
				m.ID, inv.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate investor %d: %w", inv.ID, err)
			}
		}

		// 4. Associate fields
		for _, fld := range m.Fields {
			if err := tx.Exec(
				"INSERT INTO project_fields (project_id, field_id) SELECT ?, id FROM fields WHERE id = ?",
				m.ID, fld.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate field %d: %w", fld.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		// Wrap in application error
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, fmt.Sprintf("transaction failed for project creation: %v", err), err)
	}

	return projectID, nil
}

// ListProjects retrieves all projects with their associations.
func (r *repository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	var modelsList []models.Project
	if err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		Find(&modelsList).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list projects", err)
	}
	var result []domain.Project
	for _, m := range modelsList {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

// ListProjectsByCustomerID retrieves projects filtered by customer.
func (r *repository) ListProjectsByCustomerID(ctx context.Context, customerID int64) ([]domain.Project, error) {
	var modelsList []models.Project
	if err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		Where("customer_id = ?", customerID).
		Find(&modelsList).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, fmt.Sprintf("failed to list projects for customer %d: %v", customerID, err), err)
	}
	var result []domain.Project
	for _, m := range modelsList {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

// GetProject retrieves a single project by ID.
func (r *repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, fmt.Sprintf("failed to get project %d", id), err)
	}
	proj := m.ToDomain()
	return proj, nil
}

// UpdateProject updates a Project's main fields and relinks its ID-based relations.
func (r *repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	m := models.FromDomain(d)
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// update name and customer_id
		if err := tx.Model(&models.Project{}).
			Where("id = ?", d.ID).
			Updates(map[string]interface{}{"name": d.Name, "customer_id": d.Customer.ID}).Error; err != nil {
			return err
		}
		// relink managers
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Managers").Replace(m.Managers); err != nil {
			return err
		}
		// relink investors
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Investors").Replace(m.Investors); err != nil {
			return err
		}
		// relink fields
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Fields").Replace(m.Fields); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update project", err)
	}
	return nil
}

// DeleteProject removes a project and clears all its ID-based relations.
func (r *repository) DeleteProject(ctx context.Context, id int64) error {
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// clear managers
		if err := tx.Model(&models.Project{ID: id}).Association("Managers").Clear(); err != nil {
			return err
		}
		// clear investors
		if err := tx.Model(&models.Project{ID: id}).Association("Investors").Clear(); err != nil {
			return err
		}
		// clear fields
		if err := tx.Model(&models.Project{ID: id}).Association("Fields").Clear(); err != nil {
			return err
		}
		// delete project
		if err := tx.Delete(&models.Project{}, id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete project", err)
	}
	return nil
}
