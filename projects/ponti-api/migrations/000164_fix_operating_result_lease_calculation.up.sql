-- ========================================
-- MIGRATION 000164: FIX OPERATING RESULT LEASE CALCULATION (UP)
-- ========================================
-- 
-- Purpose: Corregir el cálculo del arriendo en la función operating_result_total_for_project
-- Date: 2025-10-21
-- Author: System
-- 
-- Problema identificado en Control 13:
--   La función v3_dashboard_ssot.operating_result_total_for_project solo calcula arriendo
--   para tipos fijos (lease_type_id IN (3,4)) pero NO para tipos porcentuales (1,2)
--
-- Impacto:
--   - Proyectos con arriendo % INGRESO NETO (tipo 1) muestran resultado inflado
--   - Proyectos con arriendo % UTILIDAD (tipo 2) muestran resultado inflado
--   - Diferencia encontrada: -3574.69 USD en proyecto 11
--
-- Corrección:
--   Usar el cálculo de arriendo desde los lotes (v3_lot_ssot.rent_per_ha_for_lot)
--   que SÍ calcula correctamente TODOS los tipos de arriendo
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- RECREAR FUNCIÓN: v3_dashboard_ssot.operating_result_total_for_project
-- ============================================================================

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
  -- CORREGIDO: Calcular arriendo sumando desde cada lote (incluye tipos porcentuales)
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v3_lot_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
        0
      )::double precision as total_lease
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
    -- Costos directos ejecutados (desde v3_workorder_metrics)
    v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
    -
    -- Arriendo total (ahora calcula TODOS los tipos correctamente)
    (SELECT total_lease FROM lease_cost)
    -
    -- Estructura (admin) total = admin_cost × total_hectares
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 
'Calcula el resultado operativo total del proyecto. CORREGIDO 164: Arriendo calculado desde lotes (incluye tipos porcentuales 1,2).';

