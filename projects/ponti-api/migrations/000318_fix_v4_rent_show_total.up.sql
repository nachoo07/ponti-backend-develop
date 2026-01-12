-- =============================================================================
-- MIGRACIÓN 000318: FASE 2 - Mostrar arriendo TOTAL en toda la app
-- =============================================================================
--
-- PROBLEMA:
-- Las vistas muestran rent_fixed_only (solo parte fija del arriendo)
-- Pero el resultado operativo resta rent_total (fijo + variable)
-- Esto causa inconsistencia: usuario ve "Arriendo: 100" pero se resta 115
--
-- SOLUCIÓN:
-- Cambiar arriendo_usd para que muestre rent_total
-- Así el usuario ve el mismo valor que se resta en resultado operativo
--

-- =============================================================================
-- 1. Agregar funciones faltantes a v4_ssot
-- =============================================================================

CREATE OR REPLACE FUNCTION v4_ssot.net_price_usd_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.net_price_usd_for_lot(p_lot_id), 0)::numeric;
$$;

CREATE OR REPLACE FUNCTION v4_ssot.income_net_per_ha_for_lot(p_lot_id bigint)
RETURNS numeric LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.income_net_per_ha_for_lot(p_lot_id), 0)::numeric;
$$;

-- =============================================================================
-- 2. Recrear v4_calc.lot_base_costs con rent_total como mostrado
-- =============================================================================

DROP VIEW IF EXISTS v4_report.lot_list CASCADE;
DROP VIEW IF EXISTS v4_report.lot_metrics CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_income CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_costs CASCADE;

CREATE OR REPLACE VIEW v4_calc.lot_base_costs AS
WITH raw AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    l.id AS lot_id,
    l.name AS lot_name,
    l.hectares,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
ssot AS (
  SELECT
    r.project_id, r.field_id, r.current_crop_id, r.lot_id, r.lot_name,
    r.hectares,
    r.tons,
    v4_ssot.seeded_area_for_lot(r.lot_id) AS seeded_area_ha,
    v4_ssot.labor_cost_for_lot(r.lot_id) AS labor_cost_usd,
    v4_ssot.supply_cost_for_lot(r.lot_id) AS supply_cost_usd,
    -- FIX 000318: rent_per_ha = TOTAL (para mostrar)
    v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_per_ha_usd,
    -- rent_fixed para Total Invertido
    v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_fixed_per_ha_usd,
    v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd,
    v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha
  FROM raw r
),
derived AS (
  SELECT
    s.project_id, s.field_id, s.current_crop_id, s.lot_id, s.lot_name,
    s.hectares,
    s.tons,
    s.seeded_area_ha,
    s.yield_tn_per_ha,
    COALESCE(s.labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(s.supply_cost_usd, 0)::numeric AS supply_cost_usd,
    COALESCE(s.rent_per_ha_usd, 0)::numeric AS rent_per_ha_usd,
    COALESCE(s.rent_fixed_per_ha_usd, 0)::numeric AS rent_fixed_per_ha_usd,
    COALESCE(s.admin_cost_per_ha_usd, 0)::numeric AS admin_cost_per_ha_usd
  FROM ssot s
)
SELECT
  d.project_id, d.field_id, d.current_crop_id, d.lot_id, d.lot_name,
  d.hectares,
  d.tons,
  d.seeded_area_ha,
  d.yield_tn_per_ha,
  d.labor_cost_usd,
  d.supply_cost_usd,
  v4_core.per_ha(d.labor_cost_usd + d.supply_cost_usd, d.hectares::numeric) AS direct_cost_per_ha_usd,
  d.rent_per_ha_usd,
  d.rent_fixed_per_ha_usd,
  d.admin_cost_per_ha_usd
FROM derived d;

COMMENT ON VIEW v4_calc.lot_base_costs IS 
'FIX 000318: rent_per_ha_usd = TOTAL. rent_fixed_per_ha_usd = capitalizable.';

-- =============================================================================
-- 3. Recrear v4_calc.lot_base_income
-- =============================================================================

CREATE OR REPLACE VIEW v4_calc.lot_base_income AS
SELECT
  c.project_id, c.field_id, c.current_crop_id, c.lot_id, c.lot_name,
  c.hectares,
  c.tons,
  c.seeded_area_ha,
  c.yield_tn_per_ha,
  c.labor_cost_usd,
  c.supply_cost_usd,
  c.direct_cost_per_ha_usd,
  c.rent_per_ha_usd,
  c.rent_fixed_per_ha_usd,
  c.admin_cost_per_ha_usd,
  v4_ssot.net_price_usd_for_lot(c.lot_id) AS net_price_usd,
  v4_ssot.income_net_per_ha_for_lot(c.lot_id) AS income_net_per_ha_usd
FROM v4_calc.lot_base_costs c;

-- =============================================================================
-- 4. Recrear v4_report.lot_metrics
-- =============================================================================

CREATE OR REPLACE VIEW v4_report.lot_metrics AS
SELECT
  c.project_id,
  c.field_id,
  c.current_crop_id,
  c.lot_id,
  c.lot_name,
  c.hectares,
  c.tons,
  c.seeded_area_ha,
  c.yield_tn_per_ha,
  c.labor_cost_usd,
  c.supply_cost_usd,
  c.direct_cost_per_ha_usd,
  c.net_price_usd,
  c.income_net_per_ha_usd,
  -- FIX 000318: rent_per_ha_usd = TOTAL (mostrado)
  c.rent_per_ha_usd,
  c.admin_cost_per_ha_usd,
  -- active_total usa rent_FIXED (capitalizable)
  (c.direct_cost_per_ha_usd + c.rent_fixed_per_ha_usd + c.admin_cost_per_ha_usd)::numeric AS active_total_per_ha_usd,
  -- operating_result usa rent_TOTAL (consistente)
  (c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd))::numeric AS operating_result_per_ha_usd,
  (c.rent_per_ha_usd * c.hectares)::numeric AS rent_total_usd,
  (c.labor_cost_usd + c.supply_cost_usd)::numeric AS direct_cost_total_usd,
  ((c.direct_cost_per_ha_usd + c.rent_fixed_per_ha_usd + c.admin_cost_per_ha_usd) * c.hectares)::numeric AS active_total_usd,
  ((c.income_net_per_ha_usd - (c.direct_cost_per_ha_usd + c.rent_per_ha_usd + c.admin_cost_per_ha_usd)) * c.hectares)::numeric AS operating_result_total_usd,
  (c.income_net_per_ha_usd * c.hectares)::numeric AS income_net_total_usd
