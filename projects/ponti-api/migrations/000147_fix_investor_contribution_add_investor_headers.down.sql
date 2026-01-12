-- ========================================
-- MIGRACIÓN 000147: FIX v3_investor_contribution_data_view - Add investor_headers (DOWN)
-- ========================================

BEGIN;

-- Revertir a la versión sin investor_headers (migración 000146)
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- Nota: Esta es la versión de la migración 000146 sin investor_headers
-- En caso de rollback, se recomienda volver a aplicar la migración 000146 completa

COMMIT;

