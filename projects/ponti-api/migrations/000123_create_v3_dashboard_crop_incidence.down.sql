-- ========================================
-- MIGRACIÓN 000123: CREATE v3_dashboard_crop_incidence VIEW (DOWN)
-- ========================================
-- 
-- Propósito: Revertir vista de incidencia por cultivo del dashboard
-- Fecha: 2025-10-04
-- Autor: Sistema

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_crop_incidence CASCADE;

COMMIT;
