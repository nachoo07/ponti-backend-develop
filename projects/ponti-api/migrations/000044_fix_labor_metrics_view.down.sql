-- =======================
-- ROLLBACK MIGRACIÓN 000044: RESTAURAR LABOR_METRICS_VIEW ORIGINAL
-- =======================

-- Restaurar la vista original de la migración 000043
CREATE OR REPLACE VIEW labor_metrics_view AS
WITH
-- =======================
-- BASE DE ÓRDENES DE TRABAJO (pre-agregada para evitar duplicados)
-- =======================
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    w.labor_id,
    -- Pre-calcular costo de labor por hectárea
    COALESCE(lb.price, 0) AS labor_price_per_ha,
    -- Pre-calcular costo de labor por workorder
    COALESCE(lb.price, 0) * w.effective_area AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL 
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),

-- =======================
-- INSUMOS POR WORKORDER (evitar duplicados por supplies)
-- =======================
supply_costs AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    -- Costo total de insumos por workorder (dose * price * effective_area)
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0) * w.effective_area) AS total_supplies_cost,
    -- Costo de insumos por hectárea
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0)) AS supplies_cost_per_ha
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.id, w.project_id, w.field_id
),

-- =======================
-- MÉTRICAS POR PROYECTO/CAMPO (agregación final)
-- =======================
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    -- Superficie total (hectáreas)
    SUM(wb.effective_area) AS total_surface_ha,
    -- Costos desglosados
    SUM(wb.labor_cost_per_wo) AS total_labor_cost,
    SUM(COALESCE(sc.total_supplies_cost, 0)) AS total_supplies_cost,
    -- Costo total por hectárea (labor + supplies)
    SUM(wb.labor_price_per_ha + COALESCE(sc.supplies_cost_per_ha, 0)) AS total_cost_per_ha,
    -- Costo total neto
    SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0)) AS total_net_cost,
    -- Métricas calculadas
    COUNT(DISTINCT wb.workorder_id) AS total_workorders,
    -- Costo promedio por hectárea
    CASE 
      WHEN SUM(wb.effective_area) > 0 
      THEN SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0)) / SUM(wb.effective_area)
      ELSE 0 
    END AS avg_cost_per_ha_calculated,
    -- Porcentaje de costos por tipo
    CASE 
      WHEN SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0)) > 0
      THEN (SUM(wb.labor_cost_per_wo) / SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0))) * 100
      ELSE 0 
    END AS labor_cost_percentage,
    CASE 
      WHEN SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0)) > 0
      THEN (SUM(COALESCE(sc.total_supplies_cost, 0)) / SUM(wb.labor_cost_per_wo + COALESCE(sc.total_supplies_cost, 0))) * 100
      ELSE 0 
    END AS supplies_cost_percentage
  FROM workorder_base wb
  LEFT JOIN supply_costs sc ON sc.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id
)

SELECT
  fm.project_id,
  fm.field_id,
  -- COMPATIBILIDAD: Mantener nombres originales para el código Go
  fm.total_surface_ha AS surface_ha,                    -- ← NOMBRE ORIGINAL
  fm.total_net_cost AS net_total_cost,                  -- ← NOMBRE ORIGINAL
  fm.avg_cost_per_ha_calculated AS avg_cost_per_ha,     -- ← NOMBRE ORIGINAL
  -- Campos adicionales optimizados
  fm.total_labor_cost,
  fm.total_supplies_cost,
  fm.total_cost_per_ha,
  fm.total_workorders,
  fm.labor_cost_percentage,
  fm.supplies_cost_percentage,
  -- Indicadores de eficiencia
  CASE 
    WHEN fm.total_surface_ha > 0 AND fm.total_workorders > 0
    THEN fm.total_surface_ha / fm.total_workorders
    ELSE 0 
  END AS avg_surface_per_workorder,
  -- Densidad de costos por hectárea
  CASE 
    WHEN fm.total_surface_ha > 0
    THEN fm.total_labor_cost / fm.total_surface_ha
    ELSE 0 
  END AS labor_cost_per_ha,
  CASE 
    WHEN fm.total_surface_ha > 0
    THEN fm.total_supplies_cost / fm.total_surface_ha
    ELSE 0 
  END AS supplies_cost_per_ha
FROM field_metrics fm
WHERE fm.total_surface_ha > 0; -- Solo campos con superficie válida
