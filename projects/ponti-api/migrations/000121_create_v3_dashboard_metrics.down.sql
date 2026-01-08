-- ========================================
-- MIGRACIÓN 000121: CREATE DASHBOARD SEPARATED VIEWS (DOWN)
-- ========================================
-- 
-- Propósito: Eliminar vistas separadas del dashboard
-- Fecha: 2025-10-04
-- Autor: Sistema

-- Eliminar las 5 cards
DROP VIEW IF EXISTS public.v3_dashboard_sowing_metrics CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_harvest_metrics CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_costs_metrics CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_operating_result CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress CASCADE;

-- Eliminar los 3 módulos adicionales
DROP VIEW IF EXISTS public.v3_dashboard_operational_indicators CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_crop_incidence CASCADE;

