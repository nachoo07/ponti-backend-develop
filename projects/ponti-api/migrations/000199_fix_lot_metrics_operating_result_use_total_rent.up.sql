-- ========================================
-- MIGRACIÓN 000199: FIX Lot Metrics Operating Result - Use Total Rent (UP)
-- ========================================
--
-- Propósito: Corregir cálculo de Resultado Operativo en v3_lot_metrics
-- Problema: La vista usa rent_fixed_only para calcular operating_result_per_ha_usd,
--           pero debe usar rent_total (fijo + variable) para coincidir con
--           v3_report_summary_results_view y v3_report_field_crop_economicos
-- Solución: Calcular rent_total_per_ha_usd separado y usarlo solo para operating_result_per_ha_usd
--           Mantener rent_per_ha_usd como fixed_only para Total Invertido y UI
-- Fecha: 2025-11-17
-- Autor: Sistema
--
-- Impacto: Control 11 y 12 pasarán correctamente
--          Lotes coincidirán con reportes en resultado operativo
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR: v3_lot_metrics
-- Cambio: Calcular arriendo TOTAL separado para resultado operativo
-- ========================================

DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

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
    -- Arriendo FIJO (para mostrar en UI y para Total Invertido)
    COALESCE(v3_lot_ssot.rent_fixed_only_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
    -- ========================================
    -- FIX 000199: Arriendo TOTAL (para Resultado Operativo)
    -- ========================================
    -- Incluye fijo + variable (% sobre ingresos)
    COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0)::numeric AS rent_total_per_ha_usd,
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
  -- Arriendo FIJO (para mostrar en UI y para Total Invertido)
  rent_per_ha_usd,
  admin_cost_per_ha_usd,
  -- Total Invertido usa arriendo FIJO
  (direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  -- ========================================
  -- FIX 000199: Resultado Operativo usa Arriendo TOTAL
  -- ========================================
  -- Ingreso Neto - Costos Directos - Arriendo TOTAL - Administración
  -- Usa rent_total_per_ha_usd (fijo + variable) no rent_per_ha_usd (solo fijo)
  (income_net_per_ha_usd - (direct_cost_per_ha_usd + rent_total_per_ha_usd + admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  -- Totales usando arriendo FIJO (para mostrar en UI)
  (rent_per_ha_usd * hectares)::numeric AS rent_total_usd,
  (admin_cost_per_ha_usd * hectares)::numeric AS admin_total_usd,
  -- Total Invertido usa arriendo FIJO
  ((direct_cost_per_ha_usd + rent_per_ha_usd + admin_cost_per_ha_usd) * hectares)::numeric AS active_total_usd,
  -- Resultado Operativo Total usa arriendo TOTAL
  ((income_net_per_ha_usd - (direct_cost_per_ha_usd + rent_total_per_ha_usd + admin_cost_per_ha_usd)) * hectares)::numeric AS operating_result_total_usd,
  direct_cost_per_ha_usd,
  project_total_hectares
FROM lot_per_ha_values;

COMMENT ON VIEW public.v3_lot_metrics IS 
'Métricas por lote. FIX 000199: rent_per_ha_usd usa fixed_only (para Total Invertido), operating_result_per_ha_usd usa total (fijo + variable).';

-- ========================================
-- RECREAR: v3_lot_list
-- (Se eliminó por CASCADE al dropear metrics)
-- ========================================

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

COMMENT ON VIEW public.v3_lot_list IS 
'Lista de lotes. FIX 000199: rent_per_ha_usd usa fixed_only (para Total Invertido), operating_result_per_ha_usd usa total (fijo + variable).';

-- ========================================
-- RECREAR: v3_dashboard_metrics
-- (Se eliminó por CASCADE al dropear v3_lot_metrics)
-- ========================================

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
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
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
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 
'Dashboard metrics. FIX 000199: Recreada después de CASCADE, mantiene planned_cost (000196) y rent_fixed_only().';

COMMIT;

