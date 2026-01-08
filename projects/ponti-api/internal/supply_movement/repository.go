package supply_movement

import (
	"context"
	"errors"
	"fmt"
	"strings"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	providermodel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/repository/models"
	providerdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/provider/usecase/domain"
	stockmodel "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
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

func (r *Repository) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if movement == nil {
		return 0, types.NewError(types.ErrValidation, "supply movement is nil", nil)
	}
	model := models.FromDomain(movement)
	db := r.db.Client().WithContext(ctx)
	if err := db.Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) CreateProvider(ctx context.Context, provider *providerdomain.Provider) (int64, error) {
	if provider == nil {
		return 0, types.NewError(types.ErrValidation, "provider is nil", nil)
	}

	client := r.db.Client().WithContext(ctx)

	provider.Name = strings.TrimSpace(provider.Name)

	if provider.Name == "" {
		return 0, types.NewError(types.ErrValidation, "provider name is empty", nil)
	}

	model := providermodel.FromDomain(provider)
	providerId, err := ensureProvider(client, model)
	if err != nil {
		return 0, err
	}

	provider.ID = providerId

	return providerId, nil
}

func ensureProvider(tx *gorm.DB, i *providermodel.Provider) (int64, error) {
	if i.ID != 0 {
		var existing providermodel.Provider
		if err := tx.First(&existing, i.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
	}
	var existing providermodel.Provider
	if err := tx.Where("name = ?", i.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check provider: %w", err)
	}

	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create provider: %w", err)
	}
	return i.ID, nil
}

func (r *Repository) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error) {
	db := r.db.Client().WithContext(ctx)

	var modelSupplyMovements []models.SupplyMovement

	if err := db.
		Model(&models.SupplyMovement{}).
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("Provider").
		Joins("JOIN stocks ON supply_movements.stock_id = stocks.id").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Where("projects.id = ?", projectId).
		Where("is_entry = TRUE").
		Find(&modelSupplyMovements).
		Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplyEntriesMovement", err)
	}

	domainSupplyMovements := make([]*domain.SupplyMovement, len(modelSupplyMovements))
	for i, moddomainSupplyMovement := range modelSupplyMovements {
		domainSupplyMovements[i] = moddomainSupplyMovement.ToDomain()
	}

	return domainSupplyMovements, nil
}

func (r *Repository) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	db := r.db.Client().WithContext(ctx)

	var modelSupplyMovement models.SupplyMovement

	if err := db.
		Preload("Supply").
		Preload("Supply.Unit").
		Preload("Investor").
		Preload("Provider").
		First(&modelSupplyMovement, "id = ?", id).
		Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "supply movement not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply movement", err)
	}

	return modelSupplyMovement.ToDomain(), nil
}

func (r *Repository) UpdateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement == nil {
		return types.NewError(types.ErrValidation, "supply movement is nil", nil)
	}

	model := models.FromDomain(movement)
	db := r.db.Client().WithContext(ctx)

	if err := db.Model(&models.SupplyMovement{}).
		Where("id = ?", movement.ID).
		Updates(model).
		Error; err != nil {

		return types.NewError(types.ErrInternal, "failed to update supply movement", err)
	}

	return nil
}

