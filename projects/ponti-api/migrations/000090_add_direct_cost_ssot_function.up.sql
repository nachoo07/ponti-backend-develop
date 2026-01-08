-- ========================================
-- MIGRATION 000090: ADD DIRECT COST SSOT FUNCTION (UP)
-- ========================================
--
-- Purpose: Add new SSOT function for direct_cost_usd calculation
--          direct_cost_usd = labor_cost_usd + supplies_cost_usd
-- Date: 2025-01-27
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- =============================================================================
-- NUEVA FUNCIÓN SSOT: direct_cost_usd
-- =============================================================================
-- Purpose: Single Source of Truth for direct_cost_usd calculation
-- Formula: direct_cost_usd = labor_cost_usd + supplies_cost_usd

-- Función para calcular costo directo usando labor_cost_usd + supplies_cost_usd
CREATE OR REPLACE FUNCTION v3_calc.direct_cost_usd(
  p_labor_cost_usd numeric,
  p_supplies_cost_usd numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(p_labor_cost_usd, 0)::numeric + COALESCE(p_supplies_cost_usd, 0)::numeric
$$;

-- Función para calcular costo directo por hectárea usando la nueva función SSOT
CREATE OR REPLACE FUNCTION v3_calc.direct_cost_per_ha_usd(
  p_labor_cost_usd numeric,
  p_supplies_cost_usd numeric,
  p_surface_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.cost_per_ha(
    v3_calc.direct_cost_usd(p_labor_cost_usd, p_supplies_cost_usd),
    p_surface_ha
  )
$$;
