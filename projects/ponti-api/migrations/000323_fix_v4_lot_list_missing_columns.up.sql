-- =============================================================================
-- Migration: 000323_fix_v4_lot_list_missing_columns
-- Description: Arregla v4_calc.lot_base_costs que perdió columnas en 000318
-- Problema: 000318 recreó lot_base_costs sin sowed_area_ha, harvested_area_ha, etc
-- =============================================================================

-- Drop cascade para recrear toda la cadena
DROP VIEW IF EXISTS v4_report.lot_list CASCADE;
DROP VIEW IF EXISTS v4_report.lot_metrics CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_income CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_costs CASCADE;

-- =============================================================================
-- 1. Recrear v4_calc.lot_base_costs con TODAS las columnas necesarias
-- =============================================================================
CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH 
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0) AS hectares,
    COALESCE(l.tons, 0) AS tons,
    l.sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
areas AS (
  SELECT
    r.lot_id,
    v4_ssot.seeded_area_for_lot(r.lot_id) AS sowed_area_ha,
    v4_ssot.harvested_area_for_lot(r.lot_id) AS harvested_area_ha
  FROM raw r
),
costs AS (
  SELECT
    lot_id,
    MAX(COALESCE(labor_cost_usd, 0))::numeric AS labor_cost_usd,
    MAX(COALESCE(supplies_cost_usd, 0))::numeric AS supplies_cost_usd,
    MAX(COALESCE(direct_cost_usd, 0))::numeric AS direct_cost_usd
  FROM public.v3_workorder_metrics
  GROUP BY lot_id
),
ssot_values AS (
  SELECT
    r.lot_id,
    v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha,
    v4_ssot.income_net_total_for_lot(r.lot_id) AS income_net_total_usd,
    -- rent_per_ha = TOTAL (fijo + %) - para mostrar y calcular
    v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_per_ha_usd,
    -- rent_fixed = solo fijo (para compatibilidad)
    v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_fixed_per_ha_usd,
    v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd
  FROM raw r
),
derived AS (
  SELECT
    r.project_id, r.field_id, r.current_crop_id, r.lot_id, r.lot_name,
    r.hectares, r.tons, r.sowing_date,
    COALESCE(a.sowed_area_ha, 0)::numeric AS sowed_area_ha,
    COALESCE(a.harvested_area_ha, 0)::numeric AS harvested_area_ha,
    COALESCE(c.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(c.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(c.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    COALESCE(s.yield_tn_per_ha, 0) AS yield_tn_per_ha,
    COALESCE(s.income_net_total_usd, 0)::numeric AS income_net_total_usd,
    COALESCE(s.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
    COALESCE(s.rent_fixed_per_ha_usd, 0)::numeric AS rent_fixed_per_ha_usd,
    COALESCE(s.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd
  FROM raw r
  LEFT JOIN areas a ON a.lot_id = r.lot_id
  LEFT JOIN costs c ON c.lot_id = r.lot_id
  LEFT JOIN ssot_values s ON s.lot_id = r.lot_id
)
SELECT
  d.project_id, d.field_id, d.current_crop_id, d.lot_id, d.lot_name,
  d.hectares, d.tons, d.sowing_date,
  d.sowed_area_ha,
  d.harvested_area_ha,
  d.yield_tn_per_ha,
  d.labor_cost_usd,
  d.supplies_cost_usd,
  d.direct_cost_usd,
  d.income_net_total_usd,
  v4_core.per_ha(d.income_net_total_usd, d.hectares::numeric) AS income_net_per_ha_usd,
  v4_core.per_ha(d.direct_cost_usd, d.sowed_area_ha) AS direct_cost_per_ha_usd,
  d.rent_per_ha_usd,
  d.rent_fixed_per_ha_usd,
  d.admin_cost_per_ha_usd
FROM derived d;

COMMENT ON VIEW v4_calc.lot_base_costs IS 
'FIX 000323: Restaura columnas perdidas en 000318 (sowed_area_ha, harvested_area_ha, etc).';

-- =============================================================================
-- 2. Recrear v4_calc.lot_base_income
-- =============================================================================
CREATE OR REPLACE VIEW v4_calc.lot_base_income AS
SELECT
  c.project_id, c.field_id, c.current_crop_id, c.lot_id, c.lot_name,
  c.hectares,
  c.tons,
  c.sowed_area_ha,
  c.yield_tn_per_ha,
  COALESCE(v4_ssot.net_price_usd_for_lot(c.lot_id), 0)::numeric AS net_price_usd_tn,
  c.income_net_total_usd,
  c.income_net_per_ha_usd
FROM v4_calc.lot_base_costs c;

COMMENT ON VIEW v4_calc.lot_base_income IS 'FIX 000323: Recreada tras fix de lot_base_costs.';

-- =============================================================================
-- 3. Recrear v4_report.lot_metrics
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.lot_metrics AS
WITH 
project_totals AS (
  SELECT project_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY project_id
),
field_totals AS (
  SELECT field_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY field_id
)
SELECT
  c.project_id,
  c.field_id,
  c.lot_id,
  c.lot_name,
  c.hectares,
  c.sowed_area_ha,
  c.harvested_area_ha,
  c.yield_tn_per_ha,
  c.tons,
  c.sowing_date,
  c.labor_cost_usd,
  c.supplies_cost_usd AS supply_cost_usd,
  c.direct_cost_usd,
  c.income_net_total_usd,
  c.income_net_per_ha_usd,
  c.rent_per_ha_usd,
  c.admin_cost_per_ha_usd,
  c.direct_cost_per_ha_usd,
  -- Calculados
  (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  (c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  (c.rent_per_ha_usd * c.hectares)::numeric AS rent_total_usd,
  (c.admin_cost_per_ha_usd * c.hectares)::numeric AS admin_total_usd,
  ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd) * c.hectares)::numeric AS active_total_usd,
  ((c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)) * c.hectares)::numeric AS operating_result_total_usd,
  c.direct_cost_usd AS direct_cost_total_usd,
  COALESCE(pt.total_hectares, 0) AS project_total_hectares,
  COALESCE(ft.total_hectares, 0) AS field_total_hectares
FROM v4_calc.lot_base_costs c
LEFT JOIN project_totals pt ON pt.project_id = c.project_id
LEFT JOIN field_totals ft ON ft.field_id = c.field_id;

COMMENT ON VIEW v4_report.lot_metrics IS 
'FIX 000323: Restaurada tras fix de lot_base_costs.';

-- =============================================================================
-- 4. Recrear v4_report.lot_list
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.lot_list AS
SELECT
  f.project_id,
  p.name AS project_name,
  f.id AS field_id,
  f.name AS field_name,
  l.id AS id,
  l.name AS lot_name,
  l.variety,
  l.season,
  l.previous_crop_id,
  prev_crop.name AS previous_crop,
  l.current_crop_id,
  curr_crop.name AS current_crop,
  l.hectares,
  l.updated_at,
  COALESCE(lm.sowed_area_ha, 0)::numeric AS sowed_area_ha,
  COALESCE(lm.harvested_area_ha, 0)::numeric AS harvested_area_ha,
  COALESCE(lm.yield_tn_per_ha, 0) AS yield_tn_per_ha,
  COALESCE(lm.direct_cost_per_ha_usd, 0)::numeric AS cost_usd_per_ha,
  COALESCE(lm.income_net_per_ha_usd, 0)::numeric AS income_net_per_ha_usd,
  COALESCE(lm.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
  COALESCE(lm.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd,
  COALESCE(lm.active_total_per_ha_usd, 0)::numeric AS active_total_per_ha_usd,
  COALESCE(lm.operating_result_per_ha_usd, 0)::numeric AS operating_result_per_ha_usd,
  COALESCE(lm.income_net_total_usd, 0)::numeric AS income_net_total_usd,
  COALESCE(lm.direct_cost_total_usd, 0)::numeric AS direct_cost_total_usd,
  COALESCE(lm.rent_total_usd, 0)::numeric AS rent_total_usd,
  COALESCE(lm.admin_total_usd, 0)::numeric AS admin_total_usd,
  COALESCE(lm.active_total_usd, 0)::numeric AS active_total_usd,
  COALESCE(lm.operating_result_total_usd, 0)::numeric AS operating_result_total_usd,
  l.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  l.tons,
  (
    SELECT MIN(w.date)
    FROM public.workorders w
    JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
    WHERE w.lot_id = l.id AND w.deleted_at IS NULL
  ) AS raw_sowing_date
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
LEFT JOIN v4_report.lot_metrics lm ON lm.lot_id = l.id
WHERE l.deleted_at IS NULL;

COMMENT ON VIEW v4_report.lot_list IS 
'FIX 000323: Paridad con v3_lot_list. Todas las columnas restauradas.';
