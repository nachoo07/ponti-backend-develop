-- ========================================
-- MIGRACIÓN 000186: FIX Dashboard Budget Use Planned Cost (UP)
-- ========================================
--
-- Propósito: Corregir budget_cost_usd para usar planned_cost de projects
-- Problema: Dashboard usa fórmula hardcodeada (admin_cost * 10) 
--           en lugar del costo planificado cargado en Clientes y Sociedades
-- Solución: Reemplazar (p.admin_cost * 10) por p.planned_cost
-- Fecha: 2025-11-08
-- Autor: Sistema
--
-- Impacto: Card "Avance de costos" usará el presupuesto real cargado
--
-- Note: Código en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_metrics CASCADE;

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
  
  -- CARD 1: AVANCE DE SIEMBRA
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- CARD 2: AVANCE DE COSECHA
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- ========================================
  -- CARD 3: AVANCE DE COSTOS
  -- FIX 000186: Usar planned_cost en lugar de admin_cost * 10
  -- ========================================
  COALESCE(
    SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
    0
  )::double precision AS executed_costs_usd,
  -- CORRECCIÓN: Usar planned_cost de la tabla projects
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
      0
    )::numeric,
    -- CORRECCIÓN: Usar planned_cost de la tabla projects
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  -- CARD 4: RESULTADO OPERATIVO
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
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 
'Métricas consolidadas del dashboard. FIX 000186: budget_cost_usd usa planned_cost (no hardcodeado).';

COMMIT;

