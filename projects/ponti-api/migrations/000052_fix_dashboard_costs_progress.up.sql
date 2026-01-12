-- Migración 000051: Corregir SOLO el avance de costos
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de costos
-- Partir de la migración 000050 (no tocar la 000051)

DROP VIEW IF EXISTS dashboard_costs_progress_view;

CREATE OR REPLACE VIEW dashboard_costs_progress_view AS
WITH labors_costs AS (
  SELECT
    w.project_id,
    SUM(lb.price * w.effective_area) AS executed_labors_usd
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.effective_area > 0
  GROUP BY w.project_id
),
supplies_costs AS (
  SELECT
    w.project_id,
    SUM(wi.total_used * s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id
  JOIN supplies s ON s.id = wi.supply_id
  WHERE wi.final_dose > 0
  GROUP BY w.project_id
)
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(lc.executed_labors_usd,0) AS executed_labors_usd,
    COALESCE(sc.executed_supplies_usd,0) AS executed_supplies_usd,
    COALESCE(lc.executed_labors_usd,0)+COALESCE(sc.executed_supplies_usd,0) AS executed_costs_usd,
    p.admin_cost AS budget_cost_usd,  -- Costos administrativos
    20000::numeric AS budget_total_usd,  -- hardcodeado
    LEAST(
      CASE WHEN 20000>0
          THEN (COALESCE(lc.executed_labors_usd,0)+COALESCE(sc.executed_supplies_usd,0)) / 20000.0 * 100
      ELSE 0 END,100
    ) AS costs_progress_pct
FROM projects p
LEFT JOIN labors_costs lc ON lc.project_id = p.id
LEFT JOIN supplies_costs sc ON sc.project_id = p.id;
