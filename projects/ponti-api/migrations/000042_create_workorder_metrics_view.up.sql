-- ========================================
-- VISTA: workorder_metrics_view
-- ========================================
-- Propósito: Agregar las 4 métricas específicas requeridas por proyecto y campo
-- Incluye: Superficie, consumo en liters/kilograms, y costo directo
-- Filtros: Solo workorders activos con área efectiva válida
-- Optimizada para: Cloud SQL (GCP) con índices parciales
-- ========================================

-- Eliminar la vista existente primero
DROP VIEW IF EXISTS workorder_metrics_view;

-- Crear la nueva vista
CREATE VIEW workorder_metrics_view AS
WITH
-- =======================
-- BASE DE ÓRDENES DE TRABAJO (pre-agregada para evitar duplicados)
-- =======================
-- Pre-calcula costos de labor por workorder individual
-- Filtra workorders activos con área efectiva válida
-- =======================
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
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
-- INSUMOS CENTRALIZADOS (simplificado sin unidades)
-- =======================
-- Agrega costos de insumos por workorder
-- Calcula: dose * price * effective_area para cada workorder
-- Filtra items y supplies activos
-- =======================
supply_aggregation AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    -- Costo total de insumos por workorder (dose * price * effective_area)
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0) * w.effective_area) AS total_supplies_cost,
    -- Consumo en liters (final_dose * effective_area - placeholder hasta definir unidades)
    SUM(COALESCE(wi.final_dose, 0) * w.effective_area) AS total_liters,
    -- Consumo en kilograms (final_dose * effective_area - placeholder hasta definir unidades)
    SUM(COALESCE(wi.final_dose, 0) * w.effective_area) AS total_kilograms
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
-- Agrega métricas por proyecto y campo
-- Incluye: superficie, costos, consumo en liters y kilograms
-- Aplica filtros finales de validación
-- =======================
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    -- Superficie total (hectáreas)
    SUM(wb.effective_area) AS total_surface_ha,
    -- Costos
    SUM(wb.labor_cost_per_wo) AS total_labor_cost,
    SUM(COALESCE(sa.total_supplies_cost, 0)) AS total_supplies_cost,
    -- Costo directo total
    SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) AS total_direct_cost,
    -- Consumo en liters y kilograms
    SUM(COALESCE(sa.total_liters, 0)) AS total_liters,
    SUM(COALESCE(sa.total_kilograms, 0)) AS total_kilograms,
    -- Métricas calculadas
    COUNT(DISTINCT wb.workorder_id) AS total_workorders
  FROM workorder_base wb
  LEFT JOIN supply_aggregation sa ON sa.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id
)

SELECT
  fm.project_id,
  fm.field_id,
  -- LOS 4 VALORES ESPECÍFICOS REQUERIDOS:
  fm.total_surface_ha AS surface_ha,           -- 1. Superficie ejecutada total
  COALESCE(fm.total_liters, 0) AS liters,     -- 2. Consumo en liters total
  COALESCE(fm.total_kilograms, 0) AS kilograms, -- 3. Consumo en kilograms total
  fm.total_direct_cost AS direct_cost          -- 4. Costo directo total (labor + insumos)
FROM field_metrics fm
WHERE fm.total_surface_ha > 0; -- Solo campos con superficie válida

-- ========================================
-- RESUMEN DE LA VISTA
-- ========================================
-- Esta vista proporciona las 4 métricas específicas requeridas por proyecto y campo:
-- 
-- MÉTRICAS ESPECÍFICAS:
-- - surface_ha: Superficie ejecutada total (suma de effective_area de workorders)
-- - liters: Consumo total (final_dose * effective_area - placeholder hasta definir unidades)
-- - kilograms: Consumo total (final_dose * effective_area - placeholder hasta definir unidades)
-- - direct_cost: Costo directo total (labor + insumos)
-- 
-- CÁLCULOS:
-- - Los costos de labor se obtienen de la tabla labors (precio por hectárea)
-- - Los costos de insumos se calculan como: dose * precio * área efectiva
-- - El consumo en liters/kilograms se calcula como: dose * área efectiva (placeholder)
-- - NOTA: Para diferenciar unidades, se requiere implementar tabla de unidades
-- 
-- FILTROS APLICADOS:
-- - Solo workorders activos (deleted_at IS NULL)
-- - Solo con área efectiva válida (> 0)
-- - Solo campos con superficie total > 0
-- 
-- OPTIMIZACIONES:
-- - CTEs para evitar duplicados y mejorar performance
-- - Índices parciales para soft-delete
-- - Índices compuestos para JOINs frecuentes
-- - Compatible con Cloud SQL (GCP)
-- ========================================
