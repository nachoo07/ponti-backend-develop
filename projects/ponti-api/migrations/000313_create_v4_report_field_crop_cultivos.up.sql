-- =============================================================================
-- MIGRACIÓN 000313: v4_report.field_crop_cultivos - Paridad con v3
-- =============================================================================
--
-- Fuente: 000130_create_v3_report_field_crop_metrics.up.sql (líneas 32-80)
-- FASE 1: Paridad exacta
--

CREATE OR REPLACE VIEW v4_report.field_crop_cultivos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id AS crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha,
    v3_lot_ssot.harvested_area_for_lot(l.id)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
)
SELECT
  project_id,
  field_id,
  field_name,
  crop_id AS current_crop_id,
  crop_name,
  
  -- Superficies
  SUM(hectares)::numeric AS superficie_total,
  SUM(sowed_area_ha)::numeric AS superficie_sembrada_ha,
  SUM(harvested_area_ha)::numeric AS area_cosechada_ha,
  
  -- Producción
  SUM(tons)::numeric AS produccion_tn,
  
  -- Precios (del primer lote, son iguales por crop)
  v3_report_ssot.board_price_for_lot(MIN(lot_id)) AS precio_bruto_usd_tn,
  v3_report_ssot.freight_cost_for_lot(MIN(lot_id)) AS gasto_flete_usd_tn,
  v3_report_ssot.commercial_cost_for_lot(MIN(lot_id)) AS gasto_comercial_usd_tn,
  -- Precio Neto usando función SSOT
  v3_lot_ssot.net_price_usd_for_lot(MIN(lot_id)) AS precio_neto_usd_tn,
  
  -- Rendimiento (Producción Total / Superficie Total - datos directos de lots)
  v3_core_ssot.safe_div(SUM(tons), SUM(hectares)::numeric) AS rendimiento_tn_ha,
  
  -- Ingreso Neto (Rendimiento * Precio Neto = USD/ha) usando SSOT
  (v3_core_ssot.safe_div(SUM(tons), SUM(hectares)::numeric) * 
   v3_lot_ssot.net_price_usd_for_lot(MIN(lot_id))) AS ingreso_neto_por_ha

FROM lot_base
GROUP BY project_id, field_id, field_name, crop_id, crop_name;

COMMENT ON VIEW v4_report.field_crop_cultivos IS 
'Paridad exacta con v3_report_field_crop_cultivos (000130). FASE 1.';
