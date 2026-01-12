-- =============================================================================
-- MIGRACIÓN 000314: v4_report.field_crop_labores - Paridad con v3
-- =============================================================================
--
-- Fuente: 000130_create_v3_report_field_crop_metrics.up.sql (líneas 87-208)
-- FASE 1: Paridad exacta
--

CREATE OR REPLACE VIEW v4_report.field_crop_labores AS
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
labor_costs AS (
  SELECT
    lb.project_id,
    lb.field_id,
    lb.crop_id,
    lb.lot_id,
    lb.sowed_area_ha,
    
    -- Siembra
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id 
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS siembra_usd,
    
    -- Pulverización
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id 
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Pulverización'
        AND cat.type_id = 4
    ), 0)::numeric AS pulverizacion_usd,
    
    -- Riego
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id 
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Riego'
        AND cat.type_id = 4
    ), 0)::numeric AS riego_usd,
    
    -- Cosecha
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id 
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name = 'Cosecha'
        AND cat.type_id = 4
    ), 0)::numeric AS cosecha_usd,
    
    -- Otras labores
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON lab.id = w.labor_id
      JOIN public.categories cat ON cat.id = lab.category_id
      WHERE w.lot_id = lb.lot_id 
        AND w.deleted_at IS NULL
        AND w.effective_area > 0
        AND lab.price IS NOT NULL
        AND cat.name NOT IN ('Siembra', 'Pulverización', 'Riego', 'Cosecha')
        AND cat.type_id = 4
    ), 0)::numeric AS otras_labores_usd
    
  FROM lot_base lb
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  
  -- Totales por categoría
  COALESCE(SUM(siembra_usd), 0) AS siembra_total_usd,
  COALESCE(SUM(pulverizacion_usd), 0) AS pulverizacion_total_usd,
  COALESCE(SUM(riego_usd), 0) AS riego_total_usd,
  COALESCE(SUM(cosecha_usd), 0) AS cosecha_total_usd,
  COALESCE(SUM(otras_labores_usd), 0) AS otras_labores_total_usd,
  
  -- Por hectárea
  v3_core_ssot.safe_div(COALESCE(SUM(siembra_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS siembra_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(pulverizacion_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS pulverizacion_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(riego_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS riego_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(cosecha_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS cosecha_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otras_labores_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS otras_labores_usd_ha,
  
  -- Total labores
  COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
   SUM(cosecha_usd) + SUM(otras_labores_usd), 0) AS total_labores_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
     SUM(cosecha_usd) + SUM(otras_labores_usd), 0),
    COALESCE(SUM(sowed_area_ha), 1)
  ) AS total_labores_usd_ha

FROM labor_costs
GROUP BY project_id, field_id, crop_id;

COMMENT ON VIEW v4_report.field_crop_labores IS 
'Paridad exacta con v3_report_field_crop_labores (000130). FASE 1.';
