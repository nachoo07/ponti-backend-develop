-- ========================================
-- MIGRATION 000077: DROP calc SCHEMA AND WRAPPERS (DOWN)
-- ========================================
-- 
-- Purpose: Revert creation of calc schema and public wrappers
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- No hay wrappers públicos que eliminar

-- Eliminar funciones en schema calc
DROP FUNCTION IF EXISTS v3_calc.norm_dose(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.units_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.indifference_price_usd_tn(double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.renta_pct(double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.operating_result_per_ha(double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.active_total_per_ha(double precision, double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.rent_per_ha(integer, double precision, double precision, double precision, double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.income_net_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.income_net_total(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.cost_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.supply_cost(double precision, numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.labor_cost(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.yield_tn_per_ha_over_harvested(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.yield_tn_per_ha_over_hectares(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.harvested_area(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.seeded_area(date, numeric);
DROP FUNCTION IF EXISTS v3_calc.dose_per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.per_ha_dp(double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.per_ha(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.percentage_capped(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.percentage(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.safe_div_dp(double precision, double precision);
DROP FUNCTION IF EXISTS v3_calc.safe_div(numeric, numeric);
DROP FUNCTION IF EXISTS v3_calc.coalesce0(double precision);
DROP FUNCTION IF EXISTS v3_calc.coalesce0(numeric);
DROP FUNCTION IF EXISTS v3_calc.calculate_campaign_closing_date(date);

-- Eliminar schema v3_calc
DROP SCHEMA IF EXISTS v3_calc;

COMMIT;


