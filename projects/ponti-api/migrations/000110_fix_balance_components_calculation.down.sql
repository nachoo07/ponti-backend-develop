-- ========================================
-- MIGRACIÓN 000110: FIX BALANCE COMPONENTS CALCULATION (DOWN)
-- ========================================
-- 
-- Propósito: Revertir a la vista de 000104

-- Eliminar vista existente si existe
DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

-- Crear vista con estructura original de 000104
CREATE VIEW public.v3_dashboard_management_balance AS
WITH lots_base AS (
  SELECT
    l.id         AS lot_id,
    f.project_id AS project_id,
    l.hectares   AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM lots_base
  GROUP BY project_id
)
SELECT
  p.id AS project_id,
  
  -- ========================================
  -- INGRESOS Y RESULTADO OPERATIVO
  -- ========================================
  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0) AS income_usd,
  v3_calc.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_calc.renta_pct(
    v3_calc.operating_result_total_for_project(p.id),
    (COALESCE(v3_calc.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct,
  
  -- ========================================
  -- COSTOS DIRECTOS
  -- ========================================
  v3_calc.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  v3_calc.direct_costs_invested_for_project_mb(p.id) AS costos_directos_invertidos_usd,
  v3_calc.stock_value_for_project_mb(p.id) AS costos_directos_stock_usd,
  
  -- ========================================
  -- SEMILLAS
  -- ========================================
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_ejecutados_usd,
  v3_calc.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  v3_calc.seeds_invested_for_project_mb(p.id) - 
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_stock_usd,
  
  -- ========================================
  -- AGROQUÍMICOS
  -- ========================================
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_ejecutados_usd,
  v3_calc.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  v3_calc.agrochemicals_invested_for_project_mb(p.id) - 
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_stock_usd,
  
  -- ========================================
  -- LABORES
  -- ========================================
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_invertidos_usd,
  
  -- ========================================
  -- ARRIENDO
  -- ========================================
  0::double precision AS arriendo_ejecutados_usd,
  COALESCE((SELECT f.lease_type_value * ph.total_hectares
            FROM public.fields f
            JOIN project_hectares ph ON ph.project_id = f.project_id
            WHERE f.project_id = p.id AND f.deleted_at IS NULL
            LIMIT 1), 0)::double precision AS arriendo_invertidos_usd,
  
  -- ========================================
  -- ESTRUCTURA
  -- ========================================
  0::double precision AS estructura_ejecutados_usd,
  COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision AS estructura_invertidos_usd,
  
  -- ========================================
  -- COSTOS CALCULADOS (VERSIÓN ORIGINAL)
  -- ========================================
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semilla_cost,
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS insumos_cost,
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_cost

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY 
  p.id, 
  p.admin_cost, 
  ph.total_hectares;

