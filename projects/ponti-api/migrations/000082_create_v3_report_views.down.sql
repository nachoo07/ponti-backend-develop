-- ========================================
-- MIGRATION 000082: DROP v3_report_views (DOWN)
-- ========================================
-- 
-- Purpose: Drop the views created in the UP migration
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics_view;

-- -------------------------------------------------------------------
-- v3_report_summary_results_view: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_report_summary_results_view;

-- -------------------------------------------------------------------
-- v3_investor_contribution_data_view: rollback elimina la vista
-- -------------------------------------------------------------------
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view;
