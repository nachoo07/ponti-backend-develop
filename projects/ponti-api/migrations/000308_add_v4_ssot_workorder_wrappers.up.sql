-- =============================================================================
-- MIGRACIÓN 000308: v4_ssot - Wrappers adicionales para workorder_metrics
-- =============================================================================
--
-- Propósito: Wrappers thin para funciones usadas por v3_workorder_metrics
-- Fuente: 000117_create_v3_workorder_metrics.up.sql
--

-- Superficie trabajada (suma de effective_area)
CREATE OR REPLACE FUNCTION v4_ssot.surface_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.surface_for_lot(p_lot_id), 0)::numeric;
$$;

-- Litros de insumos
CREATE OR REPLACE FUNCTION v4_ssot.liters_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.liters_for_lot(p_lot_id), 0)::numeric;
$$;

-- Kilogramos de insumos
CREATE OR REPLACE FUNCTION v4_ssot.kilograms_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.kilograms_for_lot(p_lot_id), 0)::numeric;
$$;

COMMENT ON FUNCTION v4_ssot.surface_for_lot IS 'Wrapper thin a v3_lot_ssot.surface_for_lot';
COMMENT ON FUNCTION v4_ssot.liters_for_lot IS 'Wrapper thin a v3_lot_ssot.liters_for_lot';
COMMENT ON FUNCTION v4_ssot.kilograms_for_lot IS 'Wrapper thin a v3_lot_ssot.kilograms_for_lot';
