-- ========================================
-- MIGRACIÓN 000158: FIX OPERATING RESULT FINAL (DOWN)
-- ========================================
--
-- Propósito: Revertir los cambios de la migración 158
-- Fecha: 2025-10-21
-- Autor: Sistema

BEGIN;

-- Revertir a la versión de la migración 156
DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

-- Recrear v3_lot_metrics como estaba en la migración 156
CREATE OR REPLACE VIEW public.v3_lot_metrics AS
WITH base AS (
  SELECT
    f.project_id,
    l.id              AS lot_id,
    l.name            AS lot_name,
    l.hectares,
    l.tons,
    l.sowing_date,
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0)::numeric AS sowed_area_ha,
    COALESCE(SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END), 0)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  GROUP BY f.project_id, l.id, l.name, l.hectares, l.tons, l.sowing_date
),
workorder_costs AS (
  SELECT
    lot_id,
    COALESCE(labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(direct_cost_usd, 0)::numeric AS direct_cost_usd
  FROM v3_workorder_metrics
),
project_total_direct_cost AS (
  SELECT
    p.id AS project_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares,
    COALESCE(SUM(wc.direct_cost_usd), 0)::numeric AS total_direct_cost
  FROM public.projects p
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorder_costs wc ON wc.lot_id = l.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
)
SELECT
  b.project_id,
  b.lot_id,
  b.lot_name,
  b.hectares,
  b.sowed_area_ha,
  b.harvested_area_ha,
  v3_lot_ssot.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
  b.tons,
  b.sowing_date,
  COALESCE(wc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(wc.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  COALESCE(wc.direct_cost_usd, 0)::numeric AS direct_cost_usd,
  COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric AS income_net_total_usd,
  v3_core_ssot.per_ha(
    COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
    b.hectares::numeric
  ) AS income_net_per_ha_usd,
  COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
  COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_cost_per_ha_usd,
  COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)::numeric AS active_total_per_ha_usd,
  (v3_core_ssot.per_ha(
     COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
     b.hectares::numeric
   ) - COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)
  )::numeric AS operating_result_per_ha_usd,
  (COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS rent_total_usd,
  (COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS admin_total_usd,
  (COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS active_total_usd,
  ((v3_core_ssot.per_ha(
      COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
      b.hectares::numeric
    ) - COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)
   ) * b.hectares::numeric
  )::numeric AS operating_result_total_usd,
  v3_core_ssot.cost_per_ha(
    COALESCE(wc.direct_cost_usd, 0)::numeric,
    COALESCE(b.sowed_area_ha, 0)::numeric
  ) AS direct_cost_per_ha_usd,
  COALESCE(ptdc.total_hectares, 0)::numeric AS project_total_hectares
FROM base b
LEFT JOIN workorder_costs wc ON wc.lot_id = b.lot_id
LEFT JOIN project_total_direct_cost ptdc ON ptdc.project_id = b.project_id;

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas agregadas por lote - operating_result calculado directamente desde valores de la vista';

-- Recrear v3_lot_list
CREATE OR REPLACE VIEW public.v3_lot_list AS
WITH base AS (
  SELECT
    f.project_id,
    p.name AS project_name,
    f.id AS field_id,
    f.name AS field_name,
    l.id AS lot_id,
    l.name AS lot_name,
    l.variety,
    l.season,
    l.previous_crop_id,
    prev_crop.name AS previous_crop,
    l.current_crop_id,
    curr_crop.name AS current_crop,
    l.hectares,
    l.updated_at,
    l.sowing_date,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
  LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
wo_dates AS (
  SELECT
    w.lot_id,
    MIN(w.date) AS raw_sowing_date
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
  GROUP BY w.lot_id
),
lot_metrics_data AS (
  SELECT
    project_id,
    lot_id,
    sowed_area_ha,
    harvested_area_ha,
    yield_tn_per_ha,
    direct_cost_per_ha_usd,
    direct_cost_usd,
    income_net_total_usd,
    income_net_per_ha_usd,
    rent_per_ha_usd,
    admin_cost_per_ha_usd,
    operating_result_per_ha_usd,
    rent_total_usd,
    admin_total_usd,
    operating_result_total_usd
  FROM v3_lot_metrics
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
  lm.sowed_area_ha,
  lm.harvested_area_ha,
  lm.yield_tn_per_ha,
  lm.direct_cost_per_ha_usd::numeric AS cost_usd_per_ha,
  lm.income_net_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  (COALESCE(lm.direct_cost_per_ha_usd, 0) 
   + COALESCE(lm.rent_per_ha_usd, 0) 
   + COALESCE(lm.admin_cost_per_ha_usd, 0))::numeric AS active_total_per_ha_usd,
  lm.operating_result_per_ha_usd,
  lm.income_net_total_usd,
  lm.direct_cost_usd AS direct_cost_total_usd,
  lm.rent_total_usd,
  lm.admin_total_usd,
  ((COALESCE(lm.direct_cost_per_ha_usd, 0) 
    + COALESCE(lm.rent_per_ha_usd, 0) 
    + COALESCE(lm.admin_cost_per_ha_usd, 0)) 
   * b.hectares)::numeric AS active_total_usd,
  lm.operating_result_total_usd,
  b.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  b.tons,
  wd.raw_sowing_date
FROM base b
LEFT JOIN wo_dates wd ON wd.lot_id = b.lot_id
LEFT JOIN lot_metrics_data lm ON lm.lot_id = b.lot_id;

COMMENT ON VIEW public.v3_lot_list IS 'Listado de lotes - usa costo específico de cada lote desde v3_lot_metrics';

-- Recrear v3_dashboard_metrics
CREATE OR REPLACE VIEW public.v3_dashboard_metrics AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    hectares,
    sowed_area_ha,
    harvested_area_ha,
    direct_cost_per_ha_usd,
    lot_id
  FROM public.v3_lot_metrics
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  COALESCE(
    SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  (p.admin_cost * 10)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0),
      0
    )::numeric,
    (p.admin_cost * 10)::numeric
  ) AS costs_progress_pct,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(lm.lot_id)), 0) AS operating_result_income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
   COALESCE(p.admin_cost * ph.total_hectares, 0) +
   COALESCE((SELECT f.lease_type_value * ph.total_hectares
             FROM fields f
             WHERE f.project_id = p.id AND f.deleted_at IS NULL
             LIMIT 1), 0))::double precision AS operating_result_total_costs_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
     COALESCE(p.admin_cost * ph.total_hectares, 0) +
     COALESCE((SELECT f.lease_type_value * ph.total_hectares
               FROM fields f
               WHERE f.project_id = p.id AND f.deleted_at IS NULL
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct,
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 'Métricas consolidadas del dashboard';

COMMIT;

