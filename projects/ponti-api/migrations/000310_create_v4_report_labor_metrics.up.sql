-- =============================================================================
-- MIGRACIÓN 000310: v4_report.labor_metrics - Paridad exacta con v3_labor_metrics
-- =============================================================================
--
-- Propósito: Métricas agregadas de labores por proyecto/campo
-- Fuente: 000080_create_v3_labor_views.up.sql (líneas 21-59)
--
-- Dependencias:
--   - v3_calc.labor_cost (se mantiene para paridad)
--   - v3_calc.cost_per_ha (se mantiene para paridad)
--
-- FASE 1: Paridad exacta con v3_labor_metrics
--

CREATE OR REPLACE VIEW v4_report.labor_metrics AS
WITH wo AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.date,
    w.effective_area::numeric AS effective_area,
    lb.price::numeric AS labor_price_per_ha
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
),
agg AS (
  SELECT
    project_id,
    field_id,
    COUNT(DISTINCT workorder_id) AS total_workorders,
    SUM(effective_area) AS surface_ha,
    SUM(v3_calc.labor_cost(labor_price_per_ha, effective_area)) AS total_labor_cost,
    MIN(date) AS first_workorder_date,
    MAX(date) AS last_workorder_date
  FROM wo
  GROUP BY project_id, field_id
)
SELECT
  a.project_id,
  a.field_id,
  a.surface_ha,
  a.total_labor_cost,
  v3_calc.cost_per_ha(a.total_labor_cost, a.surface_ha) AS avg_labor_cost_per_ha,
  a.total_workorders,
  a.first_workorder_date,
  a.last_workorder_date
FROM agg a;

COMMENT ON VIEW v4_report.labor_metrics IS 
'Paridad exacta con v3_labor_metrics (000080). FASE 1: usa v3_calc.labor_cost/cost_per_ha.';
