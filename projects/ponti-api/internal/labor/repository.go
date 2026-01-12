// Package labor contiene la implementación del repositorio para el módulo de labor
package labor

import (
	"context"
	"errors"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	shareddb "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/db"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/shopspring/decimal"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	if labor == nil {
		return 0, types.NewError(types.ErrValidation, "labor is nil", nil)
	}
	model := models.FromDomain(labor)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create labor", err)
	}
	return model.ID, nil
}

func (r *Repository) GetWorkordersByLaborID(ctx context.Context, laborID int64) (int64, error) {
	var count int64
	if err := r.db.Client().WithContext(ctx).
		Model(&workordermodels.Workorder{}).
		Joins("JOIN labors ON labors.id = workorders.labor_id").
		Where("labors.id = ?", laborID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return count, nil
}

func (r *Repository) DeleteLabor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Labor{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete labor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("labor with id %d does not exist", id), nil)
	}
	return nil
}

func (r *Repository) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if labor == nil {
		return types.NewError(types.ErrValidation, "labor is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("id = ?", labor.ID).
		Updates(models.FromDomain(labor))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update labor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("labor with id %d does not exist", labor.ID), nil)
	}
	return nil
}

func (r *Repository) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	var list []models.Labor
	var total int64

	db0 := r.db.Client().WithContext(ctx).Model(&models.Labor{})

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count labors", err)
	}

	if err := db0.
		Preload("Category").
		Select("id, name, contractor_name, price, category_id").
		Where("project_id = ?", projectID).
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list labor", err)
	}

	// Mapear a dominio ligero
	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:             labor.ID,
			Name:           labor.Name,
			Price:          labor.Price,
			ContractorName: labor.ContractorName,
			CategoryId:     labor.LaborCategoryID,
			CategoryName:   labor.Category.Name,
		}
	}

	return labors, total, nil
}

func (r *Repository) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	var laborCategoriesModels []models.LaborCategory
	db0 := r.db.
		Client().
		WithContext(ctx).
		Model(&models.LaborCategory{}).
		Where("type_id = ?", typeID)

	if err := db0.Find(&laborCategoriesModels).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list labor categories", err)
	}

	laborCategories := make([]domain.LaborCategory, len(laborCategoriesModels))
	for i, laborCategory := range laborCategoriesModels {
		laborCategories[i] = domain.LaborCategory{
			ID:   laborCategory.ID,
			Name: laborCategory.Name,
			LaborType: domain.LaborType{
				ID:   laborCategory.LaborType.ID,
				Name: laborCategory.LaborType.Name,
			},
		}
	}
	return laborCategories, nil
}

func (r *Repository) ListByWorkorder(ctx context.Context, workorderID int64) ([]domain.LaborRawItem, error) {
	var v3Models []models.LaborListItem

	// Usar la vista labor_list como base y agregar campos adicionales
	err := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v3").
		Select(`
            v3.workorder_id,
            v3.workorder_number,
            v3.date,
            v3.project_name,
            v3.field_name,
            COALESCE(v3.crop_name, '') AS crop_name,
            v3.labor_name,
            v3.contractor,
            v3.surface_ha,
            v3.cost_per_ha,
            COALESCE(v3.labor_category_name, '') AS category_name,
            COALESCE(v3.investor_name, '') AS investor_name,
			v3.dollar_average_month,
			v3.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
        `).
		Joins("LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id").
		Where("v3.workorder_id = ?", workorderID).
		Scan(&v3Models).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list labors by workorder", err)
	}

	// Convertir a LaborRawItem para mantener compatibilidad
	raws := make([]domain.LaborRawItem, len(v3Models))
	for i, m := range v3Models {
		// Manejar InvoiceID de forma segura
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		raws[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}
	return raws, nil
}

