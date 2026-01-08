// Package lot implementa el repositorio de persistencia para la entidad Lot.
package lot

import (
	// standard library
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	// third-party
	"github.com/shopspring/decimal"
	gorm "gorm.io/gorm"
	"gorm.io/gorm/clause"

	// pkg
	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	// project
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
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

// CreateLot crea un nuevo lote si no existe uno con mismo nombre en el mismo field.
func (r *Repository) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	var lotID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Lot
		if err := tx.Where("name = ? AND field_id = ? AND deleted_at IS NULL", l.Name, l.FieldID).
			First(&existing).Error; err == nil {
			lotID = existing.ID
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrInternal, "failed to check lot", err)
		}
		model := models.FromDomain(l)
		model.CreatedBy = l.CreatedBy
		model.UpdatedBy = l.UpdatedBy

		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create lot", err)
		}
		lotID = model.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lotID, nil
}

// ListLotsByField lista los lotes por ID de field.
func (r *Repository) ListLotsByField(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	if err := r.db.Client().WithContext(ctx).
		Where("field_id = ? AND deleted_at IS NULL", fieldID).
		Find(&lots).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots", err)
	}
	return mapLotsToDomain(lots), nil
}

// GetLot obtiene un lote por ID.
func (r *Repository) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	var m models.Lot
	err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get lot", err)
	}
	return m.ToDomain(), nil
}

// UpdateLot actualiza datos del lote y fechas con upsert por secuencia.
func (r *Repository) UpdateLot(ctx context.Context, l *domain.Lot) error {
	if l.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", l.ID), nil)
	}

	userID, err := convertStringToID(ctx)
	if err != nil {
		return err
	}

	model := models.FromDomain(l)
	model.ID = l.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unicidad de nombre dentro del field (si aplica renombrado)
		if err := tx.Where("name = ? AND field_id = ? AND id <> ? AND deleted_at IS NULL",
			l.Name, l.FieldID, l.ID).First(&models.Lot{}).Error; err == nil {
			return types.NewError(types.ErrConflict, "lot with same name already exists in this field", nil)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrInternal, "failed to check lot unique name", err)
		}

		// Verificación de existencia (distingue 404 de 409)
		var exists int64
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", l.ID).
			Count(&exists).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if exists == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", l.ID), nil)
		}

		// Optimistic locking por fecha: WHERE id = ? AND updated_at = ? (última fecha conocida)
		nowTS := time.Now().UTC().Truncate(time.Microsecond)

		// Obtener la fecha de actualización actual para optimistic locking
		var currentLot models.Lot
		if err := tx.Where("id = ? AND deleted_at IS NULL", l.ID).First(&currentLot).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to get current lot for optimistic locking", err)
		}

		// Construir mapa de actualización solo con campos no vacíos
		updateFields := map[string]any{
			"updated_by": &userID,
			"updated_at": nowTS,
		}

		// Solo agregar campos que no estén vacíos
		if l.Name != "" {
			updateFields["name"] = l.Name
		}
		if l.Hectares.GreaterThan(decimal.Zero) {
			updateFields["hectares"] = l.Hectares
		}
		// Solo actualizar cultivos si se envían explícitamente
		if l.PreviousCrop.ID > 0 {
			updateFields["previous_crop_id"] = l.PreviousCrop.ID
		}
		if l.CurrentCrop.ID > 0 {
			updateFields["current_crop_id"] = l.CurrentCrop.ID
		}
		if l.Season != "" {
			updateFields["season"] = l.Season
		}
		if l.Variety != "" {
			updateFields["variety"] = l.Variety
		}

		res := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", l.ID).
			Updates(updateFields)
		if res.Error != nil {
			return types.NewError(types.ErrInternal, "failed to update lot", res.Error)
		}
		if res.RowsAffected == 0 {
			// Fila existe pero fecha de actualización no matchea -> conflicto por concurrencia
			return types.NewError(types.ErrConflict, "concurrent update conflict - lot was modified by another user", nil)
		}

		// Upsert de fechas por secuencia
		for _, date := range l.Dates {
			lotDate := models.LotDates{
				LotID:       l.ID,
				SowingDate:  date.SowingDate,
				HarvestDate: date.HarvestDate,
				Sequence:    date.Sequence,
				Base: sharedmodels.Base{
					CreatedBy: &userID,
					UpdatedBy: &userID,
					UpdatedAt: nowTS,
				},
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "lot_id"}, {Name: "sequence"}},
				DoUpdates: clause.AssignmentColumns([]string{"sowing_date", "harvest_date", "updated_by", "updated_at"}),
			}).Create(&lotDate).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to upsert lot dates", err)
			}
		}
		return nil
	})
}

