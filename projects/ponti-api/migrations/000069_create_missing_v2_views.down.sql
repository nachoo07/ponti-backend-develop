-- ========================================
-- MIGRACIÓN 000069: REVERTIR VISTAS _v2 FALTANTES
-- Entidad: views (Revertir vistas _v2 que faltan)
-- Funcionalidad: Eliminar las vistas _v2 creadas
-- ========================================

DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2;
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2;