FROM v4_calc.lot_base_income c;

COMMENT ON VIEW v4_report.lot_metrics IS 
'FIX 000318: rent_per_ha_usd = TOTAL. Arriendo mostrado = arriendo restado.';

-- =============================================================================
-- 5. Recrear v4_report.lot_list (simplificado)
-- =============================================================================

CREATE OR REPLACE VIEW v4_report.lot_list AS
SELECT
  l.id AS lot_id,
  l.name AS lot_name,
  l.hectares,
  l.tons,
  f.id AS field_id,
  f.name AS field_name,
  p.id AS project_id,
  p.name AS project_name,
  l.current_crop_id,
  curr_crop.name AS current_crop,
  lm.yield_tn_per_ha,
  lm.net_price_usd,
  lm.income_net_per_ha_usd,
  lm.direct_cost_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  lm.active_total_per_ha_usd,
  lm.operating_result_per_ha_usd,
  lm.income_net_total_usd,
  lm.direct_cost_total_usd,
  lm.rent_total_usd,
  lm.active_total_usd,
  lm.operating_result_total_usd
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
LEFT JOIN v4_report.lot_metrics lm ON lm.lot_id = l.id
WHERE l.deleted_at IS NULL;

-- =============================================================================
-- 6. Recrear v4_report.field_crop_economicos
-- =============================================================================

DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;

