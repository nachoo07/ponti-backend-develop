-- ========================================
-- MIGRATION 000091: FIX lot metrics cost per hectare calculation (UP)
-- ========================================
--
-- Purpose: Fix lot_metrics to calculate cost per hectare using total project surface
--          instead of individual lot surface
-- Date: 2025-01-27
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- Corregir el cálculo de costo por hectárea para usar superficie total del proyecto
CREATE OR REPLACE VIEW public.v3_lot_metrics AS
WITH
/* ------------------------------------------------------------------
   CTE: lot_base
   Qué hace (ES):
   - Devuelve una fila por lote con sus IDs (lot/field/project) y
     superficies principales del lote.
   - sowed_area_ha = hectáreas nominales del lote (área sembrada).
   - harvested_area_ha = suma de áreas efectivas SOLO de workorders
     de cosecha (lb.category_id = 13).
------------------------------------------------------------------- */
lot_base AS (
  SELECT
    l.id AS lot_id,
    f.id AS field_id,
    f.project_id,             -- necesario para exponer project_id en la vista
    l.hectares,               -- área nominal del lote (ha)

    -- ############# 1) EN/ES: sowed_area_ha / área sembrada ##############
    -- Fórmula (ES): superficie real sembrada desde workorders de siembra (category_id = 9)
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0)::numeric AS sowed_area_ha,

    -- ############# 2) EN/ES: harvested_area_ha / área cosechada ##############
    -- Fórmula (ES): SUM(w.effective_area) SOLO de workorders de cosecha (labors.category_id = 13)
    COALESCE(
      SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END),
      0
    )::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f
    ON f.id = l.field_id
   AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w
    ON w.lot_id = l.id
   AND w.deleted_at IS NULL
  LEFT JOIN public.labors lb
    ON lb.id = w.labor_id
   AND lb.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  -- IMPORTANTE: incluir project_id en el GROUP BY porque lo seleccionamos arriba
  GROUP BY l.id, f.id, f.project_id, l.hectares
),

/* ------------------------------------------------------------------
   CTE: field_total_area
   Qué hace (ES):
   - Calcula la superficie total del campo: suma de hectáreas de
     todos los lotes activos del mismo campo.
   - Devuelve una fila por field_id con total_hectares.
------------------------------------------------------------------- */
field_total_area AS (
  SELECT 
    f.id AS field_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares
  FROM public.fields f
  LEFT JOIN public.lots l
    ON l.field_id = f.id
   AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.id
),

/* ------------------------------------------------------------------
   CTE: project_total_area
   Qué hace (ES):
   - Calcula la superficie total del proyecto: suma de hectáreas de
     todos los lotes activos del mismo proyecto.
   - Calcula el costo directo total del proyecto: suma de costos directos
     de todos los lotes del proyecto.
   - Devuelve una fila por project_id con total_hectares y total_direct_cost.
------------------------------------------------------------------- */
project_total_area AS (
  SELECT 
    f.project_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares,
    COALESCE(SUM(
      v3_calc.direct_cost_usd(
        COALESCE(v3_calc.labor_cost_for_lot(l.id), 0)::numeric,
        COALESCE(v3_calc.supply_cost_for_lot(l.id), 0)::numeric
      )
    ), 0)::numeric AS total_direct_cost
  FROM public.fields f
  LEFT JOIN public.lots l
    ON l.field_id = f.id
   AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id
),

/* ------------------------------------------------------------------
   CTE: lot_worked_area
   Qué hace (ES):
   - Calcula la superficie REAL trabajada por lote (todas las categorías),
     sumando w.effective_area de todos los workorders del lote.
   - Se usa como denominador para costo directo por ha.
------------------------------------------------------------------- */
lot_worked_area AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(SUM(w.effective_area), 0)::numeric AS worked_hectares
  FROM public.lots l
  LEFT JOIN public.workorders w
    ON w.lot_id = l.id
   AND w.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  GROUP BY l.id
)

