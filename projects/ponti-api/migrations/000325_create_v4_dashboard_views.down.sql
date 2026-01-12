-- ========================================
-- MIGRACIÓN 000325: CREATE v4_report.dashboard_* views (DOWN)
-- ========================================
-- Rollback: eliminar vistas v4_report.dashboard_*

BEGIN;

DROP VIEW IF EXISTS v4_report.dashboard_contributions_progress;
DROP VIEW IF EXISTS v4_report.dashboard_operational_indicators;
DROP VIEW IF EXISTS v4_report.dashboard_crop_incidence;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance;
DROP VIEW IF EXISTS v4_report.dashboard_metrics;

COMMIT;
