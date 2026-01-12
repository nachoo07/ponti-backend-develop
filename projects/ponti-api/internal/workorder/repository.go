package workorder

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	shareddb "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/db"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWorkorder(ctx context.Context, o *domain.Workorder) (int64, error) {
	// 1) convertir a modelo GORM (cabecera + items sin WorkorderID)
	model := models.FromDomain(o)

	// 2) poblar auditoría
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	// 3) crear todo en una transacción
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) insertar la cabecera para obtener model.ID
		if err := tx.Create(&model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create workorder header", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return model.ID, nil
}

func (r *Repository) GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error) {
	var m models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) GetWorkorderByNumberAndProjectID(ctx context.Context, number string, projectID int64) (*domain.Workorder, error) {
	var m models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Where("number = ?", number).
		Where("project_id = ?", projectID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) UpdateWorkorderByID(ctx context.Context, o *domain.Workorder) error {
	// 1) Convertimos dominio → GORM y fijamos el ID
	model := models.FromDomain(o)
	model.ID = o.ID

	// 2) Poblar UpdatedBy si hay usuario en contexto
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) Recuperar original para validar existencia y conservar auditoría
		var orig models.Workorder
		if err := tx.Preload("Items").First(&orig, model.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, "workorder not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to find workorder before update", err)
		}

		// 3.2) Eliminar todos los items antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkorderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old items", err)
		}

		// 3.3) Actualizar sólo la cabecera, omitiendo campos de auditoría y la asociación Items
		if err := tx.Model(&orig).
			Omit("CreatedAt", "CreatedBy", "DeletedAt", "DeletedBy", "Items").
			Updates(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update workorder header", err)
		}

		// 3.4) Insertar los items nuevos, asignando WorkorderID
		for i := range model.Items {
			model.Items[i].WorkorderID = model.ID
		}
		if len(model.Items) > 0 {
			if err := tx.Create(&model.Items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new items", err)
			}
		}

		return nil
	})
}

func (r *Repository) DeleteWorkorderByID(ctx context.Context, id int64) error {
	var m models.Workorder

	// 1) Buscamos la workorder por su ID
	if err := r.db.Client().
		WithContext(ctx).
		First(&m, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return types.NewError(types.ErrInternal, "failed to find workorder", err)
	}

	// 2) Soft-delete dentro de una transacción
	if err := r.db.Client().
		WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&m).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to soft delete workorder", err)
			}
			return nil
		}); err != nil {
		return err
	}

	return nil
}

func (r *Repository) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	// 1) Base del query: vinculada a la vista
	base := r.db.Client().
		WithContext(ctx).
		Model(&models.WorkorderListElement{})

	// 2) Aplicar filtros
	if filt.ProjectID != nil {
		base = base.Where("project_id = ?", *filt.ProjectID)
	}
	if filt.FieldID != nil {
		base = base.Where("field_id = ?", *filt.FieldID)
	}

	// 3) Contar total
	var total int64
	if err := base.
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count workorders", err)
	}

	// 4) Paginación
	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// 5) Recuperar filas paginadas (reutiliza 'base' con filtros)
	var rows []models.WorkorderListElement
	if err := base.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("number desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list workorders", err)
	}

	// 6) Mapear a dominio
	list := make([]domain.WorkorderListElement, len(rows))
	for i, m := range rows {
		list[i] = domain.WorkorderListElement{
			ID:                m.ID,
			Number:            m.Number,
			ProjectName:       m.ProjectName,
			FieldName:         m.FieldName,
			LotName:           m.LotName,
			Date:              m.Date,
			CropName:          m.CropName,
			LaborName:         m.LaborName,
			LaborCategoryName: m.LaborCategoryName,
			TypeName:          m.TypeName,
			Contractor:        m.Contractor,
			SurfaceHa:         m.SurfaceHa,
			SupplyName:        m.SupplyName,
			Consumption:       m.Consumption,
			CategoryName:      m.CategoryName,
			Dose:              m.Dose,
			CostPerHa:         m.CostPerHa,
			UnitPrice:         m.UnitPrice,
			TotalCost:         m.TotalCost,
		}
	}

	// 7) Construir PageInfo y devolver
	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

