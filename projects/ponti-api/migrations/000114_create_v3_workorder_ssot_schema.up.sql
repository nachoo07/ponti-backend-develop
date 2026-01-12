-- ========================================
-- MIGRACIÓN 000114: CREATE v3_workorder_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Funciones SSOT para workorders individuales (REDUCIDO por consolidación DRY)
-- Dependencias: Requiere v3_core_ssot (000113)
-- Alcance: 2 funciones (solo cálculos por workorder individual)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY:
-- - Eliminadas: labor_cost_for_lot_wo, supply_cost_for_lot_wo (movidas a v3_lot_ssot)
-- - Eliminadas: surface_for_lot, liters_for_lot, kilograms_for_lot (movidas a v3_lot_ssot)
-- - Mantiene: Solo funciones por workorder individual
-- 
-- Nota: Código en inglés, comentarios en español
-- Usa v3_core_ssot para operaciones básicas

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_workorder_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_workorder_ssot;

COMMENT ON SCHEMA v3_workorder_ssot IS 'Funciones SSOT de workorders: cálculos por workorder, lote y proyecto';

-- ========================================
-- GRUPO 1: COSTOS POR WORKORDER INDIVIDUAL (2 funciones) - ÚNICO SCOPE
-- ========================================
-- Propósito: Calcular costos básicos por workorder individual
-- Nota: Funciones *_for_lot movidas a v3_lot_ssot (consolidación DRY)

CREATE OR REPLACE FUNCTION v3_workorder_ssot.labor_cost_for_workorder(p_workorder_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.labor_cost(lb.price::numeric, w.effective_area::numeric)
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.id = p_workorder_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

CREATE OR REPLACE FUNCTION v3_workorder_ssot.supply_cost_for_workorder(p_workorder_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.id = p_workorder_id
    AND w.deleted_at IS NULL
$$;

COMMIT;

-- Comentarios sobre funciones
COMMENT ON FUNCTION v3_workorder_ssot.labor_cost_for_workorder(bigint) IS 'Calcula costo de labor por workorder individual';
COMMENT ON FUNCTION v3_workorder_ssot.supply_cost_for_workorder(bigint) IS 'Calcula costo de insumos por workorder individual';