func (r *Repository) DeleteSupplyMovement(ctx context.Context, projectId, supplyId int64) error {
	var stockModel stockmodel.Stock
	var supplyModel models.SupplyMovement
	client := r.db.Client().WithContext(ctx)

	// Obtener el movimiento a eliminar
	err := client.
		Where("project_id = ?", projectId).
		Where("id = ?", supplyId).
		First(&supplyModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.NewError(types.ErrNotFound, "supply movement not found", nil)
		}
		return err
	}

	// Verificar si hay stock cerrado
	err = client.
		Where("project_id = ?", projectId).
		Where("supply_id = ?", supplyModel.SupplyID).
		Where("close_date IS NOT NULL").
		First(&stockModel).Error
	if err == nil {
		return types.NewError(types.ErrConflict, "ya existe un movimiento de stock cerrado para este supply en el proyecto", nil)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Verificar si es un movimiento interno
	isInternalMovement := supplyModel.MovementType == "Movimiento interno" ||
		supplyModel.MovementType == "Movimiento interno entrada"

	err = r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if isInternalMovement {
			// Buscar todos los registros relacionados del movimiento interno
			// Los registros relacionados comparten: movement_date, reference_number, supply_id, investor_id, provider_id
			var relatedMovements []models.SupplyMovement
			err := tx.Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Find(&relatedMovements).Error
			if err != nil {
				return types.NewError(types.ErrInternal, "failed to find related movements", err)
			}

			// Recolectar todos los project_ids y stock_ids afectados
			affectedProjects := make(map[int64]bool)
			affectedStocks := make(map[int64]bool)

			for _, mov := range relatedMovements {
				affectedProjects[mov.ProjectId] = true
				affectedStocks[mov.StockId] = true
			}

			// Eliminar todos los registros relacionados
			if err := tx.Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Delete(&models.SupplyMovement{}).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete related supply movements", err)
			}

			// Eliminar stocks de todos los proyectos afectados solo si no tienen más movimientos
			// O actualizar RealStockUnits si quedan movimientos
			for projectIDAffected := range affectedProjects {
				// Obtener movimientos restantes para recalcular el stock
				var remainingMovements []models.SupplyMovement
				if err := tx.Where("project_id = ? AND supply_id = ?", projectIDAffected, supplyModel.SupplyID).
					Find(&remainingMovements).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to get remaining movements", err)
				}

				if len(remainingMovements) == 0 {
					// No quedan movimientos, eliminar el stock
					if err := tx.Delete(&stockmodel.Stock{}, "project_id = ? AND supply_id = ?", projectIDAffected, supplyModel.SupplyID).Error; err != nil {
						return types.NewError(types.ErrInternal, "failed to delete stock", err)
					}
				} else {
					// Recalcular RealStockUnits basándose en los movimientos restantes
					realStockUnits := decimal.Zero
					for _, mov := range remainingMovements {
						if mov.IsEntry {
							realStockUnits = realStockUnits.Add(mov.Quantity)
						} else {
							realStockUnits = realStockUnits.Sub(mov.Quantity)
						}
					}

					// Actualizar el stock con el nuevo RealStockUnits
					if err := tx.Model(&stockmodel.Stock{}).
						Where("project_id = ? AND supply_id = ?", projectIDAffected, supplyModel.SupplyID).
						Update("real_stock_units", realStockUnits).Error; err != nil {
						return types.NewError(types.ErrInternal, "failed to update stock real units", err)
					}
				}
			}
		} else {
			// Movimiento normal (no interno)
			// Primero eliminar el movimiento
			if err := tx.Delete(&models.SupplyMovement{}, "project_id = ? AND id = ?", projectId, supplyId).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete supply movement", err)
			}

			// Obtener movimientos restantes para recalcular el stock
			var remainingMovements []models.SupplyMovement
			if err := tx.Where("project_id = ? AND supply_id = ?", projectId, supplyModel.SupplyID).
				Find(&remainingMovements).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to get remaining movements", err)
			}

			if len(remainingMovements) == 0 {
				// No quedan movimientos, eliminar el stock
				if err := tx.Delete(&stockmodel.Stock{}, "project_id = ? AND supply_id = ?", projectId, supplyModel.SupplyID).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to delete stock", err)
				}
			} else {
				// Recalcular RealStockUnits basándose en los movimientos restantes
				realStockUnits := decimal.Zero
				for _, mov := range remainingMovements {
					if mov.IsEntry {
						realStockUnits = realStockUnits.Add(mov.Quantity)
					} else {
						realStockUnits = realStockUnits.Sub(mov.Quantity)
					}
				}

				// Actualizar el stock con el nuevo RealStockUnits
				if err := tx.Model(&stockmodel.Stock{}).
					Where("project_id = ? AND supply_id = ?", projectId, supplyModel.SupplyID).
					Update("real_stock_units", realStockUnits).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to update stock real units", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetProviders(ctx context.Context) ([]providerdomain.Provider, error) {
	var providers []providermodel.Provider
	if err := r.db.Client().WithContext(ctx).Find(&providers).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list providers", err)
	}
	res := make([]providerdomain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
