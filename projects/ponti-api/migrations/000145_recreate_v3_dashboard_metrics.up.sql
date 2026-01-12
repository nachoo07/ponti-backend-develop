-- ========================================
-- MIGRATION 000145: RECREATE V3_DASHBOARD_METRICS (UP)
-- ========================================
-- 
-- Purpose: Recrear v3_dashboard_metrics que fue eliminada en CASCADE por migración 000143
-- Date: 2025-10-14
-- Author: System
-- 
-- Context: La migración 000143 hizo DROP CASCADE de v3_lot_metrics, eliminando
--          también v3_dashboard_metrics que dependía de ella. Esta migración
--          la recrea con la misma definición que tenía en migración 000126.
--
-- Note: Code in English, comments in Spanish.

BEGIN;

-- ========================================
-- RECREAR VISTA v3_dashboard_metrics
-- ========================================
-- Esta vista consolida las 5 cards principales del dashboard
-- Depende de: v3_lot_metrics, v3_core_ssot, v3_dashboard_ssot, v3_lot_ssot

CREATE VIEW public.v3_dashboard_metrics AS
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
  
  -- ========================================
  -- CARD 1: AVANCE DE SIEMBRA
  -- ========================================
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- ========================================
  -- CARD 2: AVANCE DE COSECHA
  -- ========================================
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- ========================================
  -- CARD 3: AVANCE DE COSTOS
  -- ========================================
  -- Costo ejecutado (promedio ponderado por ha sembrada)
  COALESCE(
    SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
    0
  )::double precision AS executed_costs_usd,
  
  -- ⚠️ PRESUPUESTO HARDCODEADO TEMPORAL ⚠️
  -- TODO: Definir fórmula correcta para calcular el presupuesto dinámicamente
  -- Por ahora: admin_cost * 10 (valor temporal hasta definir cálculo)
  (p.admin_cost * 10)::double precision AS budget_cost_usd,
  
  -- Porcentaje de progreso
  v3_core_ssot.percentage(
    COALESCE(
      SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
      0
    )::numeric,
    (p.admin_cost * 10)::numeric
  ) AS costs_progress_pct,
  
  -- ========================================
  -- CARD 4: RESULTADO OPERATIVO
  -- ========================================
  -- Ingresos (suma de ingresos netos por lote)
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(lm.lot_id)), 0) AS operating_result_income_usd,
  
  -- Resultado operativo
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  
  -- Total costos (usa función SSOT, NO repetir cálculo)
  v3_dashboard_ssot.total_costs_for_project(p.id) AS operating_result_total_costs_usd,
  
  -- Porcentaje de rentabilidad (usa función SSOT para costos)
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  
  -- ========================================
  -- CAMPOS ADICIONALES (para compatibilidad)
  -- ========================================
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, ph.total_hectares;

-- Comentario de la vista
COMMENT ON VIEW public.v3_dashboard_metrics IS 'Métricas consolidadas del dashboard (5 cards en 1 vista) - Recreada después de eliminación CASCADE en migración 000143';

COMMIT;

