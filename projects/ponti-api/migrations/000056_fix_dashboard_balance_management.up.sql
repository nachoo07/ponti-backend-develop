-- ========================================
-- MIGRACIÓN 000055: FIX BALANCE DE GESTIÓN
-- ========================================
-- 
-- Objetivo: Crear vista para el balance de gestión con desglose completo
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Crear vista para el balance de gestión con desglose completo
DROP VIEW IF EXISTS dashboard_balance_management_view;
CREATE VIEW dashboard_balance_management_view AS
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
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- SEMILLA
  COALESCE(ec.seeds_executed_usd, 0) AS seeds_executed_usd,             -- Semilla Ejecutados
  COALESCE(ic.seeds_invested_usd, 0) AS seeds_invested_usd,             -- Semilla Invertidos  
  COALESCE(sc.seeds_stock_usd, 0) AS seeds_stock_usd,                   -- Semilla Stock (Invertidos - Ejecutados)
  
  -- INSUMOS
  COALESCE(ec.supplies_executed_usd, 0) AS supplies_executed_usd,        -- Insumos Ejecutados (no semilla)
  COALESCE(ic.supplies_invested_usd, 0) AS supplies_invested_usd,        -- Insumos Invertidos (no semilla)
  COALESCE(sc.supplies_stock_usd, 0) AS supplies_stock_usd,             -- Insumos Stock (Invertidos - Ejecutados)
  
  -- LABORES
  COALESCE(ec.labors_executed_usd, 0) AS labors_executed_usd,            -- Labores Ejecutados
  COALESCE(ic.labors_invested_usd, 0) AS labors_invested_usd,            -- Labores Invertidos
  COALESCE(sc.labors_stock_usd, 0) AS labors_stock_usd,                 -- Labores Stock (Invertidos - Ejecutados)
  
  -- COSTOS DIRECTOS TOTALES
  COALESCE(ec.seeds_executed_usd, 0) + COALESCE(ec.supplies_executed_usd, 0) + COALESCE(ec.labors_executed_usd, 0) AS direct_costs_executed_usd,      -- Costos Directos Ejecutados (Semilla + Insumos + Labores)
  COALESCE(ic.seeds_invested_usd, 0) + COALESCE(ic.supplies_invested_usd, 0) + COALESCE(ic.labors_invested_usd, 0) AS direct_costs_invested_usd,      -- Costos Directos Invertidos (Semilla + Insumos + Labores)
  COALESCE(sc.seeds_stock_usd, 0) + COALESCE(sc.supplies_stock_usd, 0) + COALESCE(sc.labors_stock_usd, 0) AS direct_costs_stock_usd,         -- Costos Directos Stock (Invertidos - Ejecutados)
  
  -- OTROS COSTOS
  0::numeric AS lease_invested_usd,                                       -- Arriendo Invertidos (30% de comercializaciones) - placeholder
  p.admin_cost AS structure_invested_usd,                                -- Estructura Invertidos (admin_cost del proyecto)
  
  -- TOTAL INVERTIDO
  (COALESCE(ic.seeds_invested_usd, 0) + COALESCE(ic.supplies_invested_usd, 0) + COALESCE(ic.labors_invested_usd, 0) + 0 + p.admin_cost) AS total_invested_usd              -- Total Invertido (Directos + Arriendo + Estructura)
  
FROM projects p
LEFT JOIN executed_costs ec ON ec.project_id = p.id
LEFT JOIN invested_costs ic ON ic.project_id = p.id
LEFT JOIN stock_costs sc ON sc.project_id = p.id;
