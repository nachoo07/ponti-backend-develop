-- ========================================
-- MIGRACIÓN 000171: FIX General Project Data Structure (DOWN)
-- ========================================

BEGIN;

-- Revert to previous version (000168)
-- This will restore the broken structure, but allows rollback

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- Note: This down migration reverts to the 000168 version which has the broken structure
-- If you need to rollback, you should apply 000167 instead

COMMIT;

