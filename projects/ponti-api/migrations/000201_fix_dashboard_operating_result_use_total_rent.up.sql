-- ========================================
-- MIGRACIÓN 000201: FIX Dashboard Operating Result - Use Total Rent (UP)
-- ========================================
--
-- Propósito: Corregir función operating_result_total_for_project para usar arriendo TOTAL
-- Problema: La función usa rent_fixed_only para calcular operating_result,
--           pero debe usar rent_total (fijo + variable) para coincidir con
--           v3_lot_metrics, v3_report_summary_results_view y v3_report_field_crop_economicos
-- Solución: Cambiar rent_fixed_only_for_lot() por rent_per_ha_for_lot() en el cálculo
-- Fecha: 2025-11-17
-- Autor: Sistema
--
-- Impacto: Control 13 pasará correctamente
--          Dashboard coincidirá con lotes y reportes en resultado operativo
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- ACTUALIZAR: v3_dashboard_ssot.operating_result_total_for_project
-- Cambio: Usar rent_per_ha_for_lot() (arriendo TOTAL) en lugar de rent_fixed_only_for_lot()
-- ========================================

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
  -- ========================================
  -- FIX 000201: Arriendo TOTAL (para Resultado Operativo)
  -- ========================================
  -- Incluye fijo + variable (% sobre ingresos)
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v3_lot_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
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
    -- ========================================
    -- FIX 000201: Arriendo TOTAL (fijo + variable)
    -- ========================================
    (SELECT total_lease FROM lease_cost)
    -
    -- Estructura total = admin_cost × hectáreas totales
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 
'Calcula resultado operativo total usando rent_per_ha_for_lot() (arriendo TOTAL). FIX 000201: Dashboard consistente con lotes/reports.';

COMMIT;

