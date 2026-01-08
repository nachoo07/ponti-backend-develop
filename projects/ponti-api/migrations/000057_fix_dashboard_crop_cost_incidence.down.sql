-- ========================================
-- ROLLBACK MIGRACIÓN 000057: FIX INCIDENCIA DE COSTOS POR CULTIVO
-- ========================================
-- 
-- Objetivo: Restaurar vista dashboard_view al estado anterior
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Restaurar vista dashboard_view a su estado anterior (migración 000056)
CREATE OR REPLACE VIEW dashboard_view AS
WITH base_costs AS (
  SELECT w.project_id,
         SUM(lb.price*w.effective_area) AS executed_labors_usd,
         SUM(wi.total_used*s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
  GROUP BY w.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  f.id AS field_id,
  COALESCE(bc.executed_labors_usd,0) AS executed_labors_usd,
  COALESCE(bc.executed_supplies_usd,0) AS executed_supplies_usd,
  COALESCE(bc.executed_labors_usd,0)+COALESCE(bc.executed_supplies_usd,0) AS executed_costs_usd,
  p.admin_cost AS budget_cost_usd,
  (COALESCE(bc.executed_labors_usd,0)+COALESCE(bc.executed_supplies_usd,0)+p.admin_cost) AS operating_result_total_costs_usd,
  0::numeric AS operating_result_usd, -- placeholder (lo calcula la app)
  0::numeric AS operating_result_pct  -- placeholder (lo calcula la app)
FROM projects p
JOIN fields f ON f.project_id=p.id
LEFT JOIN base_costs bc ON bc.project_id=p.id;
