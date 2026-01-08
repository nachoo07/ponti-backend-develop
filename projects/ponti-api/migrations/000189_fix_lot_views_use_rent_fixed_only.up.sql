-- ========================================
-- MIGRATION 000189: FIX Lot Views Use Rent Fixed Only (UP)
-- ========================================
--
-- Purpose: Ajustar vistas de lotes y reportes para usar rent_fixed_only_for_lot()
--          y mantener consistencia con aportes e integridad de datos.
--
-- Date: 2025-11-08
-- Author: System
--
-- Note: Code in English, comentarios en español.

BEGIN;

-- ============================================================================
-- Paso 1: Eliminar vistas dependientes que usan rent_per_ha_for_lot()
-- ============================================================================
DROP VIEW IF EXISTS public.v3_dashboard_metrics CASCADE;
DROP VIEW IF EXISTS public.v3_lot_list CASCADE;
DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

-- ============================================================================
-- Paso 2: Recrear v3_lot_metrics con rent_fixed_only_for_lot()
-- ============================================================================
CREATE VIEW public.v3_lot_metrics AS
WITH base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.name AS lot_name,
    l.hectares,
    l.tons,
    l.sowing_date,
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0)::numeric AS sowed_area_ha,
    COALESCE(SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END), 0)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  GROUP BY f.project_id, l.id, l.name, l.hectares, l.tons, l.sowing_date
),
workorder_costs AS (
  SELECT
    lot_id,
    COALESCE(labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(direct_cost_usd, 0)::numeric AS direct_cost_usd
  FROM public.v3_workorder_metrics
),
project_totals AS (
  SELECT
    b.project_id,
    COALESCE(SUM(b.hectares), 0)::numeric AS total_hectares
  FROM base b
  GROUP BY b.project_id
),
lot_per_ha_values AS (
  SELECT
    b.project_id,
    b.lot_id,
    b.lot_name,
    b.hectares,
    b.sowed_area_ha,
    b.harvested_area_ha,
    v3_lot_ssot.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
    b.tons,
    b.sowing_date,
    COALESCE(wc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(wc.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(wc.direct_cost_usd, 0)::numeric AS direct_cost_usd,
    COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric AS income_net_total_usd,
    v3_core_ssot.per_ha(
      COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
      b.hectares::numeric
    ) AS income_net_per_ha_usd,
    v3_core_ssot.cost_per_ha(
      COALESCE(wc.direct_cost_usd, 0)::numeric,
      COALESCE(b.sowed_area_ha, 0)::numeric
    ) AS direct_cost_per_ha_usd,
    COALESCE(v3_lot_ssot.rent_fixed_only_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
    COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_cost_per_ha_usd,
    COALESCE(pt.total_hectares, 0)::numeric AS project_total_hectares
  FROM base b
  LEFT JOIN workorder_costs wc ON wc.lot_id = b.lot_id
  LEFT JOIN project_totals pt ON pt.project_id = b.project_id
)
SELECT
  project_id,
  lot_id,
  lot_name,
  hectares,
  sowed_area_ha,
  harvested_area_ha,
  yield_tn_per_ha,
  tons,
  sowing_date,
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

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas por lote. FIX 000189: Usa rent_fixed_only_for_lot().';

-- ============================================================================
-- Paso 3: Recrear v3_lot_list
-- ============================================================================
CREATE VIEW public.v3_lot_list AS
SELECT
  f.project_id,
  p.name AS project_name,
  f.id AS field_id,
  f.name AS field_name,
  l.id AS id,
  l.name AS lot_name,
  l.variety,
  l.season,
  l.previous_crop_id,
  prev_crop.name AS previous_crop,
  l.current_crop_id,
  curr_crop.name AS current_crop,
  l.hectares,
  l.updated_at,
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
  lm.operating_result_total_usd,
  l.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  l.tons,
  (
    SELECT MIN(w.date)
    FROM public.workorders w
    JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
    WHERE w.lot_id = l.id AND w.deleted_at IS NULL
  ) AS raw_sowing_date
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
LEFT JOIN public.v3_lot_metrics lm ON lm.lot_id = l.id
WHERE l.deleted_at IS NULL;

COMMENT ON VIEW public.v3_lot_list IS 'Lista de lotes. FIX 000189: Usa rent_fixed_only_for_lot().';

-- ============================================================================
-- Paso 4: Recrear v3_dashboard_metrics
-- ============================================================================
CREATE VIEW public.v3_dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.sowed_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM public.v3_lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  (p.admin_cost * 10)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    (p.admin_cost * 10)::numeric
  ) AS costs_progress_pct,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE((
      SELECT f.lease_type_value * ph.total_hectares
      FROM public.fields f
      WHERE f.project_id = p.id AND f.deleted_at IS NULL
      LIMIT 1
    ), 0)
  )::double precision AS operating_result_total_costs_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE((
        SELECT f.lease_type_value * ph.total_hectares
        FROM public.fields f
        WHERE f.project_id = p.id AND f.deleted_at IS NULL
        LIMIT 1
      ), 0)
    )::double precision
  ) AS operating_result_pct,
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 'Dashboard metrics. FIX 000189: Usa rent_fixed_only_for_lot().';

