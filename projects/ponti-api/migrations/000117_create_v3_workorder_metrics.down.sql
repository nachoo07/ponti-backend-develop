-- ========================================
-- MIGRACIÓN 000117: CREATE v3_workorder_metrics (DOWN)
-- ========================================

BEGIN;

DROP VIEW IF EXISTS public.v3_workorder_metrics CASCADE;

COMMIT;
