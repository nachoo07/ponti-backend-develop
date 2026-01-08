-- ========================================
-- MIGRATION 000145: RECREATE V3_DASHBOARD_METRICS (DOWN/ROLLBACK)
-- ========================================
-- 
-- Purpose: Revertir recreación de v3_dashboard_metrics
-- Date: 2025-10-14
-- Author: System

BEGIN;

-- Eliminar vista recreada
DROP VIEW IF EXISTS public.v3_dashboard_metrics CASCADE;

COMMIT;

