-- ========================================
-- MIGRATION 000099: DROP v3_dashboard VIEWS (UP)
-- ========================================
-- 
-- Purpose: Eliminar todas las vistas v3 del dashboard para reorganización
-- Date: 2025-10-01
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Drop todas las vistas v3 del dashboard en orden inverso de dependencias
DROP VIEW IF EXISTS public.v3_dashboard_crop_incidence CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress CASCADE;
DROP VIEW IF EXISTS public.v3_dashboard CASCADE;

