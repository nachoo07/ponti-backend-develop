-- ========================================
-- MIGRACIÓN 000105: CREATE v3_dashboard_management_balance VIEW (UP)
-- ========================================
-- 
-- Propósito: Crear vista de balance de gestión con cálculos correctos
-- Estructura: Encaja con DTO ManagementBalance existente
-- Fecha: 2025-01-01
-- Autor: Sistema
-- 
-- Nota: Usa funciones SSOT corregidas de 000100 y nuevas funciones _mb de 000104

-- ========================================
-- VISTA v3_dashboard_management_balance
-- ========================================
-- Propósito: Balance de gestión con EJECUTADOS, INVERTIDOS y STOCK correctos
-- Estructura: Compatible con DTO ManagementBalance

-- Eliminar vista existente si existe
DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

-- Crear vista con estructura corregida
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
  -- Ingresos: suma de ingresos netos de todos los lotes
  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0) AS income_usd,
  -- Resultado operativo: usando función SSOT corregida
  v3_calc.operating_result_total_for_project(p.id) AS operating_result_usd,
  -- Porcentaje de margen operativo
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
  -- COSTOS DIRECTOS (EJECUTADOS + INVERTIDOS + STOCK)
  -- ========================================
  -- Ejecutados: desde v3_workorder_metrics (función corregida de 000100)
  v3_calc.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  -- Invertidos: labores planificadas + insumos en stock
  v3_calc.direct_costs_invested_for_project_mb(p.id) AS costos_directos_invertidos_usd,
  -- Stock: invertidos - ejecutados
  v3_calc.stock_value_for_project_mb(p.id) AS costos_directos_stock_usd,
  
  -- ========================================
  -- SEMILLAS (EJECUTADOS + INVERTIDOS + STOCK)
  -- ========================================
  -- Ejecutados: suma de semillas desde workorder_items (unit_id = 2)
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_ejecutados_usd,
  -- Invertidos: usando función SSOT corregida (stock + movimientos stock + remito)
  v3_calc.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  -- Stock: invertidos - ejecutados
  v3_calc.seeds_invested_for_project_mb(p.id) - 
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_stock_usd,
  
  -- ========================================
  -- AGROQUÍMICOS (EJECUTADOS + INVERTIDOS + STOCK)
  -- ========================================
  -- Ejecutados: suma de agroquímicos desde workorder_items (unit_id = 1)
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_ejecutados_usd,
  -- Invertidos: usando función SSOT corregida (stock + movimientos stock + remito)
  v3_calc.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  -- Stock: invertidos - ejecutados
  v3_calc.agrochemicals_invested_for_project_mb(p.id) - 
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_stock_usd,
  
  -- ========================================
  -- LABORES (EJECUTADOS + INVERTIDOS, STOCK = 0)
  -- ========================================
  -- Ejecutados: suma de labores desde workorders
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_ejecutados_usd,
  -- Invertidos: igual a ejecutados (las labores se pagan cuando se ejecutan)
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_invertidos_usd,
  
  -- ========================================
  -- ARRIENDO (EJECUTADOS + INVERTIDOS, STOCK = NULL)
  -- ========================================
  -- Ejecutados: 0 (solo aparece al final si es porcentual)
  0::double precision AS arriendo_ejecutados_usd,
  -- Invertidos: lease_type_value * total_hectares
  COALESCE((SELECT f.lease_type_value * ph.total_hectares
            FROM public.fields f
            JOIN project_hectares ph ON ph.project_id = f.project_id
            WHERE f.project_id = p.id AND f.deleted_at IS NULL
            LIMIT 1), 0)::double precision AS arriendo_invertidos_usd,
  
  -- ========================================
  -- ESTRUCTURA (EJECUTADOS + INVERTIDOS, STOCK = NULL)
  -- ========================================
  -- Ejecutados: 0 (solo aparece al final si es porcentual)
  0::double precision AS estructura_ejecutados_usd,
  -- Invertidos: admin_cost * total_hectares
  COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision AS estructura_invertidos_usd,
  
  -- ========================================
  -- COSTOS CALCULADOS (PARA COMPATIBILIDAD CON MODELO)
  -- ========================================
  -- Costo de semillas: suma de semillas ejecutadas
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semilla_cost,
  -- Costo de insumos: suma de agroquímicos ejecutados
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS insumos_cost,
  -- Costo de labores: suma de labores ejecutadas
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_cost

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY 
  p.id, 
  p.admin_cost, 
  ph.total_hectares;
