-- ========================================
-- ROLLBACK: MIGRACIÓN 000072 - VISTAS PARA REPORTES
-- ========================================

-- Eliminar vista de métricas de campo/cultivo
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2;