-- =======================
-- SELECT FINAL DE LA VISTA
-- =======================
SELECT
  -- Identificación (IDs)
  b.project_id,
  b.field_id,
  b.lot_id,

  -- Área nominal del lote (ha)
  b.hectares,

  -- 1) EN/ES: sowed_area_ha / área sembrada
  b.sowed_area_ha,

  -- 2) EN/ES: harvested_area_ha / área cosechada
  b.harvested_area_ha,

  -- ############# 3) EN/ES: yield_tn_per_ha / rendimiento (tn/ha) ##############
  -- Fórmula típica (ES): toneladas cosechadas / harvested_area_ha
  -- (La implementación exacta y casos borde están dentro de v3_calc.yield_tn_per_ha_for_lot)
  v3_calc.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,

  -- Costos/ingresos absolutos (USD) - útiles para conciliación y totales
  COALESCE(v3_calc.labor_cost_for_lot(b.lot_id), 0)::numeric       AS labor_cost_usd,          -- costo de labores (USD)
  COALESCE(v3_calc.supply_cost_for_lot(b.lot_id), 0)::numeric      AS supplies_cost_usd,       -- costo de insumos (USD)
  v3_calc.direct_cost_usd(
    COALESCE(v3_calc.labor_cost_for_lot(b.lot_id), 0)::numeric,
    COALESCE(v3_calc.supply_cost_for_lot(b.lot_id), 0)::numeric
  ) AS direct_cost_usd,         -- costo directo total (USD) usando nueva función SSOT
  COALESCE(v3_calc.income_net_total_for_lot(b.lot_id), 0)::numeric AS income_net_total_usd,    -- ingreso neto total (USD)

  -- Ingreso neto por ha (USD/ha)
  COALESCE(v3_calc.income_net_per_ha_for_lot(b.lot_id), 0)::numeric AS income_net_per_ha_usd,

  -- Métricas por ha (USD/ha)
  COALESCE(p.admin_cost, 0)::numeric                                   AS admin_cost_per_ha_usd,          -- costo administrativo por ha
  COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0)::numeric          AS rent_per_ha_usd,                -- renta por ha
  COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0)::numeric  AS active_total_per_ha_usd,        -- activo total por ha
  COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,  -- resultado operativo por ha

  -- Totales derivados (USD) = métrica/ha × hectáreas del lote
  COALESCE(p.admin_cost, 0)::numeric                                                      AS admin_total_usd,
  (COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric             AS rent_total_usd,
  (COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric     AS active_total_usd,
  (COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares)::numeric AS operating_result_total_usd,

  -- ############# 4) EN/ES: direct_cost_per_ha_usd / costo directo por hectárea ##############
  -- Fórmula CORREGIDA: direct_cost_total_del_proyecto / superficie_total_del_proyecto
  -- Cálculo: suma de costos directos de TODOS los lotes del proyecto / superficie total del proyecto
  -- Para cada lote, mostrar el mismo valor: costo directo total del proyecto / superficie total del proyecto
  -- Resultado esperado: $25,775 USD ÷ 300 Ha = $85.91 USD/Ha
  v3_calc.cost_per_ha(
    COALESCE(pta.total_direct_cost, 0)::numeric,
    COALESCE(pta.total_hectares, 0)::numeric
  ) AS direct_cost_per_ha_usd,

  -- ############# 5) EN/ES: superficie_total / área total del proyecto (ha) ##############
  -- Fórmula (ES): suma de hectáreas (l.hectares) de todos los lotes activos del mismo proyecto
  COALESCE(pta.total_hectares, 0)::numeric AS superficie_total

FROM lot_base b
LEFT JOIN field_total_area fta ON fta.field_id = b.field_id
LEFT JOIN project_total_area pta ON pta.project_id = b.project_id
LEFT JOIN lot_worked_area lwa  ON lwa.lot_id   = b.lot_id
LEFT JOIN public.projects p    ON p.id         = b.project_id AND p.deleted_at IS NULL;
