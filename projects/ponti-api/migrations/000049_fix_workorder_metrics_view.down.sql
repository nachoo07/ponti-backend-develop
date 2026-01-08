-- =========================================================
-- REVERTIR: workorder_metrics_view a la versión anterior
-- =========================================================
DROP VIEW IF EXISTS workorder_metrics_view;

-- Restaurar la vista original de la migración 000042
CREATE VIEW workorder_metrics_view AS
WITH
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    p.customer_id,
    p.campaign_id,
    w.effective_area,
    COALESCE(lb.price, 0) AS labor_price_per_ha,
    COALESCE(lb.price, 0) * w.effective_area AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  JOIN projects p ON p.id = w.project_id
  WHERE w.deleted_at IS NULL 
    AND lb.deleted_at IS NULL
    AND p.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
supply_aggregation AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    p.customer_id,
    p.campaign_id,
    SUM(COALESCE(wi.final_dose, 0) * COALESCE(s.price, 0) * w.effective_area) AS total_supplies_cost,
    SUM(COALESCE(wi.final_dose, 0) * w.effective_area) AS total_liters,
    SUM(COALESCE(wi.final_dose, 0) * w.effective_area) AS total_kilograms
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  JOIN projects p         ON p.id = w.project_id AND p.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.id, w.project_id, w.field_id, p.customer_id, p.campaign_id
),
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    wb.customer_id,
    wb.campaign_id,
    SUM(wb.effective_area) AS total_surface_ha,
    SUM(wb.labor_cost_per_wo) AS total_labor_cost,
    SUM(COALESCE(sa.total_supplies_cost, 0)) AS total_supplies_cost,
    SUM(wb.labor_cost_per_wo + COALESCE(sa.total_supplies_cost, 0)) AS total_direct_cost,
    SUM(COALESCE(sa.total_liters, 0)) AS total_liters,
    SUM(COALESCE(sa.total_kilograms, 0)) AS total_kilograms,
    COUNT(DISTINCT wb.workorder_id) AS total_workorders
  FROM workorder_base wb
  LEFT JOIN supply_aggregation sa ON sa.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id, wb.customer_id, wb.campaign_id
)
SELECT
  fm.project_id,
  fm.field_id,
  fm.customer_id,
  fm.campaign_id,
  fm.total_surface_ha AS surface_ha,
  COALESCE(fm.total_liters, 0) AS liters,
  COALESCE(fm.total_kilograms, 0) AS kilograms,
  fm.total_direct_cost AS direct_cost
FROM field_metrics fm
WHERE fm.total_surface_ha > 0;
