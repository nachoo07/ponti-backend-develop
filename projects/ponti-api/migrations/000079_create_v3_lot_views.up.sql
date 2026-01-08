-- ========================================
-- MIGRATION 000079: CREATE v3_lot_views (UP)
-- ========================================
-- 
-- Purpose: Create lot metrics and list views
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_lot_metrics: métricas por lote (costos, ingresos, rentas, etc.)
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_lot_metrics AS
WITH lot_base AS (
  SELECT
    l.id  AS lot_id,
    f.id  AS field_id,
    f.project_id,
    l.hectares,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric) AS sowed_area_ha,
    v3_calc.harvested_area(l.tons::numeric, l.hectares::numeric) AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
-- Superficie total del campo (suma de todos los lotes del campo)
field_total_area AS (
  SELECT 
    f.id AS field_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.id
)
SELECT
  b.project_id,
  b.field_id,
  b.lot_id,
  b.hectares,
  b.sowed_area_ha,
  b.harvested_area_ha,
  v3_calc.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
  COALESCE(v3_calc.labor_cost_for_lot(b.lot_id), 0)::numeric      AS labor_cost_usd,
  COALESCE(v3_calc.supply_cost_for_lot(b.lot_id), 0)::numeric     AS supplies_cost_usd,
  COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), 0)::numeric     AS direct_cost_usd,
  COALESCE(v3_calc.income_net_total_for_lot(b.lot_id), 0)::numeric    AS income_net_total_usd,
  COALESCE(v3_calc.income_net_per_ha_for_lot(b.lot_id), 0)::numeric   AS income_net_per_ha_usd,
  COALESCE(v3_calc.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric   AS admin_cost_per_ha_usd,
  COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0)::numeric        AS rent_per_ha_usd,
  COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0)::numeric AS active_total_per_ha_usd,
  COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,
  -- Totales
  (COALESCE(v3_calc.admin_cost_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric   AS admin_total_usd,
  (COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric        AS rent_total_usd,
  (COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric AS active_total_usd,
  (COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric AS operating_result_total_usd,
  -- Per-ha usando wrapper SSOT
  v3_calc.cost_per_ha(COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), 0)::numeric, b.hectares::numeric) AS direct_cost_per_ha_usd,
  -- Superficie total del campo
  COALESCE(fta.total_hectares, 0)::numeric AS superficie_total
FROM lot_base b
LEFT JOIN field_total_area fta ON fta.field_id = b.field_id;

-- -------------------------------------------------------------------
-- v3_lot_list: listado por lote con datos base y métricas
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_lot_list AS
WITH base AS (
  SELECT
    l.id   AS lot_id,
    l.name AS lot_name,
    l.variety,
    l.season,
    l.hectares,
    l.tons,
    l.sowing_date,
    l.updated_at,
    f.id   AS field_id,
    f.name AS field_name,
    f.project_id,
    p.name AS project_name,
    l.previous_crop_id,
    pc.name AS previous_crop,
    l.current_crop_id,
    cc.name AS current_crop
  FROM public.lots l
  JOIN public.fields   f  ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p  ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops pc ON pc.id = l.previous_crop_id AND pc.deleted_at IS NULL
  LEFT JOIN public.crops cc ON cc.id = l.current_crop_id AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
wo_dates AS (
  SELECT
    w.lot_id,
    MIN(CASE WHEN lb.category_id = 9  THEN w.date END) AS lot_sowing_date,
    MAX(CASE WHEN lb.category_id = 13 THEN w.date END) AS lot_harvest_date
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
  GROUP BY w.lot_id
)
SELECT
  b.project_id,
  b.project_name,
  b.field_id,
  b.field_name,
  b.lot_id AS id,
  b.lot_name,
  b.variety,
  b.season,
  b.previous_crop_id,
  b.previous_crop,
  b.current_crop_id,
  b.current_crop,
  b.hectares,
  b.updated_at,
  v3_calc.seeded_area_for_lot(b.lot_id)::numeric           AS sowed_area_ha,
  v3_calc.harvested_area_for_lot(b.lot_id)::numeric        AS harvested_area_ha,
  v3_calc.yield_tn_per_ha_for_lot(b.lot_id)               AS yield_tn_per_ha,
  v3_calc.cost_per_ha(COALESCE(v3_calc.direct_cost_for_lot(b.lot_id),0)::numeric, b.hectares::numeric) AS cost_usd_per_ha,
  v3_calc.income_net_per_ha_for_lot(b.lot_id)::numeric     AS income_net_per_ha_usd,
  v3_calc.rent_per_ha_for_lot(b.lot_id)::numeric           AS rent_per_ha_usd,
  v3_calc.admin_cost_per_ha_for_lot(b.lot_id)::numeric     AS admin_cost_per_ha_usd,
  v3_calc.active_total_per_ha_for_lot(b.lot_id)::numeric   AS active_total_per_ha_usd,
  v3_calc.operating_result_per_ha_for_lot(b.lot_id)::numeric AS operating_result_per_ha_usd,
  -- Totales
  v3_calc.income_net_total_for_lot(b.lot_id)::numeric      AS income_net_total_usd,
  COALESCE(v3_calc.direct_cost_for_lot(b.lot_id),0)::numeric AS direct_cost_total_usd,
  (v3_calc.rent_per_ha_for_lot(b.lot_id) * b.hectares)::numeric  AS rent_total_usd,
  (v3_calc.admin_cost_per_ha_for_lot(b.lot_id) * b.hectares)::numeric AS admin_total_usd,
  (v3_calc.active_total_per_ha_for_lot(b.lot_id) * b.hectares)::numeric AS active_total_usd,
  (v3_calc.operating_result_per_ha_for_lot(b.lot_id) * b.hectares)::numeric AS operating_result_total_usd,
  wd.lot_sowing_date,
  wd.lot_harvest_date,
  b.tons,
  b.sowing_date AS raw_sowing_date
FROM base b
LEFT JOIN wo_dates wd ON wd.lot_id = b.lot_id;
