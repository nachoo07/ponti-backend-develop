-- ========================================
-- MIGRACIÓN 000122: CREATE v3_dashboard_management_balance VIEW (DOWN)
-- ========================================
-- 
-- Propósito: Revertir vista de balance de gestión del dashboard
-- Fecha: 2025-10-04
-- Autor: Sistema

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

COMMIT;
