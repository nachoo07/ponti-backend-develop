-- ========================================
-- MIGRATION 000156: FIX operating_result calculation in v3_lot_metrics (UP)
-- ========================================
--
-- Purpose: Corregir SOLO el cálculo de operating_result_per_ha en v3_lot_metrics
-- Problem: operating_result_per_ha usaba v3_lot_ssot.operating_result_per_ha_for_lot()
--          que recalculaba active_total_per_ha con valores diferentes a los mostrados
-- Solution: Calcular operating_result directamente: income_net_per_ha - active_total_per_ha
--          usando los MISMOS valores que se muestran en las columnas de la vista
-- Date: 2025-10-21
-- Author: System
--
-- Note: Code in English, comments in Spanish.
-- CAMBIO QUIRÚRGICO: Solo líneas 123 y 136 de la vista v3_lot_metrics

BEGIN;

-- ========================================
-- RECREAR v3_lot_metrics CON FIX QUIRÚRGICO
-- ========================================
DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

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
  
  -- ############### FIX QUIRÚRGICO - LÍNEA 123 ###############
  -- ANTES: COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,
  -- AHORA: Calcular directamente usando los valores mostrados en la vista
  (v3_core_ssot.per_ha(
     COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
     b.hectares::numeric
   ) - COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)
  )::numeric AS operating_result_per_ha_usd,

  (COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS rent_total_usd,
  (COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS admin_total_usd,
  (COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS active_total_usd,
  
  -- ############### FIX QUIRÚRGICO - LÍNEA 136 ###############
  -- ANTES: (COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS operating_result_total_usd,
  -- AHORA: Multiplicar el operating_result_per_ha calculado arriba
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

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas agregadas por lote - operating_result calculado como income_net_per_ha - active_total_per_ha';

-- ========================================
-- RECREAR v3_lot_list (eliminada por CASCADE)
-- ========================================
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

COMMIT;

