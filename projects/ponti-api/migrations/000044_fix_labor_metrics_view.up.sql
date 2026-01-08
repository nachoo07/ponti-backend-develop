DROP VIEW IF EXISTS labor_metrics_view CASCADE;

CREATE VIEW labor_metrics_view AS
WITH
-- =======================
-- BASE DE ÓRDENES DE TRABAJO (exigiendo lb.price IS NOT NULL)
-- =======================
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    w.labor_id,
    -- Exigir precio de labor desde BD (NO COALESCE a 0)
    lb.price AS labor_price_per_ha,
    -- Costo de labor por workorder
    lb.price * w.effective_area AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL 
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL  -- ← EXIGIR PRECIO DESDE BD
),

-- =======================
-- CONSUMO Y COSTOS DE INSUMOS POR WORKORDER
-- =======================
supply_consumption AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    -- Cantidad usada según regla: COALESCE(total_used, final_dose * effective_area, 0)
    COALESCE(wi.total_used, wi.final_dose * w.effective_area, 0) AS qty_used,
    -- Mapeo directo de unit_id a categorías (versión simplificada)
    CASE 
      WHEN s.unit_id = 1 THEN 'LITERS'  -- Litros
      WHEN s.unit_id = 2 THEN 'KILOS'   -- Kilos
      ELSE 'OTHER'
    END AS unit_category,
    -- Costo del insumo por workorder
    COALESCE(s.price, 0) * COALESCE(wi.total_used, wi.final_dose * w.effective_area, 0) AS supply_cost_per_wo
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),

-- =======================
-- AGREGACIÓN DE CONSUMOS POR WORKORDER
-- =======================
workorder_metrics AS (
  SELECT
    wb.workorder_id,
    wb.project_id,
    wb.field_id,
    wb.effective_area,
    wb.labor_cost_per_wo,
    -- Consumos por tipo de unidad
    SUM(CASE WHEN sc.unit_category = 'LITERS' THEN COALESCE(sc.qty_used, 0) ELSE 0 END) AS liters_used,
    SUM(CASE WHEN sc.unit_category = 'KILOS' THEN COALESCE(sc.qty_used, 0) ELSE 0 END) AS kilos_used,
    -- Costos totales
    SUM(COALESCE(sc.supply_cost_per_wo, 0)) AS total_supplies_cost,
    -- Costo total del workorder
    wb.labor_cost_per_wo + SUM(COALESCE(sc.supply_cost_per_wo, 0)) AS total_workorder_cost
  FROM workorder_base wb
  LEFT JOIN supply_consumption sc ON sc.workorder_id = wb.workorder_id
  GROUP BY wb.workorder_id, wb.project_id, wb.field_id, wb.effective_area, wb.labor_cost_per_wo
),

-- =======================
-- MÉTRICAS FINALES POR PROYECTO/CAMPO
-- =======================
field_metrics AS (
  SELECT
    wm.project_id,
    wm.field_id,
    -- Superficie total (hectáreas)
    SUM(wm.effective_area) AS surface_ha,
    -- Consumos totales
    SUM(wm.liters_used) AS total_liters,
    SUM(wm.kilos_used) AS total_kilos,
    -- Costos desglosados
    SUM(wm.labor_cost_per_wo) AS total_labor_cost,
    SUM(wm.total_supplies_cost) AS total_supplies_cost,
    -- Costo total neto
    SUM(wm.total_workorder_cost) AS net_total_cost,
    -- Métricas adicionales
    COUNT(DISTINCT wm.workorder_id) AS total_workorders,
    -- Costo promedio por hectárea (ponderado)
    CASE 
      WHEN SUM(wm.effective_area) > 0 
      THEN SUM(wm.total_workorder_cost) / SUM(wm.effective_area)
      ELSE 0 
    END AS avg_cost_per_ha,
    -- Consumos por hectárea
    CASE 
      WHEN SUM(wm.effective_area) > 0 
      THEN SUM(wm.liters_used) / SUM(wm.effective_area)
      ELSE 0 
    END AS liters_per_ha,
    CASE 
      WHEN SUM(wm.effective_area) > 0 
      THEN SUM(wm.kilos_used) / SUM(wm.effective_area)
      ELSE 0 
    END AS kilos_per_ha
  FROM workorder_metrics wm
  GROUP BY wm.project_id, wm.field_id
)

SELECT
  fm.project_id,
  fm.field_id,
  -- Campos principales (compatibilidad con backend)
  fm.surface_ha,
  fm.total_liters,
  fm.total_kilos,
  fm.net_total_cost,
  fm.avg_cost_per_ha,
  -- Campos adicionales
  fm.total_labor_cost,
  fm.total_supplies_cost,
  fm.total_workorders,
  fm.liters_per_ha,
  fm.kilos_per_ha
FROM field_metrics fm
WHERE fm.surface_ha > 0; -- Solo campos con superficie válida

-- =======================
-- ÍNDICES OPTIMIZADOS PARA LA VISTA CORREGIDA
-- =======================

-- Índices parciales para soft-delete (GCP optimizados)
CREATE INDEX IF NOT EXISTS idx_labor_workorders_metrics_v2 
  ON workorders(project_id, field_id, labor_id, effective_area) 
  WHERE deleted_at IS NULL AND effective_area > 0;

CREATE INDEX IF NOT EXISTS idx_labor_workorder_items_v2 
  ON workorder_items(workorder_id, supply_id, total_used, final_dose) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_labor_supplies_units_v2 
  ON supplies(id, price, unit_id) 
  WHERE deleted_at IS NULL;

-- Índice eliminado: ya no usamos tabla units

-- Índice compuesto para JOINs frecuentes (eliminado duplicado)
-- CREATE INDEX IF NOT EXISTS idx_labor_consumption_v2 
--   ON workorder_items(workorder_id, supply_id, total_used, final_dose) 
--   WHERE deleted_at IS NULL;
