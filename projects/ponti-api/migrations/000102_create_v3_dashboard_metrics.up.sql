-- ========================================
-- MIGRATION 000102: CREATE v3_dashboard VIEW (UP)
-- ========================================
-- 
-- Purpose: Vista de métricas del dashboard agregadas por proyecto desde lotes
-- Date: 2025-10-01
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- VISTA v3_dashboard (METRICS)
-- ========================================
-- Propósito: Agrega métricas de lotes a nivel proyecto para las 5 cards del dashboard
-- Basada en: Patrón de v3_lot_metrics (000091)
-- Datos desde: Lotes agregados por proyecto
CREATE OR REPLACE VIEW public.v3_dashboard AS
WITH lot_metrics_base AS (
  -- Base desde v3_lot_metrics: usa métricas ya calculadas (sowed_area_ha, harvested_area_ha)
  SELECT
    project_id,
    lot_id,
    hectares,
    sowed_area_ha,
    harvested_area_ha,
    direct_cost_usd,
    direct_cost_per_ha_usd
  FROM public.v3_lot_metrics
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
),
-- CTE para fechas operativas: primera orden de trabajo
w_min AS (
  SELECT 
    project_id, 
    MIN(date) AS start_date,
    -- Incluir número de la primera orden
    (SELECT id FROM public.workorders w2 
     WHERE w2.project_id = w.project_id 
       AND w2.date = MIN(w.date) 
       AND w2.deleted_at IS NULL 
     LIMIT 1) AS first_workorder_id
  FROM public.workorders w
  WHERE deleted_at IS NULL
  GROUP BY project_id
),
-- CTE para fechas operativas: última orden de trabajo
w_max AS (
  SELECT 
    project_id, 
    MAX(date) AS end_date,
    -- Incluir número de la última orden
    (SELECT id FROM public.workorders w2 
     WHERE w2.project_id = w.project_id 
       AND w2.date = MAX(w.date) 
       AND w2.deleted_at IS NULL 
     LIMIT 1) AS last_workorder_id
  FROM public.workorders w
  WHERE deleted_at IS NULL
  GROUP BY project_id
),
-- CTE para fecha de último arqueo de stock
last_stock_count AS (
  SELECT 
    project_id,
    MAX(close_date) AS last_stock_count_date
  FROM public.stocks
  WHERE deleted_at IS NULL
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- ========================================
  -- CARD 1: SOWING METRICS (Avance de Siembra)
  -- ========================================
  -- Superficie sembrada: desde v3_lot_metrics.sowed_area_ha (área efectiva desde workorders)
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  -- Superficie total: suma de todas las hectáreas del proyecto
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  -- Porcentaje de progreso de siembra
  v3_calc.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- ========================================
  -- CARD 2: HARVEST METRICS (Avance de Cosecha)
  -- ========================================
  -- Superficie cosechada: desde v3_lot_metrics.harvested_area_ha (área efectiva desde workorders)
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  -- Superficie total: misma que para siembra
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  -- Porcentaje de progreso de cosecha
  v3_calc.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- ========================================
  -- CARD 3: COSTS METRICS (Avance de Costos)
  -- ========================================
  -- Costo por hectárea promedio: promedio ponderado por superficie sembrada
  COALESCE(SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 0)::double precision AS executed_costs_usd,
  -- Presupuesto total del proyecto (dinámico)
  v3_calc.total_budget_cost_for_project(p.id) AS budget_cost_usd,
  -- Porcentaje de progreso de costos (costo/ha vs presupuesto total)
  v3_calc.percentage(
    COALESCE(SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 0)::numeric,
    v3_calc.total_budget_cost_for_project(p.id)
  ) AS costs_progress_pct,
  
  -- ========================================
  -- CARD 4: OPERATING RESULT METRICS (Resultado Operativo)
  -- ========================================
  -- Ingresos: suma de ingresos netos de todos los lotes desde función SSOT
  COALESCE(SUM(v3_calc.income_net_total_for_lot(lm.lot_id)), 0) AS income_usd,
  -- Resultado operativo: usando función SSOT corregida
  v3_calc.operating_result_total_for_project(p.id) AS operating_result_usd,
  -- Total activos: suma de costos directos + arriendo + admin por proyecto
  (COALESCE(v3_calc.direct_costs_total_for_project(p.id), 0) + 
   COALESCE(p.admin_cost * ph.total_hectares, 0) + 
   COALESCE((SELECT f.lease_type_value * ph.total_hectares 
             FROM fields f 
             WHERE f.project_id = p.id AND f.deleted_at IS NULL 
             LIMIT 1), 0))::double precision AS operating_result_total_costs_usd,
  -- Porcentaje de margen operativo (rentabilidad)
  v3_calc.renta_pct(
    v3_calc.operating_result_total_for_project(p.id),
    (COALESCE(v3_calc.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct,

  -- ========================================
  -- FECHAS OPERATIVAS (Operational Indicators)
  -- ========================================
  -- Fechas de inicio y fin de operaciones
  w_min.start_date,
  w_max.end_date,
  v3_calc.calculate_campaign_closing_date(w_max.end_date) AS campaign_closing_date,
  -- IDs de órdenes de trabajo (primera y última)
  w_min.first_workorder_id,
  w_max.last_workorder_id,
  -- Fecha del último arqueo de stock
  lsc.last_stock_count_date

  -- NOTA: La card 5 (Investor Contributions) NO se incluye aquí
  -- porque viene de la tabla project_investors, no de lotes.
  -- Existe en la vista separada: v3_dashboard_contributions_progress

FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
LEFT JOIN w_min ON w_min.project_id = p.id
LEFT JOIN w_max ON w_max.project_id = p.id
LEFT JOIN last_stock_count lsc ON lsc.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY
  p.customer_id, 
  p.id, 
  p.campaign_id,
  p.admin_cost,
  ph.total_hectares,
  w_min.start_date, 
  w_max.end_date, 
  w_min.first_workorder_id, 
  w_max.last_workorder_id,
  lsc.last_stock_count_date;