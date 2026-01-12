-- ========================================
-- MIGRATION 000079: DROP v3_lot_views (DOWN)
-- ========================================
-- 
-- Purpose: Drop the views created in the UP migration
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_lot_metrics: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_lot_metrics;

-- -------------------------------------------------------------------
-- v3_lot_list: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_lot_list;
