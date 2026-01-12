-- ========================================
-- MIGRACIÓN 000118: CREATE v3_workorder_list (DOWN)
-- ========================================

BEGIN;

DROP VIEW IF EXISTS public.v3_workorder_list CASCADE;

COMMIT;