func (r *Repository) ListGroupLabor(
	ctx context.Context,
	inp types.Input,
	projectID int64,
	fieldID int64,
) ([]domain.LaborListItem, types.PageInfo, error) {

	// Base: vista labor_list + joins de factura y dólar promedio del mes
	base := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v3").
		Select(`
			v3.workorder_id,
			v3.workorder_number,
			v3.date,
			v3.project_id,
			v3.field_id,
			v3.project_name,
			v3.field_name,
			v3.lot_id,
			v3.lot_name,
			v3.crop_id,
			COALESCE(v3.crop_name, '') AS crop_name,
			v3.labor_id,
			v3.labor_name,
			v3.labor_category_id,
			COALESCE(v3.labor_category_name, '') AS labor_category_name,
			v3.contractor,
			v3.contractor_name,
			v3.surface_ha,
			v3.cost_per_ha,
			v3.total_labor_cost,
			v3.dollar_average_month,
			v3.dollar_average_month AS usd_avg_value,
			v3.usd_cost_ha,
			v3.usd_net_total,
			v3.investor_id,
			COALESCE(v3.investor_name, '') AS investor_name,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
		`).
		Joins("LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id")

	if fieldID != 0 {
		base = base.Where("v3.field_id = ?", fieldID)
	} else if projectID != 0 {
		base = base.Where("v3.project_id = ?", projectID)
	} else {
		return nil, types.PageInfo{}, types.NewError(types.ErrValidation, "fieldID or projectID is required", nil)
	}

	// Conteo para paginación
	var total int64
	if err := base.Session(&gorm.Session{}).Select("COUNT(*)").Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal, "failed to count labors for workorder", err)
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// Datos
	var rows []models.LaborListItem
	if err := base.Order("v3.workorder_number DESC").
		Limit(int(inp.PageSize)).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal, "failed to list grouped labors", err)
	}

	// IVA (tasa 0.105; si viene 1.105 se normaliza en getIVAPercentage)
	ivaRate, _ := r.getIVAPercentage(ctx)        // 0.105
	ivaMul := decimal.NewFromInt(1).Add(ivaRate) // 1.105

	// Mapear a dominio
	list := make([]domain.LaborListItem, len(rows))
	for i, m := range rows {
		// Costo ARS/ha SIN IVA = USD/ha × dólar prom
		costHaARS := m.USDCostHa.Mul(m.USDAvgValue)

		// Total neto ARS (SIN IVA)
		netTotal := costHaARS.Mul(m.SurfaceHa)

		// "Total IVA" (histórico): mostramos TOTAL CON IVA para que coincida con la UI actual
		totalConIVA := netTotal.Mul(ivaMul) // net × 1.105

		// USD (vienen de la vista)
		usdCostHa := m.USDCostHa
		usdNetTotal := m.USDNetTotal

		// Invoice seguro
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		list[i] = domain.LaborListItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          costHaARS, // ARS/ha SIN IVA (10 × 1000 = 10.000)
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,    // 10.000 × 100 = 1.000.000
			TotalIVA:        totalConIVA, // MOSTRAMOS TOTAL CON IVA: 1.000.000 × 1.105 = 1.105.000
			USDCostHa:       usdCostHa,   // 10
			USDNetTotal:     usdNetTotal, // 1000
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// getIVAPercentage obtiene el porcentaje de IVA desde app_parameters
func (r *Repository) getIVAPercentage(ctx context.Context) (decimal.Decimal, error) {
	var value string
	err := r.db.Client().WithContext(ctx).
		Table("app_parameters").
		Select("value").
		Where("key = ?", "iva_percentage").
		Scan(&value).Error
	if err != nil || value == "" {
		return decimal.NewFromFloat(0.105), nil
	}
	v, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.NewFromFloat(0.105), nil
	}
	if v.GreaterThan(decimal.NewFromInt(1)) {
		v = v.Sub(decimal.NewFromInt(1)) // 1.105 -> 0.105
	}
	return v, nil
}

