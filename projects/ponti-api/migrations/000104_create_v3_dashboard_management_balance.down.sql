-- ========================================
-- MIGRACIÓN 000105: CREATE v3_dashboard_management_balance VIEW (DOWN)
-- ========================================
-- 
-- Propósito: Eliminar vista de balance de gestión
-- Fecha: 2025-01-01
-- Autor: Sistema

DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;
