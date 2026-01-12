-- ========================================
-- MIGRACIÓN 000183: FIX Field Crop Report - Use Fixed Rent Only (DOWN)
-- ========================================
--
-- Propósito: Revertir a la versión anterior (000180)
-- Fecha: 2025-11-08
-- Autor: Sistema

BEGIN;

-- Recrear la vista como estaba en 000180
-- (La versión anterior usaba rent_per_ha_for_lot())
DROP VIEW IF EXISTS public.v3_report_field_crop_rentabilidad CASCADE;

-- No recreamos aquí porque al hacer migrate down, automáticamente
-- se ejecutará la versión UP de la migración 000180

COMMIT;

