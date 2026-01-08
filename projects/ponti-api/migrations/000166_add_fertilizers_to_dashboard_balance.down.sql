-- ========================================
-- MIGRACIÓN 000166: ADD Fertilizantes to Dashboard Management Balance (DOWN)
-- ========================================
-- 
-- Propósito: Revertir cambios de la migración 000166
-- Fecha: 2025-10-29
-- Autor: Sistema

BEGIN;

-- Revertir a la versión anterior (000122)
-- Ejecutar: migrate down 1 && migrate up

DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

COMMIT;

