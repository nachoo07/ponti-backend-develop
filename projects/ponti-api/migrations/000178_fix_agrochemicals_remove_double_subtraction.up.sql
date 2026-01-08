-- ========================================
-- MIGRACIÓN 000178: FIX Agrochemicals Remove Double Subtraction (UP)
-- ========================================
--
-- Propósito: Corregir doble contabilización en agroquímicos invertidos
-- Problema: v3_dashboard_ssot.agrochemicals_invested_for_project_mb() resta movimientos internos
--           con is_entry=FALSE, causando doble contabilización negativa
--           Dashboard muestra: 33.092 (INCORRECTO)
--           Debería mostrar: 43.377 (CORRECTO según movimientos)
-- Causa: Migración 000170 introdujo resta de movimientos internos is_entry=FALSE
--        Los movimientos internos vienen en pares espejo:
--        - is_entry=TRUE, qty=-560 (salida)
--        - is_entry=FALSE, qty=+560 (devolución)
--        La función actual suma el primero (-560) y resta el segundo (+560)
--        causando: -560 - 560 = -1120 (doble contabilización)
-- Solución: Revertir a versión 000153 que NO restaba movimientos internos
-- Fecha: 2025-11-03
-- Autor: Sistema
--
-- Impacto: Dashboard agroquímicos_invertidos_usd: 33.092 → 43.377
--          Coincidirá con Excel/exportaciones de movimientos de insumos
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- FIX: agrochemicals_invested_for_project_mb
-- ========================================
-- REVERTIR a versión 000153 (SIN resta de movimientos internos)

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2  -- Agroquímicos
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb IS 
'Calcula agroquímicos invertidos para un proyecto. FIX 000178: Elimina doble contabilización de movimientos internos.';

-- ========================================
-- FIX: seeds_invested_for_project_mb
-- ========================================
-- Aplicar mismo fix para semillas (mismo problema)

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name = 'Semilla'
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb IS 
'Calcula semillas invertidas para un proyecto. FIX 000178: Elimina doble contabilización de movimientos internos.';

-- ========================================
-- FIX: supply_movements_invested_total_for_project
-- ========================================
-- Aplicar mismo fix para total de insumos

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project IS 
'Calcula total invertido en movimientos de insumos. FIX 000178: Elimina doble contabilización de movimientos internos.';

COMMIT;

