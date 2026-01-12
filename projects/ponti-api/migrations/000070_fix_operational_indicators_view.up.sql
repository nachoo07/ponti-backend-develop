-- ========================================
-- MIGRACIÓN 000070: CORREGIR VISTA OPERATIONAL INDICATORS
-- Entidad: views (Corregir vista operational indicators)
-- Funcionalidad: Corregir la vista para que calcule fechas reales en lugar de NULL hardcodeados
-- ========================================

-- Crear función para calcular fecha de cierre de campaña
CREATE OR REPLACE FUNCTION calculate_campaign_closing_date(end_date DATE)
RETURNS DATE AS $$
BEGIN
  IF end_date IS NULL THEN
    RETURN NULL;
  END IF;
  
  RETURN end_date + (get_campaign_closure_days() || ' days')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Corregir la vista dashboard_operational_indicators_view_v2
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2;

CREATE VIEW dashboard_operational_indicators_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  w_min.min_date AS start_date,
  w_max.max_date AS end_date,
  calculate_campaign_closing_date(w_max.max_date) AS campaign_closing_date
FROM projects p
LEFT JOIN (
  SELECT project_id, MIN(date) as min_date 
  FROM workorders 
  WHERE deleted_at IS NULL 
  GROUP BY project_id
) w_min ON w_min.project_id = p.id
LEFT JOIN (
  SELECT project_id, MAX(date) as max_date 
  FROM workorders 
  WHERE deleted_at IS NULL 
  GROUP BY project_id
) w_max ON w_max.project_id = p.id
WHERE p.deleted_at IS NULL;
