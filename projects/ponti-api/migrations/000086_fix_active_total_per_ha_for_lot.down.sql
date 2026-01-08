-- ========================================
-- MIGRATION 000086: FIX active_total_per_ha_for_lot (DOWN)
-- ========================================
-- 
-- Purpose: Revert active_total_per_ha_for_lot to original implementation
-- Date: 2025-01-27
-- Author: System

-- ========================================
-- REVERTIR active_total_per_ha_for_lot
-- ========================================
-- Volver a la implementación original con prorrateo

CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.active_total_per_ha(
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.rent_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;
