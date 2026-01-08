-- ========================================
-- MIGRACIÓN 000121: CREATE DASHBOARD BASE VIEWS (UP)
-- ========================================
-- 
-- Propósito: Crear vistas base del dashboard (consolidación DRY fase 2)
-- Dependencias: Requiere v3_core_ssot (000113), v3_lot_ssot (000115),
--               v3_dashboard_ssot (000116), v3_workorder_metrics (000117),
--               v3_lot_metrics (000119), y v3_lot_list (000120)
-- Arquitectura: 2 vistas base (metrics + contributions)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN FASE 2:
-- - Vista 1: v3_dashboard_metrics (5 cards consolidadas)
-- - Vista 2: v3_dashboard_contributions_progress (1:N inversores)
-- - operational_indicators → migrada a 124
-- - management_balance → migrada a 122
-- - crop_incidence → migrada a 123
-- 
-- Nota: Vistas SOLO ensamblan, NO calculan (usan funciones SSOT)

BEGIN;

-- ========================================
-- ELIMINAR VISTAS ANTIGUAS
-- ========================================
DROP VIEW IF EXISTS v3_dashboard CASCADE;
DROP VIEW IF EXISTS v3_dashboard_sowing_metrics CASCADE;
DROP VIEW IF EXISTS v3_dashboard_harvest_metrics CASCADE;
DROP VIEW IF EXISTS v3_dashboard_costs_metrics CASCADE;
DROP VIEW IF EXISTS v3_dashboard_operating_result CASCADE;
DROP VIEW IF EXISTS v3_dashboard_contributions_progress CASCADE;
DROP VIEW IF EXISTS v3_dashboard_operational_indicators CASCADE;
DROP VIEW IF EXISTS v3_dashboard_management_balance CASCADE;
DROP VIEW IF EXISTS v3_dashboard_crop_incidence CASCADE;

-- ========================================
-- VISTA 1: v3_dashboard_metrics
-- ========================================
-- Propósito: Métricas consolidadas del dashboard (todas las 5 cards en una vista)
-- Cards incluidas:
--   1. Avance de siembra (sowing_*)
--   2. Avance de cosecha (harvest_*)
--   3. Avance de costos (costs_*)
--   4. Resultado operativo (operating_result_*)
--   5. Aportes se mantiene separada (1:N)
-- Funciones SSOT: v3_core_ssot.percentage, v3_dashboard_ssot.*, v3_lot_ssot.*

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
  
  -- Total costos = directos + arriendo + admin
  (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) + 
   COALESCE(p.admin_cost * ph.total_hectares, 0) + 
   COALESCE((SELECT f.lease_type_value * ph.total_hectares 
             FROM fields f 
             WHERE f.project_id = p.id AND f.deleted_at IS NULL 
             LIMIT 1), 0))::double precision AS operating_result_total_costs_usd,
  
  -- Porcentaje de rentabilidad
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
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

COMMENT ON VIEW public.v3_dashboard_metrics IS 'Métricas consolidadas del dashboard (5 cards en 1 vista) - Unifica criterio con lot_metrics/workorder_metrics';

-- ========================================
-- VISTA 2: v3_dashboard_contributions_progress
-- ========================================
-- Propósito: Avance de aportes por inversor (se mantiene separada por ser 1:N)
-- Campos: 4 (investor_id, investor_name, percentage, progress)

CREATE OR REPLACE VIEW public.v3_dashboard_contributions_progress AS
SELECT
  p.id AS project_id,
  pi.investor_id AS investor_id,
  i.name AS investor_name,
  pi.percentage AS investor_percentage_pct,
  pi.percentage::numeric AS contributions_progress_pct
FROM public.projects p
JOIN public.project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_contributions_progress IS 'Avance de aportes de inversores por proyecto (1:N)';

COMMIT;
