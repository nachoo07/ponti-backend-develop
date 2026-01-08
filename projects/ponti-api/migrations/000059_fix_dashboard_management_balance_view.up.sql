-- ========================================
-- MIGRACIÓN 000058: FIX DASHBOARD MANAGEMENT BALANCE VIEW
-- ========================================
-- 
-- Objetivo: Crear vista dashboard_management_balance_view con columnas correctas
-- Fecha: 2025-01-27
-- Autor: Sistema

-- Crear vista para el balance de gestión con columnas correctas
DROP VIEW IF EXISTS dashboard_management_balance_view;
CREATE VIEW dashboard_management_balance_view AS
WITH 
-- Costos ejecutados por tipo
executed_costs AS (
  SELECT 
    w.project_id,
    -- Labores ejecutadas
    SUM(CASE WHEN w.effective_area > 0 THEN lb.price * w.effective_area ELSE 0 END) AS labors_executed_usd,
    -- Insumos ejecutados (NO semillas)
    SUM(CASE WHEN wi.final_dose > 0 AND s.type_id != 1 THEN wi.total_used * s.price ELSE 0 END) AS supplies_executed_usd,
    -- Semillas ejecutadas
    SUM(CASE WHEN wi.final_dose > 0 AND s.type_id = 1 THEN wi.total_used * s.price ELSE 0 END) AS seeds_executed_usd
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
  LEFT JOIN supplies s ON s.id = wi.supply_id
  GROUP BY w.project_id
),
-- Costos invertidos por tipo (ejecutados + stock)
invested_costs AS (
  SELECT 
    p.id AS project_id,
    -- Labores invertidas (ejecutadas + no ejecutadas)
    COALESCE(ec.labors_executed_usd, 0) AS labors_invested_usd,
    -- Insumos invertidos (ejecutados + stock)
    COALESCE(ec.supplies_executed_usd, 0) AS supplies_invested_usd,
    -- Semillas invertidas (ejecutadas + stock)
    COALESCE(ec.seeds_executed_usd, 0) AS seeds_invested_usd
  FROM projects p
  LEFT JOIN executed_costs ec ON ec.project_id = p.id
),
-- Stock por tipo (diferencia entre invertido y ejecutado)
stock_costs AS (
  SELECT 
    ic.project_id,
    -- Stock de labores (0 porque no hay labores en stock)
    0::numeric AS labors_stock_usd,
    -- Stock de insumos (diferencia entre invertido y ejecutado)
    GREATEST(0, ic.supplies_invested_usd - COALESCE(ec.supplies_executed_usd, 0)) AS supplies_stock_usd,
    -- Stock de semillas (diferencia entre invertido y ejecutado)
    GREATEST(0, ic.seeds_invested_usd - COALESCE(ec.seeds_executed_usd, 0)) AS seeds_stock_usd
  FROM invested_costs ic
  LEFT JOIN executed_costs ec ON ec.project_id = ic.project_id
)
SELECT
  p.id AS project_id,
  
  -- Ingresos (placeholder - se calcula desde fields/lots)
  0::numeric AS income_usd,
  
  -- Costos directos totales
  COALESCE(ec.seeds_executed_usd, 0) + COALESCE(ec.supplies_executed_usd, 0) + COALESCE(ec.labors_executed_usd, 0) AS costos_directos_ejecutados_usd,
  COALESCE(ic.seeds_invested_usd, 0) + COALESCE(ic.supplies_invested_usd, 0) + COALESCE(ic.labors_invested_usd, 0) AS costos_directos_invertidos_usd,
  COALESCE(sc.seeds_stock_usd, 0) + COALESCE(sc.supplies_stock_usd, 0) + COALESCE(sc.labors_stock_usd, 0) AS costos_directos_stock_usd,
  
  -- Otros costos
  0::numeric AS arriendo_invertidos_usd,                                       -- Arriendo Invertidos (placeholder)
  p.admin_cost AS estructura_invertidos_usd,                                -- Estructura Invertidos (admin_cost del proyecto)
  
  -- Resultado operativo (placeholder - se calcula desde income y costos)
  0::numeric AS operating_result_usd,
  0::numeric AS operating_result_pct
  
FROM projects p
LEFT JOIN executed_costs ec ON ec.project_id = p.id
LEFT JOIN invested_costs ic ON ic.project_id = p.id
LEFT JOIN stock_costs sc ON sc.project_id = p.id;
