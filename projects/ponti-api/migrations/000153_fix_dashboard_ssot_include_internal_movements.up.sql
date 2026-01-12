-- ========================================
-- MIGRACIÓN 000153: FIX Dashboard SSOT - Incluir Movimientos Internos (UP)
-- ========================================
-- 
-- Propósito: Sincronizar funciones de insumos del Dashboard con Informe de Aportes
--            para incluir movimientos internos (is_entry = TRUE)
-- Problema: Control 7 falla con diferencia de $7.62 en Proyecto 4
--           Dashboard: $14,000.00 (sin movimientos internos)
--           Aportes:   $14,007.62 (con movimientos internos)
-- Solución: Actualizar seeds_invested_for_project_mb y agrochemicals_invested_for_project_mb
--           para incluir is_entry = TRUE y movimientos internos
-- Fecha: 2025-10-19
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- CORREGIR: seeds_invested_for_project_mb
-- ========================================
-- Sincronizar con lógica del Informe de Aportes (migración 155)

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
       AND sm.is_entry = TRUE  -- AÑADIDO: incluye movimientos internos con qty negativa
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb IS 
'Calcula semillas invertidas para un proyecto (incluye movimientos internos con is_entry=TRUE). Sincronizado con Informe de Aportes.';

-- ========================================
-- CORREGIR: agrochemicals_invested_for_project_mb
-- ========================================
-- Sincronizar con lógica del Informe de Aportes (migración 155)

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
       AND c.type_id = 2  -- Agroquímicos y Fertilizantes
       AND sm.is_entry = TRUE  -- AÑADIDO: incluye movimientos internos con qty negativa
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb IS 
'Calcula agroquímicos invertidos para un proyecto (incluye movimientos internos con is_entry=TRUE). Sincronizado con Informe de Aportes.';

-- ========================================
-- CORREGIR: supply_movements_invested_total_for_project
-- ========================================
-- Esta función se usa en v3_dashboard_management_balance.costos_directos_invertidos_usd
-- Necesita incluir movimientos internos para que el total invertido sea consistente

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
       AND sm.is_entry = TRUE  -- AÑADIDO: incluye movimientos internos con qty negativa
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project IS 
'Calcula total de movimientos de insumos invertidos para un proyecto (incluye movimientos internos con is_entry=TRUE). Usado en costos_directos_invertidos_usd.';

COMMIT;

