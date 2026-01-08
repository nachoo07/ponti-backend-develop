-- ========================================
-- MIGRATION 000105: FIX v3_dashboard_crop_incidence (DOWN)
-- ========================================
-- 
-- Purpose: Revertir corrección de incidencia por superficie
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Revertir a la implementación anterior (por costos)
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
by_crop AS (
  SELECT 
    project_id, 
    current_crop_id, 
    crop_name, 
    SUM(hectares)::numeric AS crop_hectares,
    -- FIX: Calcular costos por cultivo
    v3_calc.total_costs_for_crop(project_id, current_crop_id) AS crop_costs_usd
  FROM lot_base
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
total_by_project AS (
  SELECT 
    project_id, 
    SUM(crop_costs_usd)::numeric AS total_costs_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.crop_hectares,
  -- FIX: Calcular incidencia por costos en lugar de hectáreas
  v3_calc.percentage(bc.crop_costs_usd::numeric, t.total_costs_usd) AS crop_incidence_pct,
  -- IMPLEMENTAR: Costo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop(bc.project_id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN total_by_project t ON t.project_id = bc.project_id;