-- =============================================================================
-- MIGRACIÓN 000305: v4_report.lot_metrics - Paridad exacta con v3_lot_metrics
-- =============================================================================
--
-- Propósito: Vista con paridad exacta con v3_lot_metrics (000202)
-- Fecha: 2025-01-XX
-- Autor: Sistema
--
-- Semántica de rent:
-- - rent_per_ha_usd: arriendo FIJO (expuesta, para UI y active_total)
-- - rent_total_per_ha_usd: arriendo FIJO + % (NO expuesta, solo para operating_result)
--

CREATE OR REPLACE VIEW v4_report.lot_metrics AS
WITH 
project_totals AS (
  SELECT project_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY project_id
),
field_totals AS (
  SELECT field_id, SUM(hectares)::numeric AS total_hectares
  FROM v4_calc.lot_base_costs GROUP BY field_id
)
SELECT
  c.project_id,
  c.field_id,
  c.lot_id,
  c.lot_name,
  c.hectares,
  c.sowed_area_ha,
  c.harvested_area_ha,
  c.yield_tn_per_ha,
  c.tons,
  c.sowing_date,
  c.labor_cost_usd,
  c.supplies_cost_usd,
  c.direct_cost_usd,
  c.income_net_total_usd,
  c.income_net_per_ha_usd,
  -- rent_per_ha_usd: EXPUESTA (000202:106)
  c.rent_per_ha_usd,
  c.admin_cost_per_ha_usd,
  -- active_total usa rent_per_ha (FIJO) - 000202:108
  (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  -- operating_result usa rent_total_per_ha (FIJO + %) - 000202:109
  (c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_total_per_ha_usd + c.admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  -- rent_total_usd = rent_per_ha (FIJO) × hectares - 000202:110
  (c.rent_per_ha_usd * c.hectares)::numeric AS rent_total_usd,
  (c.admin_cost_per_ha_usd * c.hectares)::numeric AS admin_total_usd,
  -- active_total_usd usa rent_per_ha (FIJO) - 000202:112
  ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd) * c.hectares)::numeric AS active_total_usd,
  -- operating_result_total usa rent_total_per_ha (FIJO + %) - 000202:113
  ((c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_total_per_ha_usd + c.admin_cost_per_ha_usd)) * c.hectares)::numeric AS operating_result_total_usd,
  c.direct_cost_per_ha_usd,
  COALESCE(pt.total_hectares, 0) AS project_total_hectares,
  COALESCE(ft.total_hectares, 0) AS field_total_hectares
FROM v4_calc.lot_base_costs c
LEFT JOIN project_totals pt ON pt.project_id = c.project_id
LEFT JOIN field_totals ft ON ft.field_id = c.field_id;

COMMENT ON VIEW v4_report.lot_metrics IS 
'Paridad exacta con v3_lot_metrics (000202). rent_total_per_ha NO expuesta pero usada en operating_result.';
