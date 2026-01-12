-- ========================================
-- MIGRATION 000078: CREATE v3_workorder_views (UP)
-- ========================================
-- 
-- Purpose: Create workorder metrics and list views
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_workorder_metrics: métricas agregadas por proyecto/field/lot
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_workorder_metrics AS
WITH base AS (
  SELECT
    w.id                         AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price                     AS labor_price
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
-- Superficie correcta: suma única por workorder, sin duplicar por insumos
surface AS (
  SELECT project_id, field_id, lot_id, SUM(effective_area)::numeric AS surface_ha
  FROM base
  GROUP BY project_id, field_id, lot_id
),
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM(v3_calc.labor_cost(labor_price, effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    SUM(v3_calc.supply_cost(wi.final_dose::double precision,
                            s.price::numeric,
                            b.effective_area))::numeric          AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
         ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id,  lc.field_id,  sm.field_id)     AS field_id,
  COALESCE(sur.lot_id,    lc.lot_id,    sm.lot_id)       AS lot_id,
  COALESCE(sur.surface_ha, 0)::numeric                    AS surface_ha,
  COALESCE(sm.liters, 0)::numeric                         AS liters,
  COALESCE(sm.kilograms, 0)::numeric                      AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric                 AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric              AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric)            AS direct_cost_usd,
  v3_calc.cost_per_ha(
    COALESCE(lc.labor_cost_usd,0)::numeric + COALESCE(sm.supplies_cost_usd,0)::numeric,
    COALESCE(sur.surface_ha,0)::numeric
  )                                                        AS avg_cost_per_ha_usd,
  v3_calc.per_ha(COALESCE(sm.liters,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)     AS liters_per_ha,
  v3_calc.per_ha(COALESCE(sm.kilograms,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)  AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs   lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id);

-- -------------------------------------------------------------------
-- v3_workorder_list: listado a nivel workorder(+item de insumo)
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_workorder_list AS
WITH workorder_surface AS (
  -- Obtener superficie única por workorder
  SELECT 
    w.id,
    w.effective_area AS surface_ha
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
)
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name  AS type_name,
  w.contractor,
  -- Usar superficie única de la CTE, no de la tabla workorders directamente
  ws.surface_ha,
  s.name                                     AS supply_name,
  wi.total_used                              AS consumption,
  cat.name                                   AS category_name,
  wi.final_dose                              AS dose_per_ha,
  s.price                                    AS unit_price,
  -- costo por ha del insumo
  CASE WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
       THEN v3_calc.cost_per_ha(
              (wi.final_dose::double precision * s.price)::numeric,
              1 -- costo ya es por ha, se documenta con 1 ha
            )::numeric
       ELSE 0
  END                                                               AS supply_cost_per_ha,
  -- costo total del insumo para la WO (usa la función SSOT)
  v3_calc.supply_cost(wi.final_dose::double precision,
                      s.price::numeric,
                      ws.surface_ha)::numeric              AS supply_total_cost
FROM public.workorders w
JOIN workorder_surface ws ON ws.id = w.id
JOIN public.projects   p ON p.id = w.project_id   AND p.deleted_at IS NULL
JOIN public.fields     f ON f.id = w.field_id     AND f.deleted_at IS NULL
JOIN public.lots       l ON l.id = w.lot_id       AND l.deleted_at IS NULL
JOIN public.crops      c ON c.id = w.crop_id      AND c.deleted_at IS NULL
JOIN public.labors     lb ON lb.id = w.labor_id   AND lb.deleted_at IS NULL
JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s        ON s.id = wi.supply_id     AND s.deleted_at IS NULL
LEFT JOIN public.types    t        ON t.id = s.type_id        AND t.deleted_at IS NULL
LEFT JOIN public.categories cat    ON cat.id = s.category_id  AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL;
