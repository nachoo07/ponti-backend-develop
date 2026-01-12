-- ========================================
-- MIGRATION 000081: CREATE v3_dashboard VIEW (UP)
-- ========================================
-- 
-- Purpose: Dashboard v3 view based on cost and income base views
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

CREATE OR REPLACE VIEW public.v3_dashboard AS
WITH lots_base AS (
  SELECT
    l.id          AS lot_id,
    f.project_id  AS project_id,
    l.hectares    AS hectares,
    l.tons        AS tons,
    l.sowing_date AS sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
w_min AS (
  SELECT 
    project_id, 
    MIN(date) AS start_date,
    -- FIX: Incluir número de la primera orden
    (SELECT id FROM public.workorders w2 
     WHERE w2.project_id = w.project_id 
       AND w2.date = MIN(w.date) 
       AND w2.deleted_at IS NULL 
     LIMIT 1) AS first_workorder_id
  FROM public.workorders w
  WHERE deleted_at IS NULL
  GROUP BY project_id
),
w_max AS (
  SELECT 
    project_id, 
    MAX(date) AS end_date,
    -- FIX: Incluir número de la última orden
    (SELECT id FROM public.workorders w2 
     WHERE w2.project_id = w.project_id 
       AND w2.date = MAX(w.date) 
       AND w2.deleted_at IS NULL 
     LIMIT 1) AS last_workorder_id
  FROM public.workorders w
  WHERE deleted_at IS NULL
  GROUP BY project_id
),
-- FIX: Agregar CTE para fecha de último arqueo de stock
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
  p.id            AS project_id,
  p.campaign_id,
  -- Progresos operativos (si no hay lotes, da 0)
  COALESCE(SUM(CASE WHEN lb.sowing_date IS NOT NULL
                    THEN lb.hectares ELSE 0 END), 0) AS sowing_hectares,
  COALESCE(SUM(lb.hectares), 0)                    AS sowing_total_hectares,
  v3_calc.percentage(
    COALESCE(SUM(CASE WHEN lb.sowing_date IS NOT NULL THEN lb.hectares ELSE 0 END), 0)::numeric,
    COALESCE(SUM(lb.hectares), 0)::numeric
  )                                                 AS sowing_progress_pct,

  COALESCE(SUM(CASE WHEN lb.tons IS NOT NULL AND lb.tons > 0
                    THEN lb.hectares ELSE 0 END), 0) AS harvest_hectares,
  COALESCE(SUM(lb.hectares), 0)                       AS harvest_total_hectares,
  v3_calc.percentage(
    COALESCE(SUM(CASE WHEN lb.tons IS NOT NULL AND lb.tons > 0 THEN lb.hectares ELSE 0 END), 0)::numeric,
    COALESCE(SUM(lb.hectares), 0)::numeric
  )                                                    AS harvest_progress_pct,

  -- Costos e ingresos agregados calculados vía funciones SSOT (v3_calc.*)
  COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0) AS executed_costs_usd,
  v3_calc.total_budget_cost_for_project(p.id)              AS budget_cost_usd,  -- FIX: Usar presupuesto real
  v3_calc.percentage(
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric,
    v3_calc.total_budget_cost_for_project(p.id)::numeric
  )                                                       AS costs_progress_pct,

  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0) AS income_usd,
  -- FIX: Usar función corregida para resultado operativo
  v3_calc.operating_result_total_for_project(p.id)        AS operating_result_usd,
  -- FIX: Usar total invertido en lugar de solo costos directos
  v3_calc.total_invested_cost_for_project(p.id)           AS operating_result_total_costs_usd,
  -- FIX: Usar solo costos directos para cálculo de porcentaje (según especificación)
  v3_calc.renta_pct(
    v3_calc.operating_result_total_for_project(p.id),
    v3_calc.total_costs_for_project(p.id)
  )                                                       AS operating_result_pct,

  -- FIX: Fechas operativas con números de órdenes
  w_min.start_date,
  w_max.end_date,
  v3_calc.calculate_campaign_closing_date(w_max.end_date) AS campaign_closing_date,
  w_min.first_workorder_id,  -- FIX: Agregar número de primera orden
  w_max.last_workorder_id,   -- FIX: Agregar número de última orden
  lsc.last_stock_count_date  -- FIX: Agregar fecha de último arqueo

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN w_min ON w_min.project_id = p.id
LEFT JOIN w_max ON w_max.project_id = p.id
LEFT JOIN last_stock_count lsc ON lsc.project_id = p.id  -- FIX: Agregar join para stock
WHERE p.deleted_at IS NULL
GROUP BY
  p.customer_id, p.id, p.campaign_id, p.admin_cost,
  w_min.start_date, w_max.end_date, w_min.first_workorder_id, w_max.last_workorder_id,
  lsc.last_stock_count_date;




