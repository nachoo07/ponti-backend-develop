-- ========================================
-- MIGRACIÓN 000122: CREATE v3_dashboard_management_balance VIEW (UP)
-- ========================================
-- 
-- Propósito: Vista de balance de gestión del dashboard (SOLO ensamblaje con SSOT)
-- Dependencias: Requiere v3_core_ssot (000113), v3_lot_ssot (000115),
--               v3_dashboard_ssot (000116)
-- Arquitectura: Vista que SOLO ensambla, NO calcula (usa funciones SSOT)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY FASE 2:
-- - Todos los cálculos inline movidos a v3_dashboard_ssot (000116)
-- - Vista solo ensambla resultados de funciones SSOT
-- - NO hay hardcodeos ni cálculos inline
-- 
-- Nota: Vistas SOLO ensamblan, NO calculan (usan funciones SSOT)

BEGIN;

-- ========================================
-- CREAR VISTA: v3_dashboard_management_balance
-- ========================================
-- Propósito: Balance de gestión con ejecutados/invertidos/stock

CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
SELECT
  p.id AS project_id,
  
  -- Ingresos y Resultado
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  
  -- Costos Directos
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) + 
   COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0)) AS costos_directos_invertidos_usd,
  ((v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) + 
    COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0)) - 
   v3_dashboard_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
  
  -- Semillas
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) - 
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)) AS semillas_stock_usd,
  
  -- Agroquímicos
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_base(l.id) - 
               v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) - 
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_base(l.id) - 
                v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)) AS agroquimicos_stock_usd,
  
  -- Labores
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_invertidos_usd,
  
  -- Arriendo (usa funciones SSOT)
  v3_dashboard_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
  v3_dashboard_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
  
  -- Estructura (usa funciones SSOT)
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  
  -- Costos calculados (para compatibilidad)
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semilla_cost,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_base(l.id) - 
               v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS insumos_cost,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost

FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

COMMENT ON VIEW public.v3_dashboard_management_balance IS 'Módulo: Balance de gestión con ejecutados/invertidos/stock (SOLO ensamblaje SSOT)';

COMMIT;