-- ============================================================================
-- Paso 5: Recrear vistas de reportes agrícolas con arriendo fijo
-- ============================================================================

CREATE OR REPLACE VIEW public.v3_report_field_crop_rentabilidad AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
aggregated AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    MIN(lot_id) AS sample_lot_id,
    SUM(tons)::numeric AS production_tn,
    SUM(sowed_area_ha)::numeric AS sown_area_ha,
    SUM(hectares)::numeric AS surface_ha,
    SUM(
      v3_lot_ssot.labor_cost_for_lot(lot_id) +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
    )::numeric AS direct_cost_usd,
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(
    (direct_cost_usd + rent_usd + administration_usd),
    sown_area_ha
  ) AS total_invertido_usd_ha,
  v3_lot_ssot.renta_pct(
    (
      (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
      direct_cost_usd - rent_usd - administration_usd
    ),
    (direct_cost_usd + rent_usd + administration_usd)
  ) AS renta_pct,
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_usd + administration_usd), sown_area_ha),
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM aggregated;

COMMENT ON VIEW public.v3_report_field_crop_rentabilidad IS 'Vista 5/5: RENTABILIDAD - FIX 000189: Arriendo usa rent_fixed_only_for_lot().';

CREATE OR REPLACE VIEW public.v3_report_field_crop_economicos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
aggregated AS (
  SELECT
    project_id,
    field_id,
    crop_id,
    MIN(lot_id) AS sample_lot_id,
    SUM(tons)::numeric AS production_tn,
    SUM(sowed_area_ha)::numeric AS sown_area_ha,
    SUM(hectares)::numeric AS surface_ha,
    SUM(v3_lot_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
    SUM(
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos')
    )::numeric AS supply_costs_usd,
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div(
    (labor_costs_usd + supply_costs_usd),
    sown_area_ha
  ) AS gastos_directos_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  (
    (v3_core_ssot.safe_div(production_tn, sown_area_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha)
  ) AS margen_bruto_usd_ha,
  rent_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_usd, sown_area_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_usd - administration_usd
  ) AS resultado_operativo_usd,
  (
    (v3_core_ssot.safe_div(production_tn, sown_area_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) -
    v3_core_ssot.safe_div(rent_usd, sown_area_ha) -
    v3_core_ssot.safe_div(administration_usd, sown_area_ha)
  ) AS resultado_operativo_usd_ha
FROM aggregated;

COMMENT ON VIEW public.v3_report_field_crop_economicos IS 'Vista 4/5: ECONÓMICOS - FIX 000189: Arriendo usa rent_fixed_only_for_lot().';

CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics AS
SELECT
  c.project_id,
  c.field_id,
  c.field_name,
  c.current_crop_id,
  c.crop_name,
  c.superficie_total AS superficie_ha,
  c.produccion_tn,
  c.superficie_sembrada_ha AS area_sembrada_ha,
  c.area_cosechada_ha,
  c.rendimiento_tn_ha,
  c.precio_bruto_usd_tn,
  c.gasto_flete_usd_tn,
  c.gasto_comercial_usd_tn,
  c.precio_neto_usd_tn,
  (c.produccion_tn * c.precio_neto_usd_tn) AS ingreso_neto_usd,
  c.ingreso_neto_por_ha AS ingreso_neto_usd_ha,
  COALESCE(l.total_labores_usd, 0) AS costos_labores_usd,
  COALESCE(l.total_labores_usd_ha, 0) AS costos_labores_usd_ha,
  COALESCE(i.total_insumos_usd, 0) AS costos_insumos_usd,
  COALESCE(i.total_insumos_usd_ha, 0) AS costos_insumos_usd_ha,
  (COALESCE(l.total_labores_usd, 0) + COALESCE(i.total_insumos_usd, 0)) AS total_costos_directos_usd,
  (COALESCE(l.total_labores_usd_ha, 0) + COALESCE(i.total_insumos_usd_ha, 0)) AS costos_directos_usd_ha,
  COALESCE(e.margen_bruto_usd, 0) AS margen_bruto_usd,
  COALESCE(e.margen_bruto_usd_ha, 0) AS margen_bruto_usd_ha,
  COALESCE(e.arriendo_usd, 0) AS arriendo_usd,
  COALESCE(e.arriendo_usd_ha, 0) AS arriendo_usd_ha,
  COALESCE(e.adm_estructura_usd, 0) AS administracion_usd,
  COALESCE(e.adm_estructura_usd_ha, 0) AS administracion_usd_ha,
  COALESCE(e.resultado_operativo_usd, 0) AS resultado_operativo_usd,
  COALESCE(e.resultado_operativo_usd_ha, 0) AS resultado_operativo_usd_ha,
  COALESCE(r.total_invertido_usd, 0) AS total_invertido_usd,
  COALESCE(r.total_invertido_usd_ha, 0) AS total_invertido_usd_ha,
  COALESCE(r.renta_pct, 0) AS renta_pct,
  COALESCE(r.rinde_indiferencia_total_usd_tn, 0) AS rinde_indiferencia_usd_tn
FROM public.v3_report_field_crop_cultivos c
LEFT JOIN public.v3_report_field_crop_labores l
  ON l.project_id = c.project_id
 AND l.field_id = c.field_id
 AND l.current_crop_id = c.current_crop_id
LEFT JOIN public.v3_report_field_crop_insumos i
  ON i.project_id = c.project_id
 AND i.field_id = c.field_id
 AND i.current_crop_id = c.current_crop_id
LEFT JOIN public.v3_report_field_crop_economicos e
  ON e.project_id = c.project_id
 AND e.field_id = c.field_id
 AND e.current_crop_id = c.current_crop_id
LEFT JOIN public.v3_report_field_crop_rentabilidad r
  ON r.project_id = c.project_id
 AND r.field_id = c.field_id
 AND r.current_crop_id = c.current_crop_id;

COMMENT ON VIEW public.v3_report_field_crop_metrics IS 'Vista consolidada field-crop. FIX 000189: Consistencia de arriendo con rent_fixed_only_for_lot().';

-- ============================================================================
-- Paso 6: Dashboard Management Balance con arriendo fijo
-- ============================================================================
CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
SELECT
  p.id AS project_id,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (
    v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
    COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)
  ) AS costos_directos_invertidos_usd,
  (
    (
      v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
      COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)
    ) - v3_dashboard_ssot.direct_costs_total_for_project(p.id)
  ) AS costos_directos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (
    v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) -
    COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)
  ) AS semillas_stock_usd,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (
    v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) -
    COALESCE(SUM(
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
      v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
    ), 0)
  ) AS agroquimicos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_ejecutados_usd,
  (
    SELECT COALESCE(SUM(sm.quantity * s.price), 0)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id
    JOIN public.categories c ON s.category_id = c.id
    WHERE sm.project_id = p.id
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.is_entry = TRUE
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND c.type_id = 3
  ) AS fertilizantes_invertidos_usd,
  (
    (
      SELECT COALESCE(SUM(sm.quantity * s.price), 0)
      FROM public.supply_movements sm
      JOIN public.supplies s ON s.id = sm.supply_id
      JOIN public.categories c ON s.category_id = c.id
      WHERE sm.project_id = p.id
        AND sm.deleted_at IS NULL
        AND s.deleted_at IS NULL
        AND sm.is_entry = TRUE
        AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
        AND c.type_id = 3
    ) - COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0)
  ) AS fertilizantes_stock_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0) AS labores_invertidos_usd,
  v3_dashboard_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
  v3_dashboard_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semilla_cost,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS insumos_cost,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_cost
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

