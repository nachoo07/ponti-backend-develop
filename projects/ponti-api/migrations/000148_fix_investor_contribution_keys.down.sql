-- ========================================
-- MIGRACIÓN 000148: FIX investor contribution keys to match frontend (DOWN)
-- ========================================

BEGIN;

-- Revertir a la versión anterior (migración 000147)
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- En caso de rollback, volver a aplicar la migración 000147

COMMIT;

