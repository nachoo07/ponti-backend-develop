-- ========================================
-- MIGRACIÓN 000120: CREATE v3_lot_list (DOWN)
-- ========================================
-- 
-- Propósito: Rollback de v3_lot_list
-- Acción: Elimina la vista
-- Nota: Esta vista usa v3_dashboard_ssot.cost_per_ha_for_crop_ssot
-- Fecha: 2025-10-04
-- Autor: Sistema

BEGIN;

-- Eliminar vista
DROP VIEW IF EXISTS public.v3_lot_list CASCADE;

COMMIT;

