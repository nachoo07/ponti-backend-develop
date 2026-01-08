-- ========================================
-- MIGRATION 000108: FIX PERCENTAGE PRECISION IN CROP INCIDENCE (DOWN)
-- ========================================
-- 
-- Purpose: Revertir corrección de precisión decimal en porcentajes
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- REVERTIR VISTA v3_dashboard_crop_incidence A VERSIÓN ANTERIOR
-- ========================================
-- Nota: Volver a usar función sin redondeo
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
  -- Volver a función original sin redondeo
  v3_calc.percentage(bc.crop_hectares, pt.total_project_hectares) AS crop_incidence_pct,
  -- FIX: Usar función SSOT para costo directo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop_ssot(bc.project_id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

-- ========================================
-- ELIMINAR FUNCIÓN DE PORCENTAJE CON REDONDEO
-- ========================================
-- Nota: Limpiar función creada en migración UP
DROP FUNCTION IF EXISTS v3_calc.percentage_rounded(numeric, numeric);
