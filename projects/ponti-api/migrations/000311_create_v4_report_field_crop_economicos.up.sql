-- =============================================================================
-- MIGRACIÓN 000311: v4_report.field_crop_economicos - Paridad con v3
-- =============================================================================
--
-- Propósito: Vista económicos del reporte Campo x Cultivo
-- Fuente: 000205_fix_field_crop_economicos_use_seeded_area.up.sql
--
-- Dependencias:
--   - v3_lot_ssot.* (para paridad exacta)
--   - v3_core_ssot.safe_div
--
-- FASE 1: Paridad exacta con v3_report_field_crop_economicos
--

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
    -- Arriendo FIJO (para mostrar en UI y Total Invertido) usa superficie sembrada
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_fixed_usd,
    -- Arriendo TOTAL (fijo + % ingresos) también sobre superficie sembrada
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
  rent_fixed_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_fixed_usd, sown_area_ha) AS arriendo_usd_ha,
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
'Paridad exacta con v3_report_field_crop_economicos (000205). FASE 1.';
