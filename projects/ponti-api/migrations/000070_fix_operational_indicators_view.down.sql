-- ========================================
-- MIGRACIÓN 000070: ROLLBACK CORRECCIÓN VISTA OPERATIONAL INDICATORS
-- ========================================

-- Eliminar función
DROP FUNCTION IF EXISTS calculate_campaign_closing_date(DATE);

-- Restaurar vista original con NULLs
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;

CREATE VIEW dashboard_operational_indicators_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  NULL::DATE AS start_date,
  NULL::DATE AS end_date,
  NULL::DATE AS campaign_closing_date
FROM projects p
WHERE p.deleted_at IS NULL;
