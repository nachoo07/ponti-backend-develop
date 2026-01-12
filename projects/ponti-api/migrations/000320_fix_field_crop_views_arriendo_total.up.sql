-- =============================================================================
-- Migration: 000321_fix_field_crop_views_arriendo_total
-- Description: Recrea vistas v4_report.field_crop_* con estructura correcta
-- Fix: arriendo_usd ahora muestra TOTAL (no solo FIJO)
-- Basado en migraciones originales 000311-000316
-- =============================================================================

-- 1. DROP en orden inverso (dependencias primero)
DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_rentabilidad CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_insumos CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_labores CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_cultivos CASCADE;

-- =============================================================================
-- 2. FIELD_CROP_CULTIVOS (copia exacta de 000313)
-- =============================================================================
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
  SUM(hectares)::numeric AS superficie_total,
  SUM(sowed_area_ha)::numeric AS superficie_sembrada_ha,
  SUM(harvested_area_ha)::numeric AS area_cosechada_ha,
  SUM(tons)::numeric AS produccion_tn,
  v3_report_ssot.board_price_for_lot(MIN(lot_id)) AS precio_bruto_usd_tn,
  v3_report_ssot.freight_cost_for_lot(MIN(lot_id)) AS gasto_flete_usd_tn,
  v3_report_ssot.commercial_cost_for_lot(MIN(lot_id)) AS gasto_comercial_usd_tn,
  v3_lot_ssot.net_price_usd_for_lot(MIN(lot_id)) AS precio_neto_usd_tn,
  v3_core_ssot.safe_div(SUM(tons), SUM(hectares)::numeric) AS rendimiento_tn_ha,
  (v3_core_ssot.safe_div(SUM(tons), SUM(hectares)::numeric) * 
   v3_lot_ssot.net_price_usd_for_lot(MIN(lot_id))) AS ingreso_neto_por_ha
FROM lot_base
GROUP BY project_id, field_id, field_name, crop_id, crop_name;

-- =============================================================================
-- 3. FIELD_CROP_LABORES (copia exacta de 000314)
-- =============================================================================
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
  COALESCE(SUM(siembra_usd), 0) AS siembra_total_usd,
  COALESCE(SUM(pulverizacion_usd), 0) AS pulverizacion_total_usd,
  COALESCE(SUM(riego_usd), 0) AS riego_total_usd,
  COALESCE(SUM(cosecha_usd), 0) AS cosecha_total_usd,
  COALESCE(SUM(otras_labores_usd), 0) AS otras_labores_total_usd,
  v3_core_ssot.safe_div(COALESCE(SUM(siembra_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS siembra_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(pulverizacion_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS pulverizacion_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(riego_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS riego_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(cosecha_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS cosecha_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otras_labores_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS otras_labores_usd_ha,
  COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
   SUM(cosecha_usd) + SUM(otras_labores_usd), 0) AS total_labores_usd,
  v3_core_ssot.safe_div(
    COALESCE(SUM(siembra_usd) + SUM(pulverizacion_usd) + SUM(riego_usd) + 
     SUM(cosecha_usd) + SUM(otras_labores_usd), 0),
    COALESCE(SUM(sowed_area_ha), 1)
  ) AS total_labores_usd_ha
FROM labor_costs
GROUP BY project_id, field_id, crop_id;

-- =============================================================================
-- 4. FIELD_CROP_INSUMOS (copia exacta de 000315)
-- =============================================================================
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
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla') AS semillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') AS curasemillas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') AS herbicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') AS insecticidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') AS fungicidas_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') AS coadyuvantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fertilizantes') AS fertilizantes_usd,
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos') AS otros_insumos_usd
  FROM lot_base lb
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  COALESCE(SUM(semillas_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_usd), 0) AS otros_insumos_total_usd,
  v3_core_ssot.safe_div(COALESCE(SUM(semillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS semillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(curasemillas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS curasemillas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(herbicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS herbicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(insecticidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS insecticidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fungicidas_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fungicidas_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(coadyuvantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS coadyuvantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(fertilizantes_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS fertilizantes_usd_ha,
  v3_core_ssot.safe_div(COALESCE(SUM(otros_insumos_usd), 0), COALESCE(SUM(sowed_area_ha), 1)) AS otros_insumos_usd_ha,
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

-- =============================================================================
-- 5. FIELD_CROP_ECONOMICOS (FIX: arriendo_usd = TOTAL, no FIJO)
-- =============================================================================
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
    -- Arriendo FIJO (para Total Invertido)
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * sowed_area_ha)::numeric AS rent_fixed_usd,
    -- Arriendo TOTAL (fijo + % ingresos) para mostrar en UI
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
  -- ============================================
  -- FIX 000321: arriendo_usd = TOTAL (no FIJO)
  -- ============================================
  rent_total_usd AS arriendo_usd,
  v3_core_ssot.safe_div(rent_total_usd, sown_area_ha) AS arriendo_usd_ha,
  administration_usd AS adm_estructura_usd,
  v3_core_ssot.safe_div(administration_usd, sown_area_ha) AS adm_estructura_usd_ha,
  -- Resultado Operativo usa arriendo TOTAL (consistente con lo mostrado)
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

-- =============================================================================
-- 6. FIELD_CROP_RENTABILIDAD (copia exacta de 000312, usa rent_fixed para Total Invertido)
-- =============================================================================
CREATE OR REPLACE VIEW v4_report.field_crop_rentabilidad AS
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
    -- Total Invertido usa arriendo FIJO (capitalizable)
    SUM(v3_lot_ssot.rent_fixed_only_for_lot(lot_id) * hectares)::numeric AS rent_usd,
    SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares)::numeric AS administration_usd
  FROM lot_base
  GROUP BY project_id, field_id, crop_id
)
SELECT
  project_id,
  field_id,
  crop_id AS current_crop_id,
  (direct_cost_usd + rent_usd + administration_usd) AS total_invertido_usd,
  v3_core_ssot.safe_div(
    (direct_cost_usd + rent_usd + administration_usd),
    sown_area_ha
  ) AS total_invertido_usd_ha,
  v3_lot_ssot.renta_pct(
    (
      (production_tn * v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)) -
      direct_cost_usd - rent_usd - administration_usd
    ),
    (direct_cost_usd + rent_usd + administration_usd)
  ) AS renta_pct,
  v3_core_ssot.safe_div(
    v3_core_ssot.safe_div((direct_cost_usd + rent_usd + administration_usd), sown_area_ha),
    v3_lot_ssot.net_price_usd_for_lot(sample_lot_id)
  ) AS rinde_indiferencia_total_usd_tn
FROM aggregated;

-- =============================================================================
-- 7. FIELD_CROP_METRICS (copia exacta de 000316)
-- =============================================================================
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
'FIX 000321: arriendo_usd = TOTAL. Basado en migraciones 000311-000316.';