func (r *Repository) UpdateLotTons(ctx context.Context, id int64, tons decimal.Decimal) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", id), nil)
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), nil)
		}
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(map[string]any{
				"tons": tons,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update lot tons", err)
		}
		return nil
	})
}

// DeleteLot elimina un lote por ID.
func (r *Repository) DeleteLot(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", id), nil)
	}
	userID, err := convertStringToID(ctx)
	if err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), nil)
		}
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": &userID,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to soft-delete lot", err)
		}
		return nil
	})
}

func (r *Repository) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND lots.deleted_at IS NULL", projectID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project", err)
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ? AND lots.deleted_at IS NULL", projectID, fieldID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project and field", err)
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	var lots []models.Lot
	db := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ? AND lots.deleted_at IS NULL", projectID, fieldID)
	switch cropType {
	case "current":
		db = db.Where("lots.current_crop_id = ?", cropID)
	case "previous":
		db = db.Where("lots.previous_crop_id = ?", cropID)
	case "both":
		db = db.Where("lots.current_crop_id = ? OR lots.previous_crop_id = ?", cropID, cropID)
	}
	err := db.Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project, field and crop", err)
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) GetMetrics(ctx context.Context, projectID, fieldID, cropID int64) (*domain.LotMetrics, error) {
	type rowAgg struct {
		SeededArea      decimal.Decimal `gorm:"column:seeded_area"`
		HarvestedArea   decimal.Decimal `gorm:"column:harvested_area"`
		YieldTnPerHa    decimal.Decimal `gorm:"column:yield_tn_per_ha"`
		CostPerHa       decimal.Decimal `gorm:"column:cost_per_ha"`
		SuperficieTotal decimal.Decimal `gorm:"column:project_total_hectares"`
		FieldTotal      decimal.Decimal `gorm:"column:field_total_hectares"`
	}

	// Construir query base
	query := r.db.Client().WithContext(ctx).Table("v3_lot_metrics")

	// Aplicar filtros
	if fieldID > 0 {
		query = query.Where("lot_id IN (SELECT id FROM lots WHERE field_id = ? AND deleted_at IS NULL)", fieldID)
	} else if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}

	if cropID > 0 {
		query = query.Where("lot_id IN (SELECT id FROM lots WHERE (current_crop_id = ? OR previous_crop_id = ?) AND deleted_at IS NULL)", cropID, cropID)
	}

	// Ejecutar query con agregaciones
	var row rowAgg
	err := query.Select(`
		COALESCE(SUM(sowed_area_ha), 0) AS seeded_area,
		COALESCE(SUM(harvested_area_ha), 0) AS harvested_area,
		COALESCE(SUM(yield_tn_per_ha * sowed_area_ha) / NULLIF(SUM(sowed_area_ha), 0), 0) AS yield_tn_per_ha,
		COALESCE(SUM(direct_cost_per_ha_usd * sowed_area_ha) / NULLIF(SUM(sowed_area_ha), 0), 0) AS cost_per_ha,
		COALESCE(MAX(project_total_hectares), 0) AS project_total_hectares,
		COALESCE(MAX(field_total_hectares), 0) AS field_total_hectares
	`).Scan(&row).Error

	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to scan lot metrics", err)
	}

	superficieTotal := row.SuperficieTotal
	if fieldID > 0 && row.FieldTotal.GreaterThan(decimal.Zero) {
		superficieTotal = row.FieldTotal
	}

	return &domain.LotMetrics{
		SeededArea:      row.SeededArea,
		HarvestedArea:   row.HarvestedArea,
		YieldTnPerHa:    row.YieldTnPerHa,
		CostPerHectare:  row.CostPerHa,
		SuperficieTotal: superficieTotal,
	}, nil
}

