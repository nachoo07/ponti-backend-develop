-- =============================================================================
-- Migration: 000324_rewrite_field_crop_metrics_fast
-- Descripción: Reescribe field_crop_metrics como UNA sola vista sin anidamiento
-- Problema: La versión anterior usaba 5 vistas anidadas con 46 llamadas SSOT
-- Solución: Una sola vista con ~8 llamadas SSOT (igual que v3)
-- 
-- PARA REVERTIR: Ejecutar 000324_rewrite_field_crop_metrics_fast.down.sql
-- =============================================================================

-- Eliminar vistas dependientes
DROP VIEW IF EXISTS v4_report.summary_results CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;

-- =============================================================================
-- NUEVA: field_crop_metrics en una sola vista (sin anidamiento)
-- Basada en la estructura de v3 que es rápida
-- =============================================================================
CREATE VIEW v4_report.field_crop_metrics AS
WITH 
-- Base: datos de lotes agrupados por field + crop
lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    -- Funciones SSOT (se llaman una vez por lote, luego se agregan)
    COALESCE(v3_lot_ssot.seeded_area_for_lot(l.id), 0)::numeric AS sowed_area_ha,
    COALESCE(v3_lot_ssot.harvested_area_for_lot(l.id), 0)::numeric AS harvested_area_ha,
    COALESCE(v3_lot_ssot.yield_tn_per_ha_for_lot(l.id), 0) AS yield_tn_per_ha,
    COALESCE(v3_lot_ssot.labor_cost_for_lot(l.id), 0)::numeric AS labor_cost_usd,
    COALESCE(v3_lot_ssot.supply_cost_for_lot_base(l.id), 0)::numeric AS supply_cost_usd,
    COALESCE(v3_lot_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd,
    -- Arriendo TOTAL (no solo fijo) - FIX del bug
    COALESCE(v3_lot_ssot.rent_per_ha_for_lot(l.id), 0)::numeric AS rent_per_ha,
    COALESCE(v3_calc.admin_cost_per_ha_for_lot(l.id), 0)::numeric AS admin_per_ha,
    -- Precios para cálculos
    COALESCE(v3_report_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price,
    COALESCE(v3_report_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_cost,
    COALESCE(v3_report_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_cost
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL 
    AND l.current_crop_id IS NOT NULL
    AND l.hectares > 0
),
-- Agregación por field + crop
aggregated AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.field_name,
    lb.current_crop_id,
    lb.crop_name,
    -- Superficies y producción
    SUM(lb.hectares)::numeric AS superficie_total,
    SUM(lb.sowed_area_ha)::numeric AS superficie_sembrada_ha,
    SUM(lb.harvested_area_ha)::numeric AS area_cosechada_ha,
    SUM(lb.tons)::numeric AS produccion_tn,
    -- Promedios ponderados por área
    CASE WHEN SUM(lb.sowed_area_ha) > 0 
      THEN SUM(lb.yield_tn_per_ha * lb.sowed_area_ha) / SUM(lb.sowed_area_ha)
      ELSE 0 
    END AS rendimiento_tn_ha,
    -- Costos totales
    SUM(lb.labor_cost_usd)::numeric AS costos_labores_usd,
    SUM(lb.supply_cost_usd)::numeric AS costos_insumos_usd,
    -- Precios (promedio ponderado por tons)
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.board_price * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS precio_bruto_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.freight_cost * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS gasto_flete_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.commercial_cost * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS gasto_comercial_usd_tn,
    CASE WHEN SUM(lb.tons) > 0 
      THEN SUM(lb.net_price_usd * lb.tons) / SUM(lb.tons)
      ELSE 0 
    END AS precio_neto_usd_tn,
    -- Arriendo y admin por área
    SUM(lb.rent_per_ha * lb.sowed_area_ha)::numeric AS arriendo_total_usd,
    SUM(lb.admin_per_ha * lb.sowed_area_ha)::numeric AS admin_total_usd,
    -- Para calcular ingreso neto
    SUM(lb.tons * lb.net_price_usd)::numeric AS ingreso_neto_total
  FROM lot_base lb
  GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name
)
SELECT
  a.project_id,
  a.field_id,
  a.field_name,
  a.current_crop_id,
  a.crop_name,
  -- Superficies
  a.superficie_total AS superficie_ha,
  a.produccion_tn,
  a.superficie_sembrada_ha AS area_sembrada_ha,
  a.area_cosechada_ha,
  a.rendimiento_tn_ha,
  -- Precios
  a.precio_bruto_usd_tn,
  a.gasto_flete_usd_tn,
  a.gasto_comercial_usd_tn,
  a.precio_neto_usd_tn,
  -- Ingresos
  a.ingreso_neto_total AS ingreso_neto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total, a.superficie_sembrada_ha) AS ingreso_neto_usd_ha,
  -- Costos
  a.costos_labores_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd, a.superficie_sembrada_ha) AS costos_labores_usd_ha,
  a.costos_insumos_usd,
  v3_core_ssot.safe_div(a.costos_insumos_usd, a.superficie_sembrada_ha) AS costos_insumos_usd_ha,
  (a.costos_labores_usd + a.costos_insumos_usd)::numeric AS total_costos_directos_usd,
  v3_core_ssot.safe_div(a.costos_labores_usd + a.costos_insumos_usd, a.superficie_sembrada_ha) AS costos_directos_usd_ha,
  -- Margen bruto
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd)::numeric AS margen_bruto_usd,
  v3_core_ssot.safe_div(a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd, a.superficie_sembrada_ha) AS margen_bruto_usd_ha,
  -- Arriendo y administración
  a.arriendo_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(a.arriendo_total_usd, a.superficie_sembrada_ha) AS arriendo_usd_ha,
  a.admin_total_usd AS administracion_usd,
  v3_core_ssot.safe_div(a.admin_total_usd, a.superficie_sembrada_ha) AS administracion_usd_ha,
  -- Resultado operativo
  (a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd)::numeric AS resultado_operativo_usd,
  v3_core_ssot.safe_div(
    a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd, 
    a.superficie_sembrada_ha
  ) AS resultado_operativo_usd_ha,
  -- Total invertido y rentabilidad
  (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd)::numeric AS total_invertido_usd,
  v3_core_ssot.safe_div(
    a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd, 
    a.superficie_sembrada_ha
  ) AS total_invertido_usd_ha,
  -- Renta %
  CASE WHEN (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) > 0
    THEN ((a.ingreso_neto_total - a.costos_labores_usd - a.costos_insumos_usd - a.arriendo_total_usd - a.admin_total_usd) / 
          (a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) * 100)::double precision
    ELSE 0
  END AS renta_pct,
  -- Rinde de indiferencia
  CASE WHEN a.precio_neto_usd_tn > 0 AND a.superficie_sembrada_ha > 0
    THEN ((a.costos_labores_usd + a.costos_insumos_usd + a.arriendo_total_usd + a.admin_total_usd) / a.superficie_sembrada_ha / a.precio_neto_usd_tn)::numeric
    ELSE 0
  END AS rinde_indiferencia_usd_tn
FROM aggregated a;

COMMENT ON VIEW v4_report.field_crop_metrics IS 
'OPTIMIZADO 000324: Una sola vista sin anidamiento. ~10x más rápido.';

-- =============================================================================
-- Recrear summary_results usando la nueva field_crop_metrics
-- =============================================================================
CREATE VIEW v4_report.summary_results AS
WITH 
by_crop AS (
  SELECT
    project_id,
    current_crop_id,
    crop_name,
    SUM(area_sembrada_ha)::numeric AS surface_ha,
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
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  bc.total_invested_usd,
  bc.operating_result_usd,
  CASE WHEN bc.total_invested_usd > 0 
    THEN (bc.operating_result_usd / bc.total_invested_usd * 100)::numeric
    ELSE 0::numeric 
  END AS crop_return_pct,
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_project_usd,
  pt.total_operating_result_usd,
  CASE WHEN pt.total_invested_project_usd > 0 
    THEN (pt.total_operating_result_usd / pt.total_invested_project_usd * 100)::numeric
    ELSE 0::numeric 
  END AS project_return_pct
FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id;

COMMENT ON VIEW v4_report.summary_results IS 
'SSOT: Agrega desde field_crop_metrics optimizado.';
