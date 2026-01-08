-- ========================================
-- MIGRATION 000080: CREATE v3_labor_views (UP)
-- ========================================
-- 
-- Purpose: Create labor metrics and list views
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Sincronizar category_id de labores para usar la tabla categories
UPDATE labors 
SET category_id = c.id
FROM labor_categories lc
JOIN categories c ON c.name = lc.name AND c.type_id = lc.type_id
WHERE labors.category_id = lc.id;

-- -------------------------------------------------------------------
-- v3_labor_metrics: métricas agregadas de labores
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_metrics AS
WITH wo AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.date,
    w.effective_area::numeric AS effective_area,
    lb.price::numeric AS labor_price_per_ha
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
),
agg AS (
  SELECT
    project_id,
    field_id,
    COUNT(DISTINCT workorder_id) AS total_workorders,
    SUM(effective_area) AS surface_ha,
    SUM(v3_calc.labor_cost(labor_price_per_ha, effective_area)) AS total_labor_cost,
    MIN(date) AS first_workorder_date,
    MAX(date) AS last_workorder_date
  FROM wo
  GROUP BY project_id, field_id
)
SELECT
  a.project_id,
  a.field_id,
  a.surface_ha,
  a.total_labor_cost,
  v3_calc.cost_per_ha(a.total_labor_cost, a.surface_ha) AS avg_labor_cost_per_ha,
  a.total_workorders,
  a.first_workorder_date,
  a.last_workorder_date
FROM agg a;

-- -------------------------------------------------------------------
-- v3_labor_list: listado de órdenes de trabajo con datos relevantes
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_list AS
SELECT
  w.id                  AS workorder_id,
  w.number              AS workorder_number,
  w.date,
  w.project_id,
  p.name                AS project_name,
  w.field_id,
  f.name                AS field_name,
  w.lot_id,
  l.name                AS lot_name,
  w.crop_id,
  c.name                AS crop_name,
  w.labor_id,
  lb.name               AS labor_name,
  cat_lb.id             AS labor_category_id,
  cat_lb.name           AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  w.effective_area      AS surface_ha,
  lb.price              AS cost_per_ha,
  v3_calc.labor_cost(lb.price::numeric, w.effective_area::numeric) AS total_labor_cost,
  w.investor_id,
  i.name                AS investor_name
FROM public.workorders w
JOIN public.projects   p  ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields     f  ON f.id = w.field_id   AND f.deleted_at IS NULL
LEFT JOIN public.lots  l  ON l.id = w.lot_id     AND l.deleted_at IS NULL
LEFT JOIN public.crops c  ON c.id = w.crop_id    AND c.deleted_at IS NULL
JOIN public.labors     lb ON lb.id = w.labor_id  AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.investors  i      ON i.id = w.investor_id        AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;
