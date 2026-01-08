-- ========================================
-- MIGRACIÓN 000187: FIX Dashboard Contributions Use Real Data (DOWN)
-- ========================================
--
-- Propósito: Revertir a la versión anterior
-- Fecha: 2025-11-08
-- Autor: Sistema

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress CASCADE;

-- Al hacer migrate down, se ejecutará la versión UP de la migración anterior

COMMIT;

