-- ========================================
-- MIGRACIÓN 000123: CREATE v3_dashboard_crop_incidence VIEW (UP)
-- ========================================
-- 
-- Propósito: Vista de incidencia de costos por cultivo (SOLO ensamblaje con SSOT)
-- Dependencias: Requiere v3_core_ssot (000113), v3_dashboard_ssot (000116)
-- Arquitectura: Vista que SOLO ensambla, NO calcula (usa funciones SSOT)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY FASE 2:
-- - Todos los cálculos inline (CTEs complejos) movidos a v3_dashboard_ssot.crop_incidence_for_project()
-- - Vista solo ensambla resultados de funciones SSOT
-- - NO hay hardcodeos ni cálculos inline
-- 
-- Nota: Vistas SOLO ensamblan, NO calculan (usan funciones SSOT)

BEGIN;

-- ========================================
-- CREAR VISTA: v3_dashboard_crop_incidence
-- ========================================
-- Propósito: Incidencia de costos por cultivo (se mantiene separada por ser 1:N por crop)

CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS
SELECT
  p.id AS project_id,
  ci.current_crop_id,
  ci.crop_name,
  ci.crop_hectares,
  ci.crop_incidence_pct,
  v3_dashboard_ssot.cost_per_ha_for_crop_ssot(p.id, ci.current_crop_id)::numeric AS cost_per_ha_usd
FROM public.projects p
CROSS JOIN LATERAL v3_dashboard_ssot.crop_incidence_for_project(p.id) ci
WHERE p.deleted_at IS NULL
ORDER BY p.id, ci.crop_name;

COMMENT ON VIEW public.v3_dashboard_crop_incidence IS 'Módulo: Incidencia de costos por cultivo (SOLO ensamblaje SSOT)';

COMMIT;
