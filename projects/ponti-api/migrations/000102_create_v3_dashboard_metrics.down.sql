-- ========================================
-- MIGRATION 000102: CREATE v3_dashboard VIEW (DOWN)
-- ========================================
-- 
-- Purpose: Eliminar vista de métricas del dashboard
-- Date: 2025-10-01
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

DROP VIEW IF EXISTS public.v3_dashboard CASCADE;

