-- ========================================
-- MIGRACIÓN 000072: CORREGIR VISTAS DEPRECADAS Y DUPLICACIONES (ROLLBACK)
-- Entidad: views (Corregir vistas deprecadas y eliminar duplicaciones)
-- Funcionalidad: Eliminar vistas corregidas
-- ========================================

-- Eliminar todas las vistas corregidas creadas
DROP VIEW IF EXISTS dashboard_management_balance_view_v2;
DROP VIEW IF EXISTS dashboard_operating_result_view_v2;
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2;
DROP VIEW IF EXISTS workorder_metrics_view_v2;
DROP VIEW IF EXISTS labor_cards_cube_view_v2;
