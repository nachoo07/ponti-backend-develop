-- ========================================
-- MIGRACIÓN 000126: FIX admin_cost en v3_lot_ssot (DOWN)
-- ========================================
-- 
-- Propósito: Revertir corrección de admin_cost_per_ha_for_lot
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- REVERTIR FUNCIÓN SSOT: admin_cost_per_ha_for_lot
-- ========================================
-- Volver a la versión anterior que dividía por hectáreas
CREATE OR REPLACE FUNCTION v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::double precision / t.total_hectares
              ELSE 0 END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    -- Calcular total_hectares inline para evitar dependencia circular
    SELECT COALESCE(
      (SELECT SUM(l2.hectares)
       FROM public.lots l2
       JOIN public.fields f2 ON f2.id = l2.field_id AND f2.deleted_at IS NULL
       WHERE f2.project_id = f.project_id AND l2.deleted_at IS NULL), 
      0)::double precision AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

COMMIT;
