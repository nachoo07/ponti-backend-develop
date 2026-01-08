-- ========================================
-- MIGRACIÓN 000072: CORREGIR VISTAS DEPRECADAS Y DUPLICACIONES
-- Entidad: views (Corregir vistas deprecadas y eliminar duplicaciones)
-- Funcionalidad: Actualizar todas las entidades para usar las vistas corregidas
-- ========================================

-- ========================================
-- 1. CREAR NUEVA VISTA LABOR_CARDS_CUBE_VIEW CORREGIDA
-- ========================================
DROP VIEW IF EXISTS labor_cards_cube_view_v2;

CREATE VIEW labor_cards_cube_view_v2 AS
WITH labor_metrics AS (
  SELECT
    bdc.project_id,
    bdc.field_id,
    -- Usar vista base para superficie y costos labor
    SUM(w.effective_area) AS surface_ha,
    SUM(bdc.labor_cost) AS total_labor_cost,
    CASE 
      WHEN SUM(w.effective_area) > 0 
      THEN SUM(bdc.labor_cost) / SUM(w.effective_area) 
      ELSE 0 
    END AS labor_cost_per_ha
  FROM base_direct_costs_view bdc
  JOIN workorders w ON w.project_id = bdc.project_id 
    AND w.field_id = bdc.field_id 
    AND w.lot_id = bdc.lot_id
  WHERE w.deleted_at IS NULL 
    AND w.effective_area > 0
  GROUP BY bdc.project_id, bdc.field_id
)
SELECT
  project_id,
  field_id,
  'project+field' AS level,
  surface_ha,
  total_labor_cost AS net_total_cost,
  labor_cost_per_ha AS avg_cost_per_ha
FROM labor_metrics
UNION ALL
SELECT
  project_id,
  NULL AS field_id,
  'project' AS level,
  SUM(surface_ha) AS surface_ha,
  SUM(total_labor_cost) AS net_total_cost,
  CASE 
    WHEN SUM(surface_ha) > 0 
    THEN SUM(total_labor_cost) / SUM(surface_ha) 
    ELSE 0 
  END AS avg_cost_per_ha
FROM labor_metrics
GROUP BY project_id
UNION ALL
SELECT
  NULL AS project_id,
  NULL AS field_id,
  'global' AS level,
  SUM(surface_ha) AS surface_ha,
  SUM(total_labor_cost) AS net_total_cost,
  CASE 
    WHEN SUM(surface_ha) > 0 
    THEN SUM(total_labor_cost) / SUM(surface_ha) 
    ELSE 0 
  END AS avg_cost_per_ha
FROM labor_metrics;

-- ========================================
-- 2. CREAR NUEVA VISTA WORKORDER_METRICS_VIEW CORREGIDA
-- ========================================
DROP VIEW IF EXISTS workorder_metrics_view_v2;

CREATE VIEW workorder_metrics_view_v2 AS
SELECT
  w.project_id,
  w.field_id,
  -- w.customer_id, -- REMOVED
  -- w.campaign_id, -- REMOVED
  SUM(w.effective_area) AS surface_ha,
  SUM(CASE WHEN s.unit_id = 1 THEN wi.final_dose * w.effective_area ELSE 0 END) AS liters,
  SUM(CASE WHEN s.unit_id = 2 THEN wi.final_dose * w.effective_area ELSE 0 END) AS kilograms,
  -- Usar vista base para costos directos
  COALESCE(bdc.direct_cost, 0) AS direct_cost
FROM workorders w
LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
LEFT JOIN supplies s ON s.id = wi.supply_id
LEFT JOIN base_direct_costs_view bdc ON bdc.project_id = w.project_id
  AND bdc.field_id = w.field_id
  AND bdc.lot_id = w.lot_id -- Added lot_id for more precise join
WHERE w.deleted_at IS NULL
GROUP BY w.project_id, w.field_id, bdc.direct_cost; -- Removed customer_id, campaign_id

-- ========================================
-- 3. CREAR NUEVA VISTA DASHBOARD_COSTS_PROGRESS_VIEW CORREGIDA
-- ========================================
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2;

CREATE VIEW dashboard_costs_progress_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para costos directos
  COALESCE(SUM(bdc.direct_cost), 0) AS executed_costs_usd,
  p.admin_cost AS budget_cost_usd,
  CASE 
    WHEN p.admin_cost > 0 
    THEN (COALESCE(SUM(bdc.direct_cost), 0) / p.admin_cost) * 100 
    ELSE 0 
  END AS costs_progress_pct
FROM projects p
LEFT JOIN base_direct_costs_view bdc ON bdc.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost;

-- ========================================
-- 4. CREAR NUEVA VISTA DASHBOARD_OPERATING_RESULT_VIEW CORREGIDA
-- ========================================
DROP VIEW IF EXISTS dashboard_operating_result_view_v2;

CREATE VIEW dashboard_operating_result_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para resultado operativo
  COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) AS operating_result_usd,
  COALESCE(SUM(bdc.direct_cost), 0) AS operating_result_total_costs_usd,
  CASE 
    WHEN COALESCE(SUM(bdc.direct_cost), 0) > 0 
    THEN (COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) / COALESCE(SUM(bdc.direct_cost), 0)) * 100 
    ELSE 0 
  END AS operating_result_pct
FROM projects p
LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 5. CREAR NUEVA VISTA DASHBOARD_MANAGEMENT_BALANCE_VIEW CORREGIDA
-- ========================================
DROP VIEW IF EXISTS dashboard_management_balance_view_v2;

CREATE VIEW dashboard_management_balance_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para ingresos netos
  COALESCE(SUM(bin.income_net_total), 0) AS income_usd,
  -- Usar vista base para costos directos
  COALESCE(SUM(bdc.direct_cost), 0) AS costos_directos_ejecutados_usd,
  COALESCE(SUM(bdc.direct_cost), 0) AS costos_directos_invertidos_usd,
  -- Usar vista base para arriendo
  COALESCE(SUM(blc.rent_per_ha * l.hectares), 0) AS arriendo_invertidos_usd,
  -- Usar vista base para costos administrativos
  COALESCE(SUM(bac.admin_cost_per_ha * l.hectares), 0) AS estructura_invertidos_usd,
  -- Usar vista base para resultado operativo
  COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) AS operating_result_usd,
  CASE 
    WHEN COALESCE(SUM(bdc.direct_cost), 0) > 0 
    THEN (COALESCE(SUM(bor.operating_result_per_ha * l.hectares), 0) / COALESCE(SUM(bdc.direct_cost), 0)) * 100 
    ELSE 0 
  END AS operating_result_pct
FROM projects p
LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_lease_calculations_view blc ON blc.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;
