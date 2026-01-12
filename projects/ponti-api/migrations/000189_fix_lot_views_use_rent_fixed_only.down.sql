-- ========================================
-- MIGRATION 000189: FIX Lot Views Use Rent Fixed Only (DOWN)
-- ========================================
-- 
-- Purpose: Revertir cambio en v3_lot_metrics - volver a usar rent_per_ha_for_lot()
-- Date: 2025-11-08
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

DROP VIEW IF EXISTS public.v3_lot_list;
DROP VIEW IF EXISTS public.v3_lot_metrics;

-- Recrear con rent_per_ha_for_lot() (versión original)
CREATE VIEW public.v3_lot_metrics AS
WITH base AS (
  SELECT
    l.id AS lot_id,
    l.field_id,
    f.name AS field_name,
    f.project_id,
    p.name AS project_name,
    l.crop_id,
    cr.name AS crop_name,
    l.name AS lot_name,
    l.hectares,
    l.variety,
    l.sowing_date,
    c.tons AS commercialized_tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  JOIN public.crops cr ON cr.id = l.crop_id AND cr.deleted_at IS NULL
  LEFT JOIN public.commercializations c ON c.lot_id = l.id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
workorder_costs AS (
  SELECT
    w.lot_id,
    SUM(CASE WHEN cat.type_id = 4 THEN lab.price * w.effective_area ELSE 0 END)::numeric AS labor_cost_usd,
    SUM(CASE WHEN cat.type_id IN (1, 2, 3) THEN wi.final_dose * w.effective_area * s.price ELSE 0 END)::numeric AS supplies_cost_usd,
    SUM(
      CASE WHEN cat.type_id = 4 THEN lab.price * w.effective_area ELSE 0 END +
      CASE WHEN cat.type_id IN (1, 2, 3) THEN wi.final_dose * w.effective_area * s.price ELSE 0 END
    )::numeric AS direct_cost_usd,
    SUM(w.effective_area)::numeric AS sowed_area_ha
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON lab.category_id = cat.id
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
  GROUP BY w.lot_id
),
project_total_direct_cost AS (
  SELECT
    b.project_id,
    SUM(b.hectares)::numeric AS total_hectares
  FROM base b
  GROUP BY b.project_id
),
lot_per_ha_values AS (
  SELECT
    b.lot_id,
    b.field_id,
    b.field_name,
    b.project_id,
    b.project_name,
    b.crop_id,
    b.crop_name,
    b.lot_name,
    b.hectares,
    b.variety,
    b.sowing_date,
    COALESCE(b.commercialized_tons, 0)::numeric AS commercialized_tons,
    COALESCE(wc.sowed_area_ha, 0)::numeric AS sowed_area_ha,
    v3_core_ssot.yield_per_ha(
      COALESCE(b.commercialized_tons, 0)::numeric,
      COALESCE(wc.sowed_area_ha, 0)::numeric
    ) AS yield_tn_per_ha,
    COALESCE(wc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(wc.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(wc.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    v3_lot_ssot.income_net_total_for_lot(b.lot_id) AS income_net_total_usd,
    v3_core_ssot.income_net_per_ha(
      v3_lot_ssot.income_net_total_for_lot(b.lot_id),
      COALESCE(wc.sowed_area_ha, 0)::numeric
    ) AS income_net_per_ha_usd,
    v3_core_ssot.cost_per_ha(
      COALESCE(wc.direct_cost_usd, 0)::numeric,
      COALESCE(b.sowed_area_ha, 0)::numeric
    ) AS direct_cost_per_ha_usd,
    -- ROLLBACK: Volver a usar rent_per_ha_for_lot()
    COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
    COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_cost_per_ha_usd,
    COALESCE(ptdc.total_hectares, 0)::numeric AS project_total_hectares
  FROM base b
  LEFT JOIN workorder_costs wc ON wc.lot_id = b.lot_id
  LEFT JOIN project_total_direct_cost ptdc ON ptdc.project_id = b.project_id
)
SELECT
  lot_id,
  field_id,
  field_name,
  project_id,
  project_name,
  crop_id,
  crop_name,
  lot_name,
  hectares,
  variety,
  sowing_date,
  commercialized_tons,
  sowed_area_ha,
  0::numeric AS harvested_area_ha,
  yield_tn_per_ha,
  labor_cost_usd,
  supplies_cost_usd,
  direct_cost_usd,
  income_net_total_usd,
  income_net_per_ha_usd,
  rent_per_ha_usd,
  admin_cost_per_ha_usd,
  (direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  (income_net_per_ha_usd - (direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  (rent_per_ha_usd * hectares)::numeric AS rent_total_usd,
  (admin_cost_per_ha_usd * hectares)::numeric AS admin_total_usd,
  ((direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd) * hectares)::numeric AS active_total_usd,
  ((income_net_per_ha_usd - (direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd)) * hectares)::numeric AS operating_result_total_usd,
  direct_cost_per_ha_usd,
  project_total_hectares
FROM lot_per_ha_values;

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas por lote - operating_result = income_net_per_ha - (direct_cost_per_ha + rent_per_ha + admin_per_ha)';

CREATE VIEW public.v3_lot_list AS
SELECT
  lm.lot_id,
  lm.field_id,
  lm.field_name,
  lm.project_id,
  lm.project_name,
  lm.crop_id,
  lm.crop_name,
  lm.lot_name,
  lm.hectares,
  lm.variety,
  lm.sowing_date,
  lm.commercialized_tons,
  lm.sowed_area_ha,
  lm.harvested_area_ha,
  lm.yield_tn_per_ha,
  lm.direct_cost_per_ha_usd::numeric AS cost_usd_per_ha,
  lm.income_net_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  lm.active_total_per_ha_usd,
  lm.operating_result_per_ha_usd,
  lm.income_net_total_usd,
  lm.direct_cost_usd AS direct_cost_total_usd,
  lm.rent_total_usd,
  lm.admin_total_usd,
  lm.active_total_usd,
  lm.operating_result_total_usd
FROM public.v3_lot_metrics lm;

COMMENT ON VIEW public.v3_lot_list IS 'Lista de lotes con métricas';

COMMIT;

