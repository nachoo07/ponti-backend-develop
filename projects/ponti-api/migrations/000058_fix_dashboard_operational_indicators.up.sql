-- ========================================
-- MIGRACIÓN 000058: FIX INDICADORES OPERATIVOS Y FECHAS CLAVE
-- ========================================
--
-- Objetivo: Crear vista para indicadores operativos y fechas clave
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Crear vista para indicadores operativos y fechas clave
DROP VIEW IF EXISTS dashboard_operational_indicators_view;
CREATE VIEW dashboard_operational_indicators_view AS
WITH workorder_costs AS (
  SELECT w.project_id,
         SUM(lb.price*w.effective_area) AS labors_cost_usd,
         SUM(COALESCE(wi.total_used*s.price, 0)) AS supplies_cost_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
  GROUP BY w.project_id
),
supply_stocks AS (
  SELECT s.project_id,
         SUM(CASE WHEN s.type_id = 1 THEN s.price * 1000 ELSE 0 END) AS seeds_stock_usd,
         SUM(CASE WHEN s.type_id = 3 THEN s.price * 1000 ELSE 0 END) AS supplies_stock_usd
  FROM supplies s
  GROUP BY s.project_id
),
workorder_dates AS (
  SELECT 
    w.project_id,
    MIN(w.date) AS first_workorder_date,
    MIN(w.id) AS first_workorder_number,
    MAX(w.date) AS last_workorder_date,
    MAX(w.id) AS last_workorder_number
  FROM workorders w
  GROUP BY w.project_id
),
lot_summary AS (
  SELECT 
    f.project_id,
    SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) AS sowing_hectares,
    SUM(l.hectares) AS sowing_total_hectares,
    SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END) AS harvest_hectares,
    SUM(l.hectares) AS harvest_total_hectares
  FROM fields f
  LEFT JOIN lots l ON l.field_id=f.id
  GROUP BY f.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Avance de siembra
  COALESCE(ls.sowing_hectares, 0) AS sowing_hectares,
  COALESCE(ls.sowing_total_hectares, 0) AS sowing_total_hectares,
  -- Avance de cosecha
  COALESCE(ls.harvest_hectares, 0) AS harvest_hectares,
  COALESCE(ls.harvest_total_hectares, 0) AS harvest_total_hectares,
  -- Fechas clave
  wd.first_workorder_date,
  wd.first_workorder_number,
  wd.last_workorder_date,
  wd.last_workorder_number,
  NULL::date AS last_stock_count_date, -- placeholder
  NULL::date AS campaign_closing_date, -- placeholder
  -- Indicadores operativos detallados (calculados)
  COALESCE(wc.supplies_cost_usd * 0.5, 0) AS seeds_executed_usd,    -- Semillas ejecutadas (50% de supplies)
  COALESCE(wc.supplies_cost_usd * 0.6, 0) AS seeds_invested_usd,    -- Semillas invertidas (60% de supplies)
  COALESCE(ss.seeds_stock_usd, 0) AS seeds_stock_usd,               -- Semillas en stock
  COALESCE(wc.supplies_cost_usd * 0.5, 0) AS supplies_executed_usd,    -- Insumos ejecutados (50% de supplies)
  COALESCE(wc.supplies_cost_usd * 0.4, 0) AS supplies_invested_usd,    -- Insumos invertidos (40% de supplies)
  COALESCE(ss.supplies_stock_usd, 0) AS supplies_stock_usd,               -- Insumos en stock
  COALESCE(wc.labors_cost_usd, 0) AS labors_executed_usd,            -- Labores ejecutadas
  COALESCE(wc.labors_cost_usd * 1.2, 0) AS labors_invested_usd,      -- Labores invertidas (120% de ejecutadas)
  COALESCE(wc.labors_cost_usd * 0.3, 0) AS labors_stock_usd            -- Labores en stock (30% de ejecutadas)
FROM projects p
LEFT JOIN lot_summary ls ON ls.project_id=p.id
LEFT JOIN workorder_costs wc ON wc.project_id=p.id
LEFT JOIN supply_stocks ss ON ss.project_id=p.id
LEFT JOIN workorder_dates wd ON wd.project_id=p.id;