func (r *Repository) GetMetrics(ctx context.Context, filt domain.WorkorderFilter) (*domain.WorkorderMetrics, error) {
	// Construimos el WHERE dinámico según los filtros presentes
	q := fmt.Sprintf(`
		SELECT
		  COALESCE(SUM(surface_ha), 0) AS surface_ha,
		  COALESCE(SUM(liters), 0) AS liters,
		  COALESCE(SUM(kilograms), 0) AS kilograms,
		  COALESCE(SUM(direct_cost_usd), 0) AS direct_cost
		FROM %s
		WHERE 1=1
	`, shareddb.ReportView("workorder_metrics"))
	var args []any

	if filt.ProjectID != nil {
		q += " AND project_id = ?"
		args = append(args, *filt.ProjectID)
	}
	if filt.FieldID != nil {
		q += " AND field_id = ?"
		args = append(args, *filt.FieldID)
	}
	if filt.CustomerID != nil {
		q += " AND customer_id = ?"
		args = append(args, *filt.CustomerID)
	}
	if filt.CampaignID != nil {
		q += " AND campaign_id = ?"
		args = append(args, *filt.CampaignID)
	}

	var row struct {
		SurfaceHa  decimal.Decimal `gorm:"column:surface_ha"`
		Liters     decimal.Decimal `gorm:"column:liters"`
		Kilograms  decimal.Decimal `gorm:"column:kilograms"`
		DirectCost decimal.Decimal `gorm:"column:direct_cost"`
	}

	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get metrics", err)
	}

	return &domain.WorkorderMetrics{
		SurfaceHa:  row.SurfaceHa,
		Liters:     row.Liters,
		Kilograms:  row.Kilograms,
		DirectCost: row.DirectCost,
	}, nil
}

// GetRawDirectCost calcula el costo directo RAW desde las tablas workorders y workorder_items
// Calcula ∑(Órdenes_de_trabajo.costo_total) como indica el CSV de controles
// Este cálculo es INDEPENDIENTE de las vistas SSOT para validar coherencia
func (r *Repository) GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error) {
	// Query RAW: suma directa desde workorders + workorder_items
	// Labor cost: effective_area × labor.price
	// Supply cost: final_dose × effective_area × price (cálculo estándar del sistema)
	// CORREGIDO: Usar final_dose × effective_area en lugar de total_used para consistencia con vistas
	q := `
		WITH workorder_costs AS (
		  SELECT 
		    wo.id,
		    -- Costo de la labor (área efectiva × precio de la labor)
		    (wo.effective_area * l.price) AS labor_cost,
		    -- Costo de insumos (suma de items: final_dose × effective_area × price)
		    COALESCE((
		      SELECT SUM(wi.final_dose * wo.effective_area * s.price)
		      FROM public.workorder_items wi
		      JOIN public.supplies s ON s.id = wi.supply_id
		      WHERE wi.workorder_id = wo.id 
		        AND wi.deleted_at IS NULL
		    ), 0) AS supply_cost
		  FROM public.workorders wo
		  JOIN public.labors l ON l.id = wo.labor_id
		  WHERE wo.deleted_at IS NULL
		    AND wo.project_id = ?
		)
		SELECT COALESCE(SUM(labor_cost + supply_cost), 0) AS total_cost
		FROM workorder_costs
	`

	var totalCost decimal.Decimal
	if err := r.db.Client().WithContext(ctx).Raw(q, projectID).Scan(&totalCost).Error; err != nil {
		return decimal.Zero, types.NewError(types.ErrInternal, "failed to get raw direct cost", err)
	}

	return totalCost, nil
}
