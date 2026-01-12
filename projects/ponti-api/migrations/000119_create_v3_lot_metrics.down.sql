-- ========================================
-- MIGRACIÓN 000119: CREATE v3_lot_metrics (DOWN)
-- ========================================
-- 
-- Propósito: Rollback de v3_lot_metrics
-- Acción: Elimina la vista
-- Fecha: 2025-10-04
-- Autor: Sistema

BEGIN;

-- Eliminar vista
DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

COMMIT;

