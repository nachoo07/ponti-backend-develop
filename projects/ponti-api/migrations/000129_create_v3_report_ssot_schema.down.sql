-- ========================================
-- MIGRACIÓN 000129: DROP v3_report_ssot SCHEMA (DOWN)
-- ========================================

BEGIN;

-- Eliminar funciones
DROP FUNCTION IF EXISTS v3_report_ssot.board_price_for_lot(bigint);
DROP FUNCTION IF EXISTS v3_report_ssot.freight_cost_for_lot(bigint);
DROP FUNCTION IF EXISTS v3_report_ssot.commercial_cost_for_lot(bigint);

-- Eliminar esquema
DROP SCHEMA IF EXISTS v3_report_ssot;

COMMIT;
