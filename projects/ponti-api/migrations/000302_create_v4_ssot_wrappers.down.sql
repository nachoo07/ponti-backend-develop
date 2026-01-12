-- =============================================================================
-- MIGRACIÓN 000302: v4_ssot wrappers (DOWN)
-- =============================================================================

DROP FUNCTION IF EXISTS v4_ssot.admin_cost_per_ha_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.rent_per_ha_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.rent_fixed_only_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.income_net_total_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.yield_tn_per_ha_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.supply_cost_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.labor_cost_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.harvested_area_for_lot(bigint);
DROP FUNCTION IF EXISTS v4_ssot.seeded_area_for_lot(bigint);
