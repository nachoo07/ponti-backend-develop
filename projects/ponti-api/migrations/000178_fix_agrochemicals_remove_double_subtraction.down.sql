-- ========================================
-- MIGRACIÓN 000178: FIX Agrochemicals Remove Double Subtraction (DOWN)
-- ========================================
--
-- Propósito: Revertir corrección de doble contabilización
-- Fecha: 2025-11-03
-- Autor: Sistema

BEGIN;

-- Restaurar versión 000170 (con resta de movimientos internos)

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Sumar entradas (is_entry=TRUE)
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
    -
    -- Restar salidas de movimientos internos (is_entry=FALSE)
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

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
    -
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name = 'Semilla'
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

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
    -
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

COMMIT;

