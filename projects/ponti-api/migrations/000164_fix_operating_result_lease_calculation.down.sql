-- ========================================
-- MIGRATION 000164: FIX OPERATING RESULT LEASE CALCULATION (DOWN)
-- ========================================
-- 
-- Purpose: Revertir la corrección del cálculo del arriendo
-- Date: 2025-10-21
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Revertir a la versión anterior (con bug)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.operating_result_total_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::double precision as total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        CASE 
          WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * pt.total_hectares
          ELSE 0
        END, 
        0
      )::double precision as total_lease
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    CROSS JOIN project_totals pt
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    LIMIT 1
  )
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados (desde v3_workorder_metrics)
    v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
    -
    -- Arriendo total
    (SELECT total_lease FROM lease_cost)
    -
    -- Estructura (admin) total
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 
'Calcula el resultado operativo total del proyecto. Estructura = admin_cost × hectáreas.';

