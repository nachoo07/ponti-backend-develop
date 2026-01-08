-- ========================================
-- MIGRATION 000086: FIX active_total_per_ha_for_lot (UP)
-- ========================================
-- 
-- Purpose: Fix active_total_per_ha_for_lot to use projects.admin_cost directly
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- CORREGIR active_total_per_ha_for_lot
-- ========================================
-- Cambiar para usar projects.admin_cost directamente sin prorrateo
-- Esto hace que sea consistente con las vistas v3 que ya usan project.admin_cost

CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.active_total_per_ha(
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.rent_per_ha_for_lot(p_lot_id),
           COALESCE(p.admin_cost, 0)::double precision
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Comentario para documentar el cambio
COMMENT ON FUNCTION v3_calc.active_total_per_ha_for_lot(bigint) IS 
'Calcula el total activo por hectárea para un lote sumando: costo directo/ha + renta/ha + admin_cost del proyecto (sin prorrateo)';