-- ========================================
-- VISTA v3_dashboard_contributions_progress
-- ========================================
-- Nota: Cálculos basados en datos reales y funciones SSOT donde aplica
CREATE OR REPLACE VIEW public.v3_dashboard_contributions_progress AS
SELECT
  p.id                       AS project_id,
  pi.investor_id             AS investor_id,
  i.name                     AS investor_name,
  pi.percentage              AS investor_percentage_pct,
  -- FIX: Mostrar directamente el porcentaje del inversor individual
  pi.percentage::numeric     AS contributions_progress_pct  -- FIX: Usar porcentaje individual
FROM public.projects p
JOIN public.project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN public.investors i          ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL;


-- ========================================
-- VISTA v3_dashboard_management_balance
-- ========================================
-- Nota: Usa funciones SSOT (v3_calc.*) para costos/ingresos y componentes
CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
WITH lots_base AS (
  SELECT
    l.id         AS lot_id,
    f.project_id AS project_id,
    l.hectares   AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
)
SELECT
  p.id AS project_id,

  -- Ingresos netos
  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)             AS income_usd,

  -- Costos directos ejecutados (labores + insumos)
  COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)                  AS costos_directos_ejecutados_usd,

  -- FIX: Costos directos invertidos (solo labores + insumos)
  v3_calc.direct_costs_invested_for_project(p.id)                           AS costos_directos_invertidos_usd,

  -- Componentes de invertidos (mantener separados)
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)     AS arriendo_invertidos_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS estructura_invertidos_usd,

  -- Resultado operativo y ratio
  COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS operating_result_usd,
  -- FIX: Usar solo costos directos para cálculo de porcentaje (según especificación)
  v3_calc.renta_pct(
    COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision,
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
  )                                                                          AS operating_result_pct
FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.id;


-- ========================================
-- VISTA v3_dashboard_crop_incidence
-- ========================================
-- Nota: Incidencia de hectáreas por cultivo dentro del proyecto
CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS
WITH lot_base AS (
  SELECT
    l.id               AS lot_id,
    f.project_id       AS project_id,
    l.current_crop_id  AS current_crop_id,
    c.name             AS crop_name,
    l.hectares         AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares IS NOT NULL AND l.hectares > 0
),
by_crop AS (
  SELECT 
    project_id, 
    current_crop_id, 
    crop_name, 
    SUM(hectares)::numeric AS crop_hectares,
    -- FIX: Calcular costos por cultivo
    v3_calc.total_costs_for_crop(project_id, current_crop_id) AS crop_costs_usd
  FROM lot_base
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
total_by_project AS (
  SELECT 
    project_id, 
    SUM(crop_costs_usd)::numeric AS total_costs_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.crop_hectares,
  -- FIX: Calcular incidencia por costos en lugar de hectáreas
  v3_calc.percentage(bc.crop_costs_usd::numeric, t.total_costs_usd) AS crop_incidence_pct,
  -- IMPLEMENTAR: Costo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop(bc.project_id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN total_by_project t ON t.project_id = bc.project_id;
