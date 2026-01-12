-- ========================================
-- MIGRATION 000135: CREATE V3 INVESTOR CONTRIBUTION VIEWS (DOWN/ROLLBACK)
-- ========================================
-- 
-- Purpose: Revertir vistas del informe de Aportes por Inversor
-- Date: 2025-10-11
-- Author: System

BEGIN;

-- Eliminar vistas en orden inverso de creación
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view;
DROP VIEW IF EXISTS public.v3_report_investor_distributions;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories;
DROP VIEW IF EXISTS public.v3_report_investor_project_base;

COMMIT;

