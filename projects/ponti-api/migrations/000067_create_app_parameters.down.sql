-- ========================================
-- MIGRACIÓN 000067: REVERTIR PARÁMETROS DE APLICACIÓN
-- ========================================

-- Eliminar funciones
DROP FUNCTION IF EXISTS get_iva_percentage();
DROP FUNCTION IF EXISTS get_campaign_closure_days();
DROP FUNCTION IF EXISTS get_default_fx_rate();
DROP FUNCTION IF EXISTS get_app_parameter(VARCHAR);
DROP FUNCTION IF EXISTS get_app_parameter_decimal(VARCHAR);
DROP FUNCTION IF EXISTS get_app_parameter_integer(VARCHAR);
DROP FUNCTION IF EXISTS calculate_campaign_closing_date(DATE);

-- Eliminar vistas
DROP VIEW IF EXISTS fix_labors_list;
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2;

-- Eliminar índices
DROP INDEX IF EXISTS idx_app_parameters_key;
DROP INDEX IF EXISTS idx_app_parameters_category;

-- Eliminar tabla
DROP TABLE IF EXISTS app_parameters;