COMMENT ON VIEW public.v3_dashboard_management_balance IS 'Balance de gestión. FIX 000189: Usa rent_fixed_only_for_lot() en cálculos de arriendo.';

-- ============================================================================
-- Paso 7: Resumen de resultados con arriendo fijo
-- ============================================================================
CREATE OR REPLACE VIEW public.v3_report_summary_results_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id
      JOIN public.categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
),
by_crop AS (
  SELECT
    lb.project_id,
    lb.current_crop_id,
    lb.crop_name::text AS crop_name,
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS surface_ha,
    COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric AS net_income_usd,
    COALESCE(SUM(
      v3_lot_ssot.labor_cost_for_lot(lb.lot_id)::double precision +
      v3_lot_ssot.supply_cost_for_lot_base(lb.lot_id)::double precision
    ), 0)::numeric AS direct_costs_usd,
    COALESCE(SUM(lb.seeded_area_ha * v3_lot_ssot.rent_fixed_only_for_lot(lb.lot_id)), 0)::numeric AS rent_usd,
    COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric AS structure_usd
  FROM lot_base lb
  WHERE lb.current_crop_id IS NOT NULL
  GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(direct_costs_usd + rent_usd + structure_usd)::numeric AS total_invested_usd,
    SUM(net_income_usd - (direct_costs_usd + rent_usd + structure_usd))::numeric AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd) AS total_invested_usd,
  (bc.net_income_usd - (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd)) AS operating_result_usd,
  v3_calc.renta_pct(
    (bc.net_income_usd - (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd))::double precision,
    (bc.direct_costs_usd + bc.rent_usd + bc.structure_usd)::double precision
  )::numeric AS crop_return_pct,
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_usd AS total_invested_project_usd,
  pt.total_operating_result_usd,
  v3_calc.renta_pct(
    pt.total_operating_result_usd::double precision,
    pt.total_invested_usd::double precision
  )::numeric AS project_return_pct
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id
ORDER BY bc.project_id, bc.current_crop_id;

COMMENT ON VIEW public.v3_report_summary_results_view IS 'Resumen general por cultivo. FIX 000189: Arriendo usa rent_fixed_only_for_lot().';

COMMIT;