CREATE OR REPLACE VIEW v4_report.field_crop_economicos AS
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
    -- FIX 000318: rent = TOTAL
    SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_total_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * sowed_area_ha)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) AS gastos_directos_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd)
  ) AS margen_bruto_usd,
  (
    (v3_core_ssot.safe_div(production_tn, sown_area_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha)
  ) AS margen_bruto_usd_ha,
  -- FIX 000318: arriendo_usd = TOTAL
  rent_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_total_usd, sown_area_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  (
    (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    (labor_costs_usd + supply_costs_usd) - rent_total_usd - administration_usd
  ) AS resultado_operativo_usd,
  (
    (v3_core_ssot.safe_div(production_tn, sown_area_ha) * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) -
    v3_core_ssot.safe_div(rent_total_usd, sown_area_ha) -
    v3_core_ssot.safe_div(administration_usd, sown_area_ha)
  ) AS resultado_operativo_usd_ha
FROM aggregated;

COMMENT ON VIEW v4_report.field_crop_economicos IS 
'FIX 000318: arriendo_usd = TOTAL. Arriendo mostrado = arriendo restado.';

-- =============================================================================
-- 7. Recrear v4_report.field_crop_metrics (nombres español = paridad v3)
-- =============================================================================

CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
SELECT
  c.project_id,
  c.field_id,
  c.field_name,
  c.current_crop_id,
  c.crop_name,
  -- Nombres en español (paridad con modelo Go)
  c.superficie_total AS superficie_ha,
  c.produccion_tn,
  c.superficie_sembrada_ha AS area_sembrada_ha,
  c.area_cosechada_ha,
  c.rendimiento_tn_ha,
  c.precio_bruto_usd_tn,
  c.gasto_flete_usd_tn,
  c.gasto_comercial_usd_tn,
  c.precio_neto_usd_tn,
  c.ingreso_neto_por_ha AS ingreso_neto_usd_ha,
  (c.ingreso_neto_por_ha * c.superficie_sembrada_ha) AS ingreso_neto_usd,
  -- Labores
  COALESCE(lb.total_labores_usd, 0) AS costos_labores_usd,
  COALESCE(lb.total_labores_usd_ha, 0) AS costos_labores_usd_ha,
  -- Insumos
  COALESCE(i.total_insumos_usd, 0) AS costos_insumos_usd,
  COALESCE(i.total_insumos_usd_ha, 0) AS costos_insumos_usd_ha,
  -- Costos directos
  COALESCE(e.gastos_directos_usd, 0) AS total_costos_directos_usd,
  COALESCE(e.gastos_directos_usd_ha, 0) AS costos_directos_usd_ha,
  -- Margen bruto
  COALESCE(e.margen_bruto_usd, 0) AS margen_bruto_usd,
  COALESCE(e.margen_bruto_usd_ha, 0) AS margen_bruto_usd_ha,
  -- Arriendo (FIX 000318: TOTAL)
  COALESCE(e.arriendo_usd, 0) AS arriendo_usd,
  COALESCE(e.arriendo_usd_ha, 0) AS arriendo_usd_ha,
  -- Administración
  COALESCE(e.adm_estructura_usd, 0) AS administracion_usd,
  COALESCE(e.adm_estructura_usd_ha, 0) AS administracion_usd_ha,
  -- Resultado operativo
  COALESCE(e.resultado_operativo_usd, 0) AS resultado_operativo_usd,
  COALESCE(e.resultado_operativo_usd_ha, 0) AS resultado_operativo_usd_ha,
  -- Rentabilidad
  COALESCE(r.total_invertido_usd, 0) AS total_invertido_usd,
  COALESCE(r.total_invertido_usd_ha, 0) AS total_invertido_usd_ha,
  COALESCE(r.renta_pct, 0) AS renta_pct,
  COALESCE(r.rinde_indiferencia_total_usd_tn, 0) AS rinde_indiferencia_usd_tn
FROM v4_report.field_crop_cultivos c
LEFT JOIN v4_report.field_crop_labores lb
  ON lb.project_id = c.project_id AND lb.field_id = c.field_id AND lb.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_insumos i
  ON i.project_id = c.project_id AND i.field_id = c.field_id AND i.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_economicos e
  ON e.project_id = c.project_id AND e.field_id = c.field_id AND e.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_rentabilidad r
  ON r.project_id = c.project_id AND r.field_id = c.field_id AND r.current_crop_id = c.current_crop_id;

COMMENT ON VIEW v4_report.field_crop_metrics IS 
'FIX 000318: Nombres español (paridad v3). arriendo_usd = TOTAL.';
