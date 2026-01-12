-- ========================================
-- MIGRACIÓN 000131: FIX v3_report_field_crop_insumos - Add Fertilizantes y Otros Insumos (UP)
-- ========================================
-- 
-- Propósito: Agregar Fertilizantes y Otros Insumos a la vista de insumos del reporte field-crop
-- Problema: Las categorías "Fertilizantes" y "Otros Insumos" no estaban incluidas en la vista
-- Dependencias: Requiere 000130 (v3_report_field_crop_metrics)
-- Fecha: 2025-10-29
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR VISTA 3: v3_report_field_crop_insumos
-- INSUMOS (Dividir por superficie sembrada) - CORREGIDA
-- ========================================
-- Propósito: Costos de insumos por categoría, divididos por superficie sembrada
-- NUEVO: Incluye Fertilizantes y Otros Insumos
DROP VIEW IF EXISTS public.v3_report_field_crop_insumos CASCADE;

CREATE OR REPLACE VIEW public.v3_report_field_crop_insumos AS
WITH lot_base AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    v3_lot_ssot.seeded_area_for_lot(l.id)::numeric AS sowed_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.current_crop_id IS NOT NULL
),
supply_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    
    -- Semillas
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla') AS semillas_usd,
    
    -- Curasemillas
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') AS curasemillas_usd,
    
    -- Herbicidas
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') AS herbicidas_usd,
    
    -- Insecticidas
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') AS insecticidas_usd,
    
    -- Fungicidas
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') AS fungicidas_usd,
    
    -- Coadyuvantes
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') AS coadyuvantes_usd,
    
    -- NUEVO: Fertilizantes (type_id = 3)
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fertilizantes') AS fertilizantes_usd,
    
    -- NUEVO: Otros Insumos (type_id = 2)
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos') AS otros_insumos_usd
    
  FROM lot_base lb
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Totales por categoría (con COALESCE para evitar NULL)
  COALESCE(SUM(semillas_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_usd), 0) AS otros_insumos_total_usd,
  
  -- Por hectárea (dividir por superficie sembrada)
  v3_core_ssot.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS semillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS curasemillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS herbicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS insecticidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fungicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS coadyuvantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fertilizantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS otros_insumos_usd_ha,
  
  -- Total insumos (ACTUALIZADO: incluye Fertilizantes y Otros Insumos)
  COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + 
   SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
   SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0) AS total_insumos_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(semillas_usd) + SUM(curasemillas_usd) + SUM(herbicidas_usd) + 
     SUM(insecticidas_usd) + SUM(fungicidas_usd) + SUM(coadyuvantes_usd) +
     SUM(fertilizantes_usd) + SUM(otros_insumos_usd), 0),
    COALESCE(SUM(sowed_area_ha), 1)
  ) AS total_insumos_usd_ha

FROM supply_costs
GROUP BY project_id, field_id, crop_id;

-- ========================================
-- RECREAR VISTA 4: v3_report_field_crop_economicos
-- ACTUALIZADA para incluir nuevas categorías en cálculo
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
    -- Labores: usar SSOT (solo workorders, correcto)
    SUM(v3_lot_ssot.labor_cost_for_lot(lot_id))::numeric AS labor_costs_usd,
    -- Insumos: calcular manualmente incluyendo TODAS las categorías (ACTUALIZADO)
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
    SUM(v3_lot_ssot.rent_per_ha_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Gastos Directos = Total Labores + Total Insumos
  (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
  v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha) AS gastos_directos_usd_ha,
  
  -- Margen Bruto = Ingreso Neto - Gastos Directos
  ((production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   (labor_costs_usd + supply_costs_usd)) AS margen_bruto_usd,
  ((v3_core_ssot.safe_div(production_tn, sown_area_ha) * 
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
   v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha)) AS margen_bruto_usd_ha,
  
  -- Arriendo
  rent_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_usd, sown_area_ha) AS arriendo_usd_ha,
  
  -- Administración y Estructura
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  
  -- Resultado Operativo = Margen Bruto - Arriendo - Adm/Estructura
  (((production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
    (labor_costs_usd + supply_costs_usd)) - 
   rent_usd - administration_usd) AS resultado_operativo_usd,
  (((v3_core_ssot.safe_div(production_tn, sown_area_ha) * 
     v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) - 
    v3_core_ssot.safe_div((labor_costs_usd + supply_costs_usd), sown_area_ha)) -
   v3_core_ssot.safe_div(rent_usd, sown_area_ha) -
   v3_core_ssot.safe_div(administration_usd, sown_area_ha)) AS resultado_operativo_usd_ha

FROM aggregated;

-- ========================================
-- RECREAR VISTA 5: v3_report_field_crop_rentabilidad
-- ACTUALIZADA para incluir nuevas categorías en cálculo
-- ========================================
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
    -- Costos directos: incluir TODAS las categorías (ACTUALIZADO)
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
  
  -- Rinde Indiferencia = Total Invertido / Rendimiento
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_usd + administration_usd), sown_area_ha),
    v3_core_ssot.safe_div(production_tn, sown_area_ha)
  ) AS rinde_indiferencia_total_usd_tn

FROM aggregated;

-- ========================================
-- RECREAR VISTA PRINCIPAL (CONSOLIDADA)
-- v3_report_field_crop_metrics
-- NO REQUIERE CAMBIOS - Se actualiza automáticamente por CASCADE
-- ========================================
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

COMMIT;

-- Comentarios finales
COMMENT ON VIEW public.v3_report_field_crop_insumos IS 'Vista 3/5: INSUMOS - ACTUALIZADA con Fertilizantes y Otros Insumos';
COMMENT ON VIEW public.v3_report_field_crop_economicos IS 'Vista 4/5: ECONÓMICOS - ACTUALIZADA para incluir todas categorías de insumos';
COMMENT ON VIEW public.v3_report_field_crop_rentabilidad IS 'Vista 5/5: RENTABILIDAD - ACTUALIZADA para incluir todas categorías de insumos';

