-- ========================================
-- MIGRATION 000190: FIX Dashboard Operating Result Rent Calculation (UP)
-- ========================================
--
-- Purpose: Adjust v3_dashboard_ssot.operating_result_total_for_project to use rent_fixed_only_for_lot()
--          ensuring dashboard matches lot/report calculations (controls 11-13).
--
-- Date: 2025-11-08
-- Author: System
--
-- Note: Code in English, comentarios en español.

BEGIN;

-- ============================================================================
-- Actualizar función: v3_dashboard_ssot.operating_result_total_for_project
-- ============================================================================
-- Usar rent_fixed_only_for_lot() por cada lote para mantener consistencia.
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.operating_result_total_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::double precision AS total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares),
        0
      )::double precision AS total_lease
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
  )
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados
    v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
    -
    -- Arriendo fijo total (solo parte capitalizable)
    (SELECT total_lease FROM lease_cost)
    -
    -- Estructura total = admin_cost × hectáreas totales
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 
'Calcula resultado operativo total usando rent_fixed_only_for_lot(). FIX 000190: Dashboard consistente con lotes/reports.';

COMMIT;
