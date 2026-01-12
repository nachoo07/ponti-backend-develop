-- ========================================
-- MIGRATION 000106: FIX v3_dashboard_crop_incidence CALCULATION (UP)
-- ========================================
-- 
-- Purpose: Corregir cálculo de incidencia por superficie - problema en lógica de agrupación
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- VISTA v3_dashboard_crop_incidence CORREGIDA
-- ========================================
-- Nota: Incidencia de superficie por cultivo dentro del proyecto
-- FIX: Calcular incidencia correctamente - superficie de cada cultivo / superficie total del proyecto
CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS
WITH lot_base AS (
  SELECT
    l.id               AS lot_id,
    f.project_id       AS project_id,
    l.current_crop_id  AS current_crop_id,
    c.name             AS crop_name,
    l.hectares         AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares IS NOT NULL AND l.hectares > 0
),
-- FIX: Calcular superficie total del proyecto PRIMERO (todos los lotes)
project_totals AS (
  SELECT 
    project_id,
    SUM(hectares)::numeric AS total_project_hectares
  FROM lot_base
  GROUP BY project_id
),
-- FIX: Agrupar por cultivo solo los lotes que tienen cultivo asignado
by_crop AS (
  SELECT 
    project_id, 
    current_crop_id, 
    crop_name, 
    SUM(hectares)::numeric AS crop_hectares
  FROM lot_base
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.crop_hectares,
  -- FIX: Calcular incidencia correctamente: superficie del cultivo / superficie total del proyecto
  v3_calc.percentage(bc.crop_hectares, pt.total_project_hectares) AS crop_incidence_pct,
  -- Costo directo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop(bc.project_id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;
