-- =============================================================================
-- MIGRACIÓN 000302: v4_ssot - Wrappers thin que delegan a v3_lot_ssot
-- =============================================================================
--
-- Propósito: FASE 1 - Wrappers que delegan a v3 (NO reimplementar todavía)
-- Fecha: 2025-01-XX
-- Autor: Sistema
--

CREATE OR REPLACE FUNCTION v4_ssot.seeded_area_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.seeded_area_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.harvested_area_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.harvested_area_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.labor_cost_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.labor_cost_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.supply_cost_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.supply_cost_for_lot_base(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.yield_tn_per_ha_for_lot(p_lot_id bigint)
RETURNS double precision LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.yield_tn_per_ha_for_lot(p_lot_id), 0);  -- double precision (paridad con v3)
$$;

CREATE OR REPLACE FUNCTION v4_ssot.income_net_total_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.income_net_total_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.rent_fixed_only_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.rent_fixed_only_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.rent_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.rent_per_ha_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id), 0)::numeric;
$$;

COMMENT ON SCHEMA v4_ssot IS 'FASE 1: Wrappers thin a v3_lot_ssot. FASE 2: Reimplementar.';
