-- ========================================
-- MIGRATION 000199: FIX Operational Indicators Workorder Numbers As Text (UP)
-- ========================================
--
-- Propósito: El dashboard necesita mostrar el número visible de OT
--            (con ceros a la izquierda o prefijos) y hoy lo castea a bigint.
--            Esta migración recrea las funciones/view para retornar TEXT.

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_operational_indicators CASCADE;

DROP FUNCTION IF EXISTS v3_dashboard_ssot.first_workorder_number_for_project(bigint);
DROP FUNCTION IF EXISTS v3_dashboard_ssot.last_workorder_number_for_project(bigint);

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.first_workorder_number_for_project IS
'Retorna número visible (texto) de la primera orden de trabajo del proyecto.';

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS text
LANGUAGE sql STABLE AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.last_workorder_number_for_project IS
'Retorna número visible (texto) de la última orden de trabajo del proyecto.';

CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  v3_dashboard_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_operational_indicators IS
'Indicadores operativos con números de OT como texto (sin castear a bigint).';

COMMIT;

