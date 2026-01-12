-- ========================================
-- MIGRATION 000196: FIX Dashboard Metrics Budget Use Planned Cost (UP)
-- ========================================
--
-- Propósito: Restaurar el uso de planned_cost en la card "Avance de costos".
-- Contexto: La migración 000189 recreó v3_dashboard_metrics y volvió a usar
--           (admin_cost * 10) como presupuesto, ignorando el valor cargado en
--           Clientes y Sociedades. Esta migración vuelve a planned_cost sin
--           perder los cambios introducidos en 000189 (rent_fixed_only_*).
--
-- Fecha: 2025-11-10
-- Autor: Sistema

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_metrics CASCADE;

CREATE VIEW public.v3_dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.sowed_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM public.v3_lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE((
      SELECT f.lease_type_value * ph.total_hectares
      FROM public.fields f
      WHERE f.project_id = p.id AND f.deleted_at IS NULL
      LIMIT 1
    ), 0)
  )::double precision AS operating_result_total_costs_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE((
        SELECT f.lease_type_value * ph.total_hectares
        FROM public.fields f
        WHERE f.project_id = p.id AND f.deleted_at IS NULL
        LIMIT 1
      ), 0)
    )::double precision
  ) AS operating_result_pct,
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 'Dashboard metrics. FIX 000196: budget_cost_usd usa planned_cost y mantiene rent_fixed_only().';

COMMIT;

