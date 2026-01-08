-- ========================================
-- MIGRACIÓN 000100: ADD SSOT FUNCTIONS FOR DASHBOARD (DOWN)
-- ========================================
-- 
-- Propósito: Revertir las funciones SSOT para dashboard (costos directos y resultado operativo)
-- Fecha: 2025-01-01
-- Autor: Sistema

-- -------------------------------------------------------------------
-- ELIMINAR FUNCIONES SSOT AGREGADAS
-- -------------------------------------------------------------------
DROP FUNCTION IF EXISTS v3_calc.direct_costs_total_for_project(bigint);
DROP FUNCTION IF EXISTS v3_calc.operating_result_total_for_project(bigint);
