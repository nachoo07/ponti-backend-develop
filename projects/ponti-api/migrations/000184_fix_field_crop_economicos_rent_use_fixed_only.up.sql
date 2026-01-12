-- ========================================
-- MIGRACIÓN 000184: FIX Field Crop Economicos - Use Fixed Rent Only (UP)
-- ========================================
--
-- Propósito: Corregir v3_report_field_crop_economicos para usar arriendo fijo solamente
-- Problema: La migración 000183 corrigió v3_report_field_crop_rentabilidad pero
--           v3_report_field_crop_economicos seguía usando rent_per_ha_for_lot()
--           Esto causa que arriendo_usd en la vista principal sea incorrecto
-- Solución: Cambiar rent_per_ha_for_lot() a rent_fixed_only_for_lot() en economicos
-- Fecha: 2025-11-08
-- Autor: Sistema
--
-- Impacto: arriendo_usd en v3_report_field_crop_metrics será consistente
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR: v3_report_field_crop_economicos
-- SOLO cambio: rent_per_ha_for_lot() → rent_fixed_only_for_lot()
-- ========================================

DROP VIEW IF EXISTS public.v3_report_field_crop_economicos CASCADE;

CREATE OR REPLACE VIEW public.v3_report_field_crop_economicos AS
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
    -- FIX 000184: ARRIENDO SOLO FIJO
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Gastos Directos
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div(
    (labor_costs_usd + supply_costs_usd), 
    sown_area_ha
  ) AS gastos_directos_usd_ha,
  
  -- Margen Bruto
  ((production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   (labor_costs_usd + supply_costs_usd)) AS margen_bruto_usd,
  ((v3_core_ssot.safe_div(production_tn, sown_area_ha) * 
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha)) AS margen_bruto_usd_ha,
  
  -- Arriendo FIJO
  rent_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_usd, sown_area_ha) AS arriendo_usd_ha,
  
  -- Administración
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  
  -- Resultado Operativo
  ((production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   (labor_costs_usd + supply_costs_usd) - rent_usd - administration_usd) AS resultado_operativo_usd,
  ((v3_core_ssot.safe_div(production_tn, sown_area_ha) * 
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) - 
   v3_core_ssot.safe_div(rent_usd, sown_area_ha) - 
   v3_core_ssot.safe_div(administration_usd, sown_area_ha)) AS resultado_operativo_usd_ha

FROM aggregated;

COMMENT ON VIEW public.v3_report_field_crop_economicos IS 
'Vista 4/5: ECONÓMICOS - FIX 000184: Arriendo usa rent_fixed_only_for_lot() consistente con rentabilidad.';

-- ========================================
-- RECREAR: v3_report_field_crop_metrics
-- (Se eliminó por CASCADE al dropear economicos)
-- ========================================

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
  c.ingreso_neto_por_ha AS ingreso_neto_usd_ha,
  
  -- De LABORES e INSUMOS (con COALESCE para manejar NULL de LEFT JOIN)
  COALESCE(l.total_labores_usd, 0) AS costos_labores_usd,
  COALESCE(l.total_labores_usd_ha, 0) AS costos_labores_usd_ha,
  COALESCE(i.total_insumos_usd, 0) AS costos_insumos_usd,
  COALESCE(i.total_insumos_usd_ha, 0) AS costos_insumos_usd_ha,
  (COALESCE(l.total_labores_usd, 0) + COALESCE(i.total_insumos_usd, 0)) AS total_costos_directos_usd,
  (COALESCE(l.total_labores_usd_ha, 0) + COALESCE(i.total_insumos_usd_ha, 0)) AS costos_directos_usd_ha,
  
  -- De ECONÓMICOS (con COALESCE para manejar NULL de LEFT JOIN)
  COALESCE(e.margen_bruto_usd, 0) AS margen_bruto_usd,
  COALESCE(e.margen_bruto_usd_ha, 0) AS margen_bruto_usd_ha,
  COALESCE(e.arriendo_usd, 0) AS arriendo_usd,
  COALESCE(e.arriendo_usd_ha, 0) AS arriendo_usd_ha,
  COALESCE(e.adm_estructura_usd, 0) AS administracion_usd,
  COALESCE(e.adm_estructura_usd_ha, 0) AS administracion_usd_ha,
  COALESCE(e.resultado_operativo_usd, 0) AS resultado_operativo_usd,
  COALESCE(e.resultado_operativo_usd_ha, 0) AS resultado_operativo_usd_ha,
  
  -- De RENTABILIDAD (con COALESCE para manejar NULL de LEFT JOIN)
  COALESCE(r.total_invertido_usd, 0) AS total_invertido_usd,
  COALESCE(r.total_invertido_usd_ha, 0) AS total_invertido_usd_ha,
  COALESCE(r.renta_pct, 0) AS renta_pct,
  COALESCE(r.rinde_indiferencia_total_usd_tn, 0) AS rinde_indiferencia_usd_tn

FROM v3_report_field_crop_cultivos c
LEFT JOIN v3_report_field_crop_labores l 
  ON l.project_id = c.project_id 
 AND l.field_id = c.field_id 
 AND l.current_crop_id = c.current_crop_id
LEFT JOIN v3_report_field_crop_insumos i 
  ON i.project_id = c.project_id 
 AND i.field_id = c.field_id 
 AND i.current_crop_id = c.current_crop_id
LEFT JOIN v3_report_field_crop_economicos e 
  ON e.project_id = c.project_id 
 AND e.field_id = c.field_id 
 AND e.current_crop_id = c.current_crop_id
LEFT JOIN v3_report_field_crop_rentabilidad r 
  ON r.project_id = c.project_id 
 AND r.field_id = c.field_id 
 AND r.current_crop_id = c.current_crop_id;

COMMENT ON VIEW public.v3_report_field_crop_metrics IS 
'Vista principal consolidada para reporte field-crop. FIX 000184: Arriendo consistente en todas las vistas.';

COMMIT;

