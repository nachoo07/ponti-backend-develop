-- ========================================
-- MIGRACIÓN 000175: FIX Lease Functions Use Lot Calculations (DOWN)
-- ========================================
--
-- Propósito: Revertir corrección de funciones de arriendo
-- Fecha: 2025-11-03
-- Autor: Sistema

BEGIN;

-- Restaurar versiones anteriores de las funciones (migración 000116)

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE((
    SELECT f.lease_type_value * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
    FROM public.fields f
    WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
    LIMIT 1
  ), 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_executed_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE((
    SELECT CASE 
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
      ELSE 0
    END
    FROM public.fields f
    WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
    LIMIT 1
  ), 0)::double precision
$$;

COMMIT;

