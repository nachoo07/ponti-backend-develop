-- ========================================
-- MIGRACIÓN 000115: CREATE v3_lot_ssot SCHEMA (DOWN)
-- ========================================
-- 
-- Propósito: Rollback de v3_lot_ssot
-- Acción: Elimina esquema y todas sus funciones
-- Fecha: 2025-10-04
-- Autor: Sistema

BEGIN;

-- Eliminar esquema y todas sus funciones
DROP SCHEMA IF EXISTS v3_lot_ssot CASCADE;

COMMIT;

