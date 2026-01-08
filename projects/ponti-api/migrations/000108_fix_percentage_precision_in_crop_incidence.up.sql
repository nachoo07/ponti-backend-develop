-- ========================================
-- MIGRATION 000108: FIX PERCENTAGE PRECISION IN CROP INCIDENCE (UP)
-- ========================================
-- 
-- Purpose: Corregir problema de precisión decimal y ajustar automáticamente para 100%
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- FUNCIÓN MEJORADA PARA CÁLCULO DE PORCENTAJES CON PRECISIÓN
-- ========================================
-- Nota: Redondea a 3 decimales para evitar problemas de precisión flotante
CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT ROUND(v3_calc.safe_div($1, $2) * 100, 3)
$$;

-- ========================================
-- VISTA v3_dashboard_crop_incidence CON AJUSTE INTELIGENTE
-- ========================================
-- Nota: Ajusta automáticamente el último cultivo para que sume exactamente 100%
-- Solo cuando la diferencia es ≤ 1% (errores de redondeo), sino muestra valores reales
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
),
-- FIX: Calcular porcentajes base y verificar si necesitan ajuste
crop_percentages AS (
  SELECT
    bc.project_id,
    bc.current_crop_id,
    bc.crop_name,
    bc.crop_hectares,
    pt.total_project_hectares,
    -- Calcular porcentaje base redondeado
    v3_calc.percentage_rounded(bc.crop_hectares, pt.total_project_hectares) AS base_percentage,
    -- Enumerar cultivos para identificar el último
    ROW_NUMBER() OVER (PARTITION BY bc.project_id ORDER BY bc.crop_name) AS crop_order,
    COUNT(*) OVER (PARTITION BY bc.project_id) AS total_crops
  FROM by_crop bc
  JOIN project_totals pt ON pt.project_id = bc.project_id
),
-- FIX: Calcular suma total y diferencia para decidir si ajustar
project_sums AS (
  SELECT
    project_id,
    SUM(base_percentage) AS total_percentage,
    ABS(100.000 - SUM(base_percentage)) AS difference_from_100
  FROM crop_percentages
  GROUP BY project_id
),
-- FIX: Ajuste condicional - solo si la diferencia es ≤ 1%
adjusted_percentages AS (
  SELECT
    cp.project_id,
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    cp.total_project_hectares,
    cp.base_percentage,
    ps.difference_from_100,
    CASE 
      -- Solo ajustar si es el último cultivo Y la diferencia es pequeña (≤ 1%)
      WHEN cp.crop_order = cp.total_crops AND ps.difference_from_100 <= 1.000 THEN
        -- Ajuste para completar exactamente 100%
        100.000 - COALESCE((
          SELECT SUM(base_percentage) 
          FROM crop_percentages cp2 
          WHERE cp2.project_id = cp.project_id 
            AND cp2.crop_order < cp.crop_order
        ), 0)
      ELSE
        -- Mostrar valor real calculado (sin ajuste)
        cp.base_percentage
    END AS crop_incidence_pct
  FROM crop_percentages cp
  JOIN project_sums ps ON ps.project_id = cp.project_id
)
SELECT
  ap.project_id,
  ap.current_crop_id,
  ap.crop_name,
  ap.crop_hectares,
  ap.crop_incidence_pct,
  -- FIX: Usar función SSOT para costo directo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop_ssot(ap.project_id, ap.current_crop_id)::numeric AS cost_per_ha_usd
FROM adjusted_percentages ap
ORDER BY ap.project_id, ap.crop_name;
