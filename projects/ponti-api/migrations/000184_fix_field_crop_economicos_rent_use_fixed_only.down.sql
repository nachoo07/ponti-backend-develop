-- ========================================
-- MIGRACIÓN 000184: FIX Field Crop Economicos - Use Fixed Rent Only (DOWN)
-- ========================================
--
-- Propósito: Revertir a la versión anterior
-- Fecha: 2025-11-08
-- Autor: Sistema

BEGIN;

DROP VIEW IF EXISTS public.v3_report_field_crop_economicos CASCADE;

-- Al hacer migrate down, se ejecutará la versión UP de la migración anterior

COMMIT;

