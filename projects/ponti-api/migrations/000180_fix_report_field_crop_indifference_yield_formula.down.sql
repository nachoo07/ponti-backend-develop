-- ========================================
-- MIGRACIÓN 000180: FIX Report Field Crop Indifference Yield Formula (DOWN)
-- ========================================
--
-- Propósito: Revertir corrección de rinde de indiferencia
-- Fecha: 2025-11-03
-- Autor: Sistema

BEGIN;

-- Restaurar versión 000165 (con fórmula incorrecta)

DROP VIEW IF EXISTS public.v3_report_field_crop_rentabilidad CASCADE;

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
    SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Total Invertido = Gastos Directos + Arriendo + Administración
  (direct_cost_usd + rent_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(
    (direct_cost_usd + rent_usd + administration_usd), 
    sown_area_ha
  ) AS total_invertido_usd_ha,
  
  -- Renta % = Resultado Operativo / Total Invertido * 100
  v3_lot_ssot.renta_pct(
    (((production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - direct_cost_usd) - 
     rent_usd - administration_usd),
    (direct_cost_usd + rent_usd + administration_usd)
  ) AS renta_pct,
  
  -- Rinde Indiferencia (versión 000165: INCORRECTA - calcula PRECIO en lugar de RINDE)
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_usd + administration_usd), sown_area_ha),
    v3_core_ssot.safe_div(production_tn, sown_area_ha)
  ) AS rinde_indiferencia_total_usd_tn

FROM aggregated;

-- Recrear vista principal

DROP VIEW IF EXISTS public.v3_report_field_crop_metrics CASCADE;

CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics AS
SELECT
  c.project_id,
  c.field_id,
  c.field_name,
  c.current_crop_id,
  c.crop_name,
  
  -- De CULTIVOS
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
  v3_core_ssot.safe_div((c.produccion_tn * c.precio_neto_usd_tn), c.superficie_sembrada_ha) AS ingreso_neto_usd_ha,
  
  -- De COSTOS
  c.costos_labores_usd,
  c.costos_labores_usd_ha,
  c.costos_insumos_usd,
  c.costos_insumos_usd_ha,
  c.total_costos_directos_usd,
  c.costos_directos_usd_ha,
  c.margen_bruto_usd,
  c.margen_bruto_usd_ha,
  c.arriendo_usd,
  c.arriendo_usd_ha,
  c.administracion_usd,
  c.administracion_usd_ha,
  c.resultado_operativo_usd,
  c.resultado_operativo_usd_ha,
  
  -- De CALCULADOS
  rent.total_invertido_usd,
  rent.total_invertido_usd_ha,
  rent.renta_pct,
  rent.rinde_indiferencia_total_usd_tn AS rinde_indiferencia_usd_tn

FROM public.v3_report_field_crop_cultivos c
JOIN public.v3_report_field_crop_rentabilidad rent
  ON c.project_id = rent.project_id
  AND c.field_id = rent.field_id
  AND c.current_crop_id = rent.current_crop_id;

COMMIT;