// TODO: Eliminar este método
// ListGroupLaborOld MÉTODO VIEJO COMPLETAMENTE COMENTADO PARA REFERENCIA
// Este método implementa la lógica original con cálculos en Go y join con project_dollar_values
func (r *Repository) ListGroupLaborOld(ctx context.Context, inp types.Input, projectID int64, fieldID int64, usdMonth string) ([]domain.LaborRawItem, types.PageInfo, error) {
	// Usar la vista labor_list como base y agregar campos adicionales
	base := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v3").
		Select(`
				v3.workorder_id,
				v3.workorder_number,
				v3.date,
				v3.project_id,
				v3.field_id,
				v3.project_name,
				v3.field_name,
				COALESCE(v3.crop_name, '') AS crop_name,
				v3.labor_name,
				COALESCE(v3.labor_category_name, '') AS category_name,
				v3.contractor,
				v3.surface_ha,
				v3.cost_per_ha,
				v3.contractor_name,
				COALESCE(v3.investor_name, '') AS investor_name,
				pdv.average_value AS usd_avg_value,
				i.id AS invoice_id,
				i.number AS invoice_number,
				i.company AS invoice_company,
				i.date AS invoice_date,
				i.status AS invoice_status
			`).
		Joins("LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id").
		Joins("INNER JOIN project_dollar_values pdv ON pdv.project_id = v3.project_id AND pdv.month = ? AND pdv.deleted_at IS NULL", usdMonth)

	if fieldID != 0 {
		base = base.Where("v3.field_id = ?", fieldID)
	} else if projectID != 0 {
		base = base.Where("v3.project_id = ?", projectID)
	} else {
		return nil, types.PageInfo{}, types.NewError(types.ErrValidation,
			"fieldID or projectID is required", nil)
	}

	var total int64
	countQuery := base.Session(&gorm.Session{})
	if err := countQuery.Select("COUNT(*)").Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count labors for workorder", err)
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	var rows []models.LaborListItem
	if err := base.Order("v3.workorder_number DESC").
		Limit(int(inp.PageSize)).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list grouped labors", err)
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Obtener porcentaje de IVA dinámicamente desde app_parameters
		ivaPercentage, err := r.getIVAPercentage(ctx)
		if err != nil {
			// Si hay error, usar valor por defecto y logear el error
			// TODO: Implementar logging apropiado
			ivaPercentage = decimal.NewFromFloat(1.105) // 10.5%
		}
		totalIVA := netTotal.Mul(ivaPercentage)

		usdCostHa := m.CostPerHa.Div(m.USDAvgValue)
		usdNetTotal := netTotal.Div(m.USDAvgValue)

		// Manejar InvoiceID de forma segura
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		list[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          usdCostHa, // TODO: invertir los nombres de las variables, se invirtio USDCostHA por CostHA
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        usdNetTotal, // TODO: invertir los nombres de las variables, se invirtio usdNetTotal por netTotal
			TotalIVA:        totalIVA,
			USDCostHa:       m.CostPerHa, // TODO: invertir los nombres de las variables, se invirtio USDCostHA por CostHA
			USDNetTotal:     netTotal,    // TODO: invertir los nombres de las variables, se invirtio usdNetTotal por netTotal
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// safeStringPtr convierte un string pointer a string seguro
func safeStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (r *Repository) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	var row struct {
		SurfaceHa    decimal.Decimal `gorm:"column:surface_ha"`
		NetTotalCost decimal.Decimal `gorm:"column:total_labor_cost"`
		AvgCostPerHa decimal.Decimal `gorm:"column:avg_labor_cost_per_ha"`
	}

	// Caso 1: project_id Y field_id → devolver métricas de un campo específico
	if f.ProjectID != nil && f.FieldID != nil {
		q := fmt.Sprintf(`
			SELECT 
				surface_ha,
				total_labor_cost,
				avg_labor_cost_per_ha
			FROM %s
			WHERE project_id = ? AND field_id = ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, *f.ProjectID, *f.FieldID).Scan(&row).Error; err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to get labor metrics", err)
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Caso 2: SOLO project_id (sin field_id) → sumar métricas de todos los campos del proyecto
	if f.ProjectID != nil {
		q := fmt.Sprintf(`
			SELECT 
				COALESCE(SUM(surface_ha), 0) as surface_ha,
				COALESCE(SUM(total_labor_cost), 0) as total_labor_cost,
				CASE 
					WHEN COALESCE(SUM(surface_ha), 0) > 0 
					THEN COALESCE(SUM(total_labor_cost), 0) / COALESCE(SUM(surface_ha), 0)
					ELSE 0 
				END as avg_labor_cost_per_ha
			FROM %s
			WHERE project_id = ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, *f.ProjectID).Scan(&row).Error; err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to get labor metrics", err)
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Caso 3: SOLO field_id → devolver métricas de ese campo específico
	if f.FieldID != nil {
		q := fmt.Sprintf(`
			SELECT 
				surface_ha,
				total_labor_cost,
				avg_labor_cost_per_ha
			FROM %s
			WHERE field_id = ?
		`, shareddb.ReportView("labor_metrics"))
		if err := r.db.Client().WithContext(ctx).Raw(q, *f.FieldID).Scan(&row).Error; err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to get labor metrics", err)
		}

		return &domain.LaborMetrics{
			SurfaceHa:    row.SurfaceHa,
			NetTotalCost: row.NetTotalCost,
			AvgCostPerHa: row.AvgCostPerHa,
		}, nil
	}

	// Si no hay filtros, devolver ceros
	return &domain.LaborMetrics{
		SurfaceHa:    decimal.Zero,
		NetTotalCost: decimal.Zero,
		AvgCostPerHa: decimal.Zero,
	}, nil
}

func (r *Repository) ListAllGroupLabor(ctx context.Context) ([]domain.LaborRawItem, error) {
	base := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v3").
		Select(`
            v3.workorder_id,
			v3.workorder_number,
			v3.date,
			v3.project_id,
			v3.field_id,
			v3.project_name,
			v3.field_name,
			COALESCE(v3.crop_name, '') AS crop_name,
			v3.labor_name,
			COALESCE(v3.labor_category_name, '') AS category_name,
			v3.contractor,
			v3.surface_ha,
			v3.cost_per_ha,
			v3.contractor_name,
			COALESCE(v3.investor_name, '') AS investor_name,
			v3.dollar_average_month,
			v3.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
        `).
		Joins(`LEFT JOIN invoices i ON i.work_order_id = v3.workorder_id AND i.deleted_at IS NULL`)

	var rows []models.LaborListItem

	if err := base.Order("v3.workorder_number DESC").Scan(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list grouped labors", err)
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Usar porcentaje de IVA por defecto (10.5%)
		// TODO: Implementar obtención dinámica desde app_parameters
		ivaPercentage := decimal.NewFromFloat(0.105) // 10.5%
		totalIVA := netTotal.Mul(ivaPercentage)

		usd := m.USDAvgValue
		var usdCostHa, usdNetTotal decimal.Decimal
		if usd.GreaterThan(decimal.Zero) {
			usdCostHa = m.CostPerHa.Div(usd)
			usdNetTotal = netTotal.Div(usd)
		} else {
			// usd <= 0 dejamos USD en 0 (sin dividir)
			usdCostHa = decimal.Zero
			usdNetTotal = decimal.Zero
		}

		// Manejar campos opcionales
		cropName := ""
		if m.CropName != nil {
			cropName = *m.CropName
		}
		categoryName := ""
		if m.LaborCategoryName != nil {
			categoryName = *m.LaborCategoryName
		}
		investorName := ""
		if m.InvestorName != nil {
			investorName = *m.InvestorName
		}
		invoiceNumber := ""
		if m.InvoiceNumber != nil {
			invoiceNumber = *m.InvoiceNumber
		}
		invoiceCompany := ""
		if m.InvoiceCompany != nil {
			invoiceCompany = *m.InvoiceCompany
		}
		invoiceStatus := ""
		if m.InvoiceStatus != nil {
			invoiceStatus = *m.InvoiceStatus
		}

		list[i] = domain.LaborRawItem{
			WorkorderID:     m.WorkorderID,
			WorkorderNumber: m.WorkorderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        cropName,
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    categoryName,
			InvestorName:    investorName,
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,
			TotalIVA:        totalIVA,
			USDCostHa:       usdCostHa,
			USDNetTotal:     usdNetTotal,
			InvoiceID:       0,
			InvoiceNumber:   invoiceNumber,
			InvoiceCompany:  invoiceCompany,
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   invoiceStatus,
		}
	}
	return list, nil
}
