-- ========================================
-- MIGRACIÓN 000128: Corregir cálculo de ACTIVO TOTAL en v3_lot_list
-- ========================================
-- 
-- Propósito: Corregir el cálculo de active_total_per_ha_usd para que use el mismo
--            costo directo que se muestra en cost_usd_per_ha (por cultivo)
-- 
-- Problema: active_total_per_ha_usd usaba direct_cost_per_ha_usd (promedio del proyecto)
--           pero cost_usd_per_ha muestra cost_per_ha_for_crop_ssot (por cultivo)
--           Esto causaba una diferencia de ~$10 en el total
-- 
-- Solución: Calcular active_total directamente como:
--           cost_per_ha_for_crop_ssot + rent_per_ha + admin_cost_per_ha
-- 
-- Fórmula: ACTIVO TOTAL = costo directo + arriendo + estructura
-- 
-- Fecha: 2025-10-08
-- Autor: Sistema
-- ========================================

CREATE OR REPLACE VIEW public.v3_lot_list AS
WITH base AS (
  SELECT
    f.project_id,
    p.name AS project_name,
    f.id AS field_id,
    f.name AS field_name,
    l.id AS lot_id,
    l.name AS lot_name,
    l.variety,
    l.season,
    l.previous_crop_id,
    prev_crop.name AS previous_crop,
    l.current_crop_id,
    curr_crop.name AS current_crop,
    l.hectares,
    l.updated_at,
    l.sowing_date,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
  LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
wo_dates AS (
  SELECT
    w.lot_id,
    MIN(w.date) AS raw_sowing_date
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
  GROUP BY w.lot_id
),
lot_metrics_data AS (
  -- Obtener métricas desde v3_lot_metrics (áreas, costos, etc)
  SELECT
    project_id,
    lot_id,
    sowed_area_ha,
    harvested_area_ha,
    yield_tn_per_ha,
    direct_cost_per_ha_usd,
    direct_cost_usd,
    income_net_total_usd,
    income_net_per_ha_usd,
    rent_per_ha_usd,
    admin_cost_per_ha_usd,
    operating_result_per_ha_usd,
    rent_total_usd,
    admin_total_usd,
    operating_result_total_usd
  FROM v3_lot_metrics
)
SELECT
  b.project_id,
  b.project_name,
  b.field_id,
  b.field_name,
  b.lot_id AS id,
  b.lot_name,
  b.variety,
  b.season,
  b.previous_crop_id,
  b.previous_crop,
  b.current_crop_id,
  b.current_crop,
  b.hectares,
  b.updated_at,
  
  -- Áreas (desde v3_lot_metrics)
  lm.sowed_area_ha,
  lm.harvested_area_ha,
  
  -- Rendimiento (desde v3_lot_metrics)
  lm.yield_tn_per_ha,
  
  -- Costo por ha (por cultivo, usa v3_dashboard_ssot para coincidir con Dashboard)
  v3_dashboard_ssot.cost_per_ha_for_crop_ssot(b.project_id, b.current_crop_id)::numeric AS cost_usd_per_ha,
  
  -- Ingresos y otros costos por ha (desde v3_lot_metrics)
  lm.income_net_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  
  -- FIX: Activo total por ha = costo_crop + arriendo + admin
  (v3_dashboard_ssot.cost_per_ha_for_crop_ssot(b.project_id, b.current_crop_id)::numeric 
   + COALESCE(lm.rent_per_ha_usd, 0) 
   + COALESCE(lm.admin_cost_per_ha_usd, 0))::numeric AS active_total_per_ha_usd,
  
  lm.operating_result_per_ha_usd,
  
  -- Totales por lote (desde v3_lot_metrics)
  lm.income_net_total_usd,
  lm.direct_cost_usd AS direct_cost_total_usd,
  lm.rent_total_usd,
  lm.admin_total_usd,
  
  -- FIX: Activo total = (costo_crop + arriendo + admin) × hectáreas
  ((v3_dashboard_ssot.cost_per_ha_for_crop_ssot(b.project_id, b.current_crop_id)::numeric 
    + COALESCE(lm.rent_per_ha_usd, 0) 
    + COALESCE(lm.admin_cost_per_ha_usd, 0)) 
   * b.hectares)::numeric AS active_total_usd,
  
  lm.operating_result_total_usd,
  
  -- Fechas
  b.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  b.tons,
  wd.raw_sowing_date
  
FROM base b
LEFT JOIN wo_dates wd ON wd.lot_id = b.lot_id
LEFT JOIN lot_metrics_data lm ON lm.lot_id = b.lot_id;

COMMENT ON VIEW public.v3_lot_list IS 
'Listado de lotes - active_total = cost_per_ha_for_crop_ssot + arriendo + admin';
