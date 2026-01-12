-- ========================================
-- MIGRACIÓN 000125: FIX workorder number in operational indicators (DOWN)
-- ========================================

BEGIN;

-- Restaurar funciones originales que retornaban ID

DROP FUNCTION IF EXISTS v3_dashboard_ssot.first_workorder_number_for_project(bigint);
DROP FUNCTION IF EXISTS v3_dashboard_ssot.last_workorder_number_for_project(bigint);

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.first_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

-- Restaurar vista original

CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  
  -- Fechas operativas
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  
  -- IDs de workorders
  v3_dashboard_ssot.first_workorder_id_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_id_for_project(p.id) AS last_workorder_id,
  
  -- Fecha último arqueo
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
  
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMIT;
