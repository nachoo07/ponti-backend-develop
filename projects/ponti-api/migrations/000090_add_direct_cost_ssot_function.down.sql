-- ========================================
-- MIGRATION 000090: ADD DIRECT COST SSOT FUNCTION (DOWN)
-- ========================================
--
-- Purpose: Remove SSOT functions for direct_cost_usd calculation
-- Date: 2025-01-27
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- Eliminar las funciones SSOT agregadas
DROP FUNCTION IF EXISTS v3_calc.direct_cost_per_ha_usd(numeric, numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.direct_cost_usd(numeric, numeric);
