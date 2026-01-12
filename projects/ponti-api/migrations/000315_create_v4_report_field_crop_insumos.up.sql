-- =============================================================================
-- MIGRACIÓN 000315: v4_report.field_crop_insumos - Paridad con v3
-- =============================================================================
--
-- Fuente: 000165_fix_report_field_crop_add_fertilizers_otros.up.sql (líneas 23-106)
-- FASE 1: Paridad exacta
--

CREATE OR REPLACE VIEW v4_report.field_crop_insumos AS
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
    -- Fertilizantes
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fertilizantes') AS fertilizantes_usd,
    -- Otros Insumos
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos') AS otros_insumos_usd
    
  FROM lot_base lb
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Totales por categoría
  COALESCE(SUM(semillas_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_usd), 0) AS otros_insumos_total_usd,
  
  -- Por hectárea
  v3_core_ssot.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS semillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS curasemillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS herbicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS insecticidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fungicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS coadyuvantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fertilizantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS otros_insumos_usd_ha,
  
  -- Total insumos
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

COMMENT ON VIEW v4_report.field_crop_insumos IS 
'Paridad exacta con v3_report_field_crop_insumos (000165). FASE 1.';
