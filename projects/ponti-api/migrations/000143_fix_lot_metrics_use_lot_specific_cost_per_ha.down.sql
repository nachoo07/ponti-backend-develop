-- ========================================
-- MIGRACIÓN 000143: FIX v3_lot_metrics - Use lot-specific cost per ha (DOWN)
-- ========================================
-- 
-- Propósito: Revertir cambio de cálculo de direct_cost_per_ha_usd
-- Revierte a: Usar promedio del proyecto (comportamiento anterior)
-- Fecha: 2025-10-13
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RESTAURAR v3_lot_metrics A VERSIÓN ANTERIOR
-- ========================================
DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

CREATE OR REPLACE VIEW public.v3_lot_metrics AS
/* --------------------------------------------------------------------
   CTE: base - Datos básicos del lote
   - Calcula áreas sembradas y cosechadas desde workorders
   - sowed_area_ha = suma de effective_area de workorders de siembra (category_id = 9)
   - harvested_area_ha = suma de effective_area de workorders de cosecha (category_id = 13)
-------------------------------------------------------------------- */
WITH base AS (
  SELECT
    f.project_id,
    l.id              AS lot_id,
    l.name            AS lot_name,
    l.hectares,
    l.tons,
    l.sowing_date,
    -- Área sembrada desde workorders de siembra
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0)::numeric AS sowed_area_ha,
    -- Área cosechada desde workorders de cosecha
    COALESCE(SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END), 0)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  GROUP BY f.project_id, l.id, l.name, l.hectares, l.tons, l.sowing_date
),
/* --------------------------------------------------------------------
   CTE: workorder_costs - Costos desde v3_workorder_metrics (SSOT único)
   - ELIMINA cálculos duplicados de v3_lot_ssot
   - USA directamente v3_workorder_metrics que ya calcula todo correctamente
-------------------------------------------------------------------- */
workorder_costs AS (
  SELECT
    lot_id,
    COALESCE(labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(direct_cost_usd, 0)::numeric AS direct_cost_usd
  FROM v3_workorder_metrics
),
/* --------------------------------------------------------------------
   CTE: project_total_direct_cost
   - Calcula el costo directo total del proyecto desde v3_workorder_metrics
   - ÚNICA fuente de verdad para costos directos
-------------------------------------------------------------------- */
project_total_direct_cost AS (
  SELECT
    p.id AS project_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares,
    COALESCE(SUM(wc.direct_cost_usd), 0)::numeric AS total_direct_cost
  FROM public.projects p
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorder_costs wc ON wc.lot_id = l.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
)
/* --------------------------------------------------------------------
   SELECT PRINCIPAL: Ensamblado final de todas las métricas
   - Combina datos de base, workorder_costs, y cálculos adicionales
-------------------------------------------------------------------- */
SELECT
  -- ############# IDENTIFICADORES ##############
  b.project_id,
  b.lot_id,
  b.lot_name,
  b.hectares,

  -- ############# ÁREAS ##############
  b.sowed_area_ha,
  b.harvested_area_ha,

  -- ############# RENDIMIENTO ##############
  v3_lot_ssot.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
  b.tons,

  -- ############# FECHAS ##############
  b.sowing_date,

  -- ############# COSTOS DIRECTOS (desde v3_workorder_metrics) ##############
  COALESCE(wc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(wc.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  COALESCE(wc.direct_cost_usd, 0)::numeric AS direct_cost_usd,

  -- ############# INGRESOS (desde v3_lot_ssot) ##############
  COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric AS income_net_total_usd,

  -- ############# COSTOS POR HECTÁREA ##############
  -- Ingreso neto por ha
  v3_core_ssot.per_ha(
    COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
    b.hectares::numeric
  ) AS income_net_per_ha_usd,
  
  -- Arriendo por ha
  COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
  
  -- Admin por ha
  COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_cost_per_ha_usd,
  
  -- Activo total por ha
  COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)::numeric AS active_total_per_ha_usd,
  
  -- Resultado operativo por ha
  COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,

  -- ############# TOTALES ##############
  -- Arriendo total
  (COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS rent_total_usd,
  
  -- Admin total
  (COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS admin_total_usd,
  
  -- Activo total
  (COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS active_total_usd,
  
  -- Resultado operativo total
  (COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS operating_result_total_usd,

  -- ############# COSTO DIRECTO POR HA (PROMEDIO DEL PROYECTO) ##############
  -- Fórmula: total_direct_cost_proyecto / hectareas_proyecto
  -- Usa v3_workorder_metrics como ÚNICA fuente
  v3_core_ssot.cost_per_ha(
    COALESCE(ptdc.total_direct_cost, 0)::numeric,
    COALESCE(ptdc.total_hectares, 0)::numeric
  ) AS direct_cost_per_ha_usd,

  -- ############# SUPERFICIE TOTAL DEL PROYECTO ##############
  COALESCE(ptdc.total_hectares, 0)::numeric AS project_total_hectares

FROM base b
LEFT JOIN workorder_costs wc ON wc.lot_id = b.lot_id
LEFT JOIN project_total_direct_cost ptdc ON ptdc.project_id = b.project_id;

COMMIT;

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas agregadas por lote usando v3_workorder_metrics como SSOT de costos directos';

