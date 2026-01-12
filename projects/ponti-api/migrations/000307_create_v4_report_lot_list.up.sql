-- =============================================================================
-- MIGRACIÓN 000307: v4_report.lot_list - Paridad exacta con v3_lot_list
-- =============================================================================
--
-- Propósito: Lista de lotes con métricas calculadas
-- Fuente: 000202_fix_lot_metrics_field_total_surface.up.sql (líneas 124-170)
--
-- Dependencias:
--   - v4_report.lot_metrics (en lugar de v3_lot_metrics)
--
-- FASE 1: Paridad exacta con v3_lot_list
--

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
  lm.sowed_area_ha,
  lm.harvested_area_ha,
  lm.yield_tn_per_ha,
  lm.direct_cost_per_ha_usd::numeric AS cost_usd_per_ha,
  lm.income_net_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  lm.active_total_per_ha_usd,
  lm.operating_result_per_ha_usd,
  lm.income_net_total_usd,
  lm.direct_cost_usd AS direct_cost_total_usd,
  lm.rent_total_usd,
  lm.admin_total_usd,
  lm.active_total_usd,
  lm.operating_result_total_usd,
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
'Paridad exacta con v3_lot_list (000202). FASE 1: usa v4_report.lot_metrics.';
