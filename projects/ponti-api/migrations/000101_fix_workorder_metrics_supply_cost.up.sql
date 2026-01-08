-- ========================================
-- MIGRACIÓN 000101: FIX WORKORDER METRICS SUPPLY COST (UP)
-- ========================================
-- 
-- Propósito: Corregir SOLO la métrica de costos directos en v3_workorder_metrics
-- Problema: La función supply_cost_for_lot incluye movimientos internos que no deberían impactar workorders
-- Solución: Crear función específica para workorders que excluya movimientos internos
-- Fecha: 2025-01-01
-- Autor: Sistema
-- 
-- Nota: Solución quirúrgica - solo afecta v3_workorder_metrics, no otras funcionalidades

-- -------------------------------------------------------------------
-- 1. CREAR FUNCIÓN ESPECÍFICA PARA WORKORDERS (SIN MOVIMIENTOS INTERNOS)
-- -------------------------------------------------------------------
-- Esta función calcula costos de insumos SOLO para workorders, excluyendo movimientos internos
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_workorder(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders) - SOLO ESTO
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- 2. ACTUALIZAR VISTA v3_workorder_metrics PARA USAR LA NUEVA FUNCIÓN
-- -------------------------------------------------------------------
-- Recrear la vista usando la función específica para workorders
DROP VIEW IF EXISTS public.v3_workorder_metrics;

CREATE OR REPLACE VIEW public.v3_workorder_metrics AS
WITH base AS (
  SELECT
    w.id                         AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price                     AS labor_price
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
-- Superficie correcta: suma única por workorder, sin duplicar por insumos
surface AS (
  SELECT project_id, field_id, lot_id, SUM(effective_area)::numeric AS surface_ha
  FROM base
  GROUP BY project_id, field_id, lot_id
),
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM(v3_calc.labor_cost(labor_price, effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    -- USAR LA NUEVA FUNCIÓN ESPECÍFICA PARA WORKORDERS
    v3_calc.supply_cost_for_workorder(b.lot_id)::numeric AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
         ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id,  lc.field_id,  sm.field_id)     AS field_id,
  COALESCE(sur.lot_id,    lc.lot_id,    sm.lot_id)       AS lot_id,
  COALESCE(sur.surface_ha, 0)::numeric                    AS surface_ha,
  COALESCE(sm.liters, 0)::numeric                         AS liters,
  COALESCE(sm.kilograms, 0)::numeric                      AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric                 AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric              AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric)            AS direct_cost_usd,
  v3_calc.cost_per_ha(
    COALESCE(lc.labor_cost_usd,0)::numeric + COALESCE(sm.supplies_cost_usd,0)::numeric,
    COALESCE(sur.surface_ha,0)::numeric
  )                                                        AS avg_cost_per_ha_usd,
  v3_calc.per_ha(COALESCE(sm.liters,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)     AS liters_per_ha,
  v3_calc.per_ha(COALESCE(sm.kilograms,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)  AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs   lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id);
