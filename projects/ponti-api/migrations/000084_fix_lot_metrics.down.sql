-- ========================================
-- MIGRATION 000084: FIX SSOT CONSISTENCY IN v3_lot_metrics (DOWN)
-- ========================================
-- 
-- Purpose: Revert v3_lot_metrics to original version
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Revertir a la versión original de v3_lot_metrics
CREATE OR REPLACE VIEW public.v3_lot_metrics AS
WITH lot_base AS (
  SELECT
    l.id  AS lot_id,
    f.id  AS field_id,
    f.project_id,
    l.hectares,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric) AS sowed_area_ha,
    v3_calc.harvested_area(l.tons::numeric, l.hectares::numeric) AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
-- Superficie total del campo (suma de todos los lotes del campo)
field_total_area AS (
  SELECT 
    f.id AS field_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.id
)
SELECT
  b.project_id,
  b.field_id,
  b.lot_id,
  b.hectares,
  b.sowed_area_ha,
  b.harvested_area_ha,
  v3_calc.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
  COALESCE(v3_calc.labor_cost_for_lot(b.lot_id), 0)::numeric      AS labor_cost_usd,
  COALESCE(v3_calc.supply_cost_for_lot(b.lot_id), 0)::numeric     AS supplies_cost_usd,
  COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), 0)::numeric     AS direct_cost_usd,
  COALESCE(v3_calc.income_net_total_for_lot(b.lot_id), 0)::numeric    AS income_net_total_usd,
  COALESCE(v3_calc.income_net_per_ha_for_lot(b.lot_id), 0)::numeric   AS income_net_per_ha_usd,
  COALESCE(v3_calc.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric   AS admin_cost_per_ha_usd,
  COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0)::numeric        AS rent_per_ha_usd,
  COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0)::numeric AS active_total_per_ha_usd,
  COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,
  -- Totales
  (COALESCE(v3_calc.admin_cost_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric   AS admin_total_usd,
  (COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric        AS rent_total_usd,
  (COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric AS active_total_usd,
  (COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric AS operating_result_total_usd,
  -- Per-ha usando wrapper SSOT
  v3_calc.cost_per_ha(COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), 0)::numeric, b.hectares::numeric) AS direct_cost_per_ha_usd,
  -- Superficie total del campo
  COALESCE(fta.total_hectares, 0)::numeric AS superficie_total
FROM lot_base b
LEFT JOIN field_total_area fta ON fta.field_id = b.field_id;
