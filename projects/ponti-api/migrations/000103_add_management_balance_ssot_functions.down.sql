-- ========================================
-- MIGRACIÓN 000104: ADD MANAGEMENT BALANCE SSOT FUNCTIONS (DOWN)
-- ========================================
-- 
-- Propósito: Revertir las funciones SSOT específicas para management balance
-- Fecha: 2025-01-01
-- Autor: Sistema

-- -------------------------------------------------------------------
-- ELIMINAR FUNCIONES SSOT AGREGADAS
-- -------------------------------------------------------------------
DROP FUNCTION IF EXISTS v3_calc.supply_cost_seeds_for_lot_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.supply_cost_agrochemicals_for_lot_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.labor_cost_for_lot_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.direct_costs_invested_for_project_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.stock_value_for_project_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.agrochemicals_invested_for_project_mb(bigint);
DROP FUNCTION IF EXISTS v3_calc.seeds_invested_for_project_mb(bigint);
