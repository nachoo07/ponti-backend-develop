-- ========================================
-- MIGRATION 000157: RECREATE v3_dashboard_metrics (DOWN)
-- ========================================
--
-- Purpose: Eliminar v3_dashboard_metrics
-- Date: 2025-10-21
-- Author: System
--
-- Note: Code in English, comments in Spanish.

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_metrics;

COMMIT;

