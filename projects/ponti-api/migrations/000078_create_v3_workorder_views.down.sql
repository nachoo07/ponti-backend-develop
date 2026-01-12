-- ========================================
-- MIGRATION 000078: DROP v3_workorder_views (DOWN)
-- ========================================
-- 
-- Purpose: Drop the views created in the UP migration
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_workorder_metrics: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_workorder_metrics;

-- -------------------------------------------------------------------
-- v3_workorder_list: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_workorder_list;
