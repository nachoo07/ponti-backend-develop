-- ========================================
-- MIGRACIÓN 000167: SEPARATE FERTILIZERS IN INVESTOR REPORT (DOWN)
-- ========================================
-- 
-- Propósito: Revertir cambios de la migración 000167
-- Fecha: 2025-10-29
-- Autor: Sistema

BEGIN;

-- Revertir a la versión anterior (ejecutar migrate down 1 && migrate up para volver a la 163)
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;

COMMIT;