func (r *Repository) ListLots(
	ctx context.Context,
	projectID, fieldID, cropID int64,
	page, pageSize int,
) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {

	base := r.db.Client().WithContext(ctx).Table("v3_lot_list")

	// filtros
	if fieldID > 0 {
		base = base.Where("field_id = ?", fieldID)
	} else if projectID > 0 {
		base = base.Where("project_id = ?", projectID)
	}
	// Si no se proporcionan filtros, se retornan todos los lotes

	if cropID > 0 {
		base = base.Where("(current_crop_id = ? OR previous_crop_id = ?)", cropID, cropID)
	}

	// totales
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, types.NewError(types.ErrInternal, "failed to count lots", err)
	}

	var sumSowedArea decimal.Decimal
	if err := base.Session(&gorm.Session{}).Select("COALESCE(SUM(sowed_area_ha),0)").Scan(&sumSowedArea).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, types.NewError(types.ErrInternal, "failed to sum sowed area", err)
	}

	// sumCost: promedio ponderado de cost_usd_per_ha por sowed_area (para la card "Costo por hectárea")
	var sumCost decimal.Decimal
	if err := base.Session(&gorm.Session{}).
		Select("COALESCE(SUM(cost_usd_per_ha * sowed_area_ha) / NULLIF(SUM(sowed_area_ha),0), 0)").
		Scan(&sumCost).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, types.NewError(types.ErrInternal, "failed to calculate cost per hectare", err)
	}

	// página
	offset := (page - 1) * pageSize

	var rows []models.LotTable
	if err := base.Session(&gorm.Session{}).
		Select(`
	             project_id, field_id, project_name, field_name,
	             id, lot_name, variety, sowed_area_ha, hectares, season, updated_at, tons,
	             previous_crop_id, previous_crop,
	             current_crop_id, current_crop,
	             admin_cost_per_ha_usd AS admin_cost_per_ha,
	             harvested_area_ha AS harvested_area, lot_harvest_date AS harvest_date,
	             cost_usd_per_ha, yield_tn_per_ha,
	             income_net_per_ha_usd AS income_net_per_ha, rent_per_ha_usd AS rent_per_ha,
	             active_total_per_ha_usd AS active_total_per_ha, operating_result_per_ha_usd AS operating_result_per_ha
	         `).
		Order("id DESC").Limit(pageSize).Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, err
	}

	// Fechas por secuencia (1..3) para todos los lotes
	domainRows := make([]domain.LotTable, len(rows))
	if len(rows) > 0 {
		lotIDs := make([]int64, len(rows))
		for i := range rows {
			lotIDs[i] = rows[i].ID
		}
		var allDates []models.LotDates
		if err := r.db.Client().WithContext(ctx).Table("lot_dates").
			Where("lot_id IN ? AND deleted_at IS NULL", lotIDs).
			Order("lot_id, sequence"). // orden consistente
			Scan(&allDates).Error; err != nil {
			return nil, 0, decimal.Zero, decimal.Zero, types.NewError(types.ErrInternal, "failed to get lot dates", err)
		}
		datesByLot := make(map[int64][]models.LotDates)
		for _, d := range allDates {
			datesByLot[d.LotID] = append(datesByLot[d.LotID], d)
		}
		for i := range rows {
			domainRows[i] = rows[i].ToDomain(datesByLot[rows[i].ID])
		}
	}

	return domainRows, int(total), sumSowedArea, sumCost, nil
}

// mapLotsToDomain convierte slice de models.Lot a slice de domain.Lot
func mapLotsToDomain(lots []models.Lot) []domain.Lot {
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res
}

func convertStringToID(ctx context.Context) (int64, error) {
	userID := ctx.Value(pkgmwr.ContextUserID)
	if i, ok := userID.(int64); ok {
		return i, nil
	}
	if s, ok := userID.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, fmt.Errorf("failed to parse user ID: %w", err)
		}
	}
	return 0, fmt.Errorf("user ID is not a string")
}
