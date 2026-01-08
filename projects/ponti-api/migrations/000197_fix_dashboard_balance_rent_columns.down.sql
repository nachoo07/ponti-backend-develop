-- ========================================
-- MIGRATION 000197: Fix dashboard management balance rent columns (DOWN)
-- ========================================
--
-- Propósito: Revertir el intercambio de columnas para arriendo en
--            v3_dashboard_management_balance.

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
SELECT
  p.id AS project_id,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
   COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) AS costos_directos_invertidos_usd,
  ((v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
    COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) -
   v3_dashboard_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) -
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)) AS semillas_stock_usd,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) -
   COALESCE(SUM(
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
   ), 0)) AS agroquimicos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_ejecutados_usd,
  (SELECT COALESCE(SUM(sm.quantity * s.price), 0)
   FROM public.supply_movements sm
   JOIN public.supplies s ON s.id = sm.supply_id
   JOIN public.categories c ON s.category_id = c.id
   WHERE sm.project_id = p.id
     AND sm.deleted_at IS NULL
     AND s.deleted_at IS NULL
     AND sm.is_entry = TRUE
     AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
     AND c.type_id = 3) AS fertilizantes_invertidos_usd,
  ((SELECT COALESCE(SUM(sm.quantity * s.price), 0)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id
    JOIN public.categories c ON s.category_id = c.id
    WHERE sm.project_id = p.id
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.is_entry = TRUE
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND c.type_id = 3) -
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0)) AS fertilizantes_stock_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0) AS labores_invertidos_usd,
  v3_dashboard_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
  v3_dashboard_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semilla_cost,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS insumos_cost,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_cost
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

COMMENT ON VIEW public.v3_dashboard_management_balance IS
'Balance de gestión con ejecutados/invertidos/stock. Estado previo a FIX 000197.';

COMMIT;

