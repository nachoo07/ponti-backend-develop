-- ========================================
-- MIGRATION 000107: CREATE SSOT COST PER HA FOR CROP (UP)
-- ========================================
-- 
-- Purpose: Crear función SSOT para costo por hectárea por cultivo usando solo órdenes de trabajo directas
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- FUNCIÓN SSOT PARA COSTO POR HECTÁREA POR CULTIVO
-- ========================================
-- Nota: Usa funciones SSOT sin movimientos internos para calcular costos directos por cultivo
-- Fórmula: (labor_cost_mb + seeds_cost_mb + agro_cost_mb) / hectares_del_cultivo
CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_crop_ssot(p_project_id bigint, p_crop_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.per_ha_dp(
    -- Suma de costos directos usando funciones SSOT (sin movimientos internos)
    COALESCE(
      (SELECT COALESCE(SUM(
        v3_calc.labor_cost_for_lot_mb(l.id) + 
        v3_calc.supply_cost_seeds_for_lot_mb(l.id) + 
        v3_calc.supply_cost_agrochemicals_for_lot_mb(l.id)
      ), 0)::double precision
       FROM public.lots l
       JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
       WHERE f.project_id = p_project_id 
         AND l.current_crop_id = p_crop_id 
         AND l.deleted_at IS NULL)
    , 0),
    -- Superficie total del cultivo
    (SELECT COALESCE(SUM(l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;

-- ========================================
-- ACTUALIZAR VISTA v3_dashboard_crop_incidence CON FUNCIÓN SSOT
-- ========================================
-- Nota: Usar la nueva función SSOT para costos por hectárea
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
  -- FIX: Usar función SSOT para costo directo por hectárea del cultivo
  v3_calc.cost_per_ha_for_crop_ssot(bc.project_id, bc.current_crop_id)::numeric AS cost_per_ha_usd
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

