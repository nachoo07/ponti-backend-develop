-- ========================================
-- MIGRACIÓN 000205: RECREATE v3_dashboard_metrics after 000202 (UP)
-- ========================================
--
-- Propósito: Recrear v3_dashboard_metrics que fue eliminada por CASCADE en migración 000202
-- Problema: La migración 000202 hizo DROP CASCADE de v3_lot_metrics pero no recreó v3_dashboard_metrics
-- Solución: Recrear la vista con la definición correcta (incluye planned_cost de 000196)
-- Fecha: 2025-11-18
-- Autor: Sistema
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR: v3_dashboard_metrics
-- (Se eliminó por CASCADE al dropear v3_lot_metrics en 000202)
-- ========================================

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
  
  -- CARD 1: AVANCE DE SIEMBRA
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- CARD 2: AVANCE DE COSECHA
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- CARD 3: AVANCE DE COSTOS
  -- Costos ejecutados: promedio ponderado de costo directo por ha sembrada
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  -- Presupuesto: usar planned_cost (FIX 000196)
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  -- CARD 4: RESULTADO OPERATIVO
  -- Ingresos netos totales (suma de todos los lotes del proyecto)
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  -- Resultado operativo total (usa función SSOT del dashboard)
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  -- Costos totales: directos + admin + arriendo fijo
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
  -- Porcentaje de renta (resultado / costos totales * 100)
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
  
  -- Total de hectáreas del proyecto
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 
'Dashboard metrics. Recreada en 000205 tras CASCADE de 000202. Usa planned_cost (000196) y rent_fixed_only (000199).';

COMMIT;

