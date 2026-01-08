-- ========================================
-- MIGRACIÓN 000104: ADD MANAGEMENT BALANCE SSOT FUNCTIONS (UP)
-- ========================================
-- 
-- Propósito: Crear funciones SSOT específicas para management balance
-- Funciones: Nuevas funciones con sufijo _mb para no romper existentes
-- Fecha: 2025-01-01
-- Autor: Sistema
-- 
-- Nota: Funciones específicas para v3_dashboard_management_balance

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DE SEMILLAS POR LOTE (SIN MOVIMIENTOS INTERNOS)
-- -------------------------------------------------------------------
-- Esta función calcula costos de semillas SOLO desde workorder_items (unit_id = 1)
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_seeds_for_lot_mb(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND s.unit_id = 2  -- Solo semillas
       AND w.lot_id = p_lot_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DE AGROQUÍMICOS POR LOTE (SIN MOVIMIENTOS INTERNOS)
-- -------------------------------------------------------------------
-- Esta función calcula costos de agroquímicos SOLO desde workorder_items (unit_id = 2)
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_agrochemicals_for_lot_mb(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND s.unit_id = 1  -- Solo agroquímicos
       AND w.lot_id = p_lot_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DE LABORES POR LOTE (SIN MOVIMIENTOS INTERNOS)
-- -------------------------------------------------------------------
-- Esta función calcula costos de labores SOLO desde workorders
CREATE OR REPLACE FUNCTION v3_calc.labor_cost_for_lot_mb(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::double precision
     FROM public.workorders w
     JOIN public.labors lb ON lb.id = w.labor_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND lb.price IS NOT NULL
       AND w.lot_id = p_lot_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DIRECTOS INVERTIDOS POR PROYECTO (CORREGIDA)
-- -------------------------------------------------------------------
-- Esta función calcula costos invertidos de forma realista
-- Lógica: Insumos en stock inicial + Labores planificadas del proyecto correcto
CREATE OR REPLACE FUNCTION v3_calc.direct_costs_invested_for_project_mb(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Insumos invertidos (stock inicial)
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    -- Labores planificadas (solo las que se ejecutaron)
    (SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::double precision
     FROM public.workorders w
     JOIN public.labors lb ON lb.id = w.labor_id
     JOIN public.lots l ON l.id = w.lot_id
     JOIN public.fields f ON f.id = l.field_id
     WHERE f.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA VALOR DE STOCK POR PROYECTO
-- -------------------------------------------------------------------
-- Esta función calcula el valor del stock (invertidos - ejecutados)
CREATE OR REPLACE FUNCTION v3_calc.stock_value_for_project_mb(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    v3_calc.direct_costs_invested_for_project_mb(p_project_id) - 
    v3_calc.direct_costs_total_for_project(p_project_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA AGROQUÍMICOS INVERTIDOS (STOCK + REMITO)
-- -------------------------------------------------------------------
-- Esta función calcula agroquímicos invertidos: stock inicial + movimientos stock + remito oficial
CREATE OR REPLACE FUNCTION v3_calc.agrochemicals_invested_for_project_mb(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock inicial de agroquímicos
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND s.unit_id = 1 AND st.initial_units IS NOT NULL)
    +
    -- Movimientos de stock de agroquímicos
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Stock' 
       AND sm.deleted_at IS NULL AND s.unit_id = 1)
    +
    -- Movimientos de remito oficial de agroquímicos
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Remito oficial' 
       AND sm.deleted_at IS NULL AND s.unit_id = 1)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA SEMILLAS INVERTIDAS (STOCK + REMITO)
-- -------------------------------------------------------------------
-- Esta función calcula semillas invertidas: stock inicial + movimientos stock + remito oficial
CREATE OR REPLACE FUNCTION v3_calc.seeds_invested_for_project_mb(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock inicial de semillas
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND s.unit_id = 2 AND st.initial_units IS NOT NULL)
    +
    -- Movimientos de stock de semillas
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Stock' 
       AND sm.deleted_at IS NULL AND s.unit_id = 2)
    +
    -- Movimientos de remito oficial de semillas
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Remito oficial' 
       AND sm.deleted_at IS NULL AND s.unit_id = 2)
  , 0)::double precision
$$;
