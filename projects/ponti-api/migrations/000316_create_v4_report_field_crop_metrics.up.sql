-- =============================================================================
-- MIGRACIÓN 000316: v4_report.field_crop_metrics - Paridad con v3
-- =============================================================================
--
-- Fuente: 000205_fix_field_crop_economicos_use_seeded_area.up.sql (líneas 101-153)
-- FASE 1: Paridad exacta - Une todas las vistas field_crop usando v4_report.*
--

CREATE OR REPLACE VIEW v4_report.field_crop_metrics AS
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
FROM v4_report.field_crop_cultivos c
LEFT JOIN v4_report.field_crop_labores l
  ON l.project_id = c.project_id
  AND l.field_id = c.field_id
  AND l.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_insumos i
  ON i.project_id = c.project_id
  AND i.field_id = c.field_id
  AND i.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_economicos e
  ON e.project_id = c.project_id
  AND e.field_id = c.field_id
  AND e.current_crop_id = c.current_crop_id
LEFT JOIN v4_report.field_crop_rentabilidad r
  ON r.project_id = c.project_id
  AND r.field_id = c.field_id
  AND r.current_crop_id = c.current_crop_id;

COMMENT ON VIEW v4_report.field_crop_metrics IS 
'Paridad exacta con v3_report_field_crop_metrics (000205). FASE 1: Une vistas v4_report.field_crop_*.';
