-- =============================================================================
-- MIGRACIÓN 000303: v4_calc.lot_base_costs - Backbone de cálculos por lote
-- =============================================================================
--
-- Propósito: Vista base con 1 fila por lot_id, paridad con v3_lot_metrics (000202)
-- Fecha: 2025-01-XX
-- Autor: Sistema
--

CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH 
raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    COALESCE(l.hectares, 0) AS hectares,  -- double precision (paridad con v3)
    COALESCE(l.tons, 0) AS tons,          -- double precision (paridad con v3)
    l.sowing_date
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
areas AS (
  SELECT
    r.lot_id,
    v4_ssot.seeded_area_for_lot(r.lot_id) AS sowed_area_ha,
    v4_ssot.harvested_area_for_lot(r.lot_id) AS harvested_area_ha
  FROM raw r
),
-- BLINDADO: garantizar 1 fila por lot_id con GROUP BY + MAX
costs AS (
  SELECT
    lot_id,
    MAX(COALESCE(labor_cost_usd, 0))::numeric AS labor_cost_usd,
    MAX(COALESCE(supplies_cost_usd, 0))::numeric AS supplies_cost_usd,
    MAX(COALESCE(direct_cost_usd, 0))::numeric AS direct_cost_usd
  FROM public.v3_workorder_metrics
  GROUP BY lot_id
),
ssot_values AS (
  SELECT
    r.lot_id,
    v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha,
    v4_ssot.income_net_total_for_lot(r.lot_id) AS income_net_total_usd,
    -- rent_per_ha = FIJO (para exponer en UI y active_total)
    v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_per_ha_usd,
    -- rent_total_per_ha = FIJO + % (para operating_result, NO expuesta)
    v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_total_per_ha_usd,
    v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd
  FROM raw r
),
derived AS (
  SELECT
    r.project_id, r.field_id, r.current_crop_id, r.lot_id, r.lot_name,
    r.hectares, r.tons, r.sowing_date,
    COALESCE(a.sowed_area_ha, 0)::numeric AS sowed_area_ha,
    COALESCE(a.harvested_area_ha, 0)::numeric AS harvested_area_ha,
    COALESCE(c.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(c.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(c.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    COALESCE(s.yield_tn_per_ha, 0) AS yield_tn_per_ha,  -- double precision (paridad con v3)
    COALESCE(s.income_net_total_usd, 0)::numeric AS income_net_total_usd,
    COALESCE(s.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
    COALESCE(s.rent_total_per_ha_usd, 0)::numeric AS rent_total_per_ha_usd,
    COALESCE(s.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd
  FROM raw r
  LEFT JOIN areas a ON a.lot_id = r.lot_id
  LEFT JOIN costs c ON c.lot_id = r.lot_id
  LEFT JOIN ssot_values s ON s.lot_id = r.lot_id
)
SELECT
  d.project_id, d.field_id, d.current_crop_id, d.lot_id, d.lot_name,
  d.hectares, d.sowed_area_ha, d.harvested_area_ha,
  d.yield_tn_per_ha, d.tons, d.sowing_date,
  d.labor_cost_usd, d.supplies_cost_usd, d.direct_cost_usd,
  d.income_net_total_usd,
  -- income_net_per_ha: divide por hectares (000202:72-75, usa v3_core_ssot.per_ha)
  v4_core.per_ha(d.income_net_total_usd, d.hectares::numeric) AS income_net_per_ha_usd,
  -- direct_cost_per_ha: divide por sowed_area (000202:76-79, usa v3_core_ssot.cost_per_ha que es per_ha)
  v4_core.per_ha(d.direct_cost_usd, d.sowed_area_ha) AS direct_cost_per_ha_usd,
  d.rent_per_ha_usd,
  d.rent_total_per_ha_usd,
  d.admin_cost_per_ha_usd
FROM derived d;

COMMENT ON VIEW v4_calc.lot_base_costs IS 
'Backbone cálculos por lote. BLINDADO: costs CTE usa GROUP BY lot_id.';
