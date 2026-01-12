-- =============================================================================
-- Migration: 000322_create_v4_report_summary_results
-- Description: Vista de resumen de resultados que AGREGA desde field_crop_metrics
-- Principio: NO recalcular, solo agregar valores ya calculados (SSOT)
-- =============================================================================
--
-- ANTES (v3): Recalculaba arriendo, costos, etc. desde lotes (bugs de inconsistencia)
-- AHORA (v4): Agrega desde field_crop_metrics (valores ya corregidos)
--
-- Beneficios:
--   - Arriendo consistente (siempre el configurado)
--   - Renta correcta (numerador y denominador usan mismo arriendo)
--   - Sin duplicación de cálculos
--

CREATE OR REPLACE VIEW v4_report.summary_results AS
WITH 
-- Agregar métricas por proyecto y cultivo desde field_crop_metrics
by_crop AS (
  SELECT
    project_id,
    current_crop_id,
    crop_name,
    
    -- Superficie: suma de áreas sembradas
    SUM(area_sembrada_ha)::numeric AS surface_ha,
    
    -- Totales = SUM(valor_total) ya calculado en field_crop
    -- NO multiplicar por área de nuevo, usar los totales directamente
    SUM(ingreso_neto_usd)::numeric AS net_income_usd,
    SUM(total_costos_directos_usd)::numeric AS direct_costs_usd,
    SUM(arriendo_usd)::numeric AS rent_usd,
    SUM(administracion_usd)::numeric AS structure_usd,
    SUM(total_invertido_usd)::numeric AS total_invested_usd,
    SUM(resultado_operativo_usd)::numeric AS operating_result_usd
    
  FROM v4_report.field_crop_metrics
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
-- Totales por proyecto (para columnas total_*)
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_project_usd,
    SUM(operating_result_usd)::numeric AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  
  -- Métricas por cultivo
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  bc.total_invested_usd,
  bc.operating_result_usd,
  
  -- Renta del cultivo: resultado / total invertido (valores consistentes)
  CASE 
    WHEN bc.total_invested_usd > 0 
    THEN (bc.operating_result_usd / bc.total_invested_usd * 100)::numeric
    ELSE 0::numeric 
  END AS crop_return_pct,
  
  -- Totales del proyecto (para comparación en UI)
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_project_usd,
  pt.total_operating_result_usd,
  
  -- Renta del proyecto
  CASE 
    WHEN pt.total_invested_project_usd > 0 
    THEN (pt.total_operating_result_usd / pt.total_invested_project_usd * 100)::numeric
    ELSE 0::numeric 
  END AS project_return_pct

FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id
ORDER BY bc.project_id, bc.current_crop_id;

COMMENT ON VIEW v4_report.summary_results IS 
'Resumen de resultados por cultivo. AGREGA desde field_crop_metrics (SSOT). NO recalcula.';
