-- =============================================================================
-- MIGRACIÓN 000301: v4_core functions (DOWN)
-- =============================================================================

DROP FUNCTION IF EXISTS v4_core.percentage(numeric, numeric);
DROP FUNCTION IF EXISTS v4_core.per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v4_core.safe_div(numeric, numeric);
