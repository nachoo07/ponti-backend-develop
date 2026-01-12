-- ========================================
-- MIGRACIÓN 000071: ACTUALIZAR VISTAS PARA USAR VISTAS BASE
-- Entidad: views (Actualizar vistas existentes)
-- Funcionalidad: Actualizar vistas existentes para usar las vistas base reutilizables
-- ========================================

-- 1. PRIMERO ELIMINAR VISTAS QUE DEPENDEN DE FIX_LOT_LIST
-- ========================================
DROP VIEW IF EXISTS dashboard_operating_result_view CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view CASCADE;
DROP VIEW IF EXISTS dashboard_crop_incidence_view CASCADE;
DROP VIEW IF EXISTS dashboard_operating_result_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2 CASCADE;

-- 2. AHORA ACTUALIZAR FIX_LOT_LIST PARA USAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS fix_lot_list;

CREATE VIEW fix_lot_list AS
SELECT
  f.project_id,
  l.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.id,
  l.name AS lot_name,
  l.variety,
  CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END AS sowed_area,
  l.hectares,
  l.season,
  l.updated_at,
  l.tons,
  l.previous_crop_id,
  pc.name AS previous_crop,
  l.current_crop_id,
  cc_crop.name AS current_crop,
  -- Usar vista base para costo administrativo
  bac.admin_cost_per_ha,
  byc.hectares AS harvested_area,
  -- Función para calcular fecha de cosecha
  CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN CURRENT_DATE ELSE NULL END AS harvest_date,
  -- Usar vista base para costo USD por hectárea
  bdc.direct_cost / NULLIF(l.hectares, 0) AS cost_usd_per_ha,
  -- Usar vista base para rendimiento
  byc.yield_tn_per_ha,
  -- Usar vista base para ingreso neto por hectárea
  bin.income_net_per_ha,
  -- Usar vista base para arriendo por hectárea
  blc.rent_per_ha,
  -- Usar vista base para total activo por hectárea
  bat.active_total_per_ha,
  -- Usar vista base para resultado operativo por hectárea
  bor.operating_result_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN crops pc ON pc.id = l.previous_crop_id AND pc.deleted_at IS NULL
LEFT JOIN crops cc_crop ON cc_crop.id = l.current_crop_id AND cc_crop.deleted_at IS NULL
-- Usar vistas base
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_yield_calculations_view byc ON byc.lot_id = l.id
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
LEFT JOIN base_lease_calculations_view blc ON blc.lot_id = l.id
LEFT JOIN base_active_total_view bat ON bat.lot_id = l.id
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE l.deleted_at IS NULL;

-- ========================================
-- 2. ACTUALIZAR FIX_LOTS_METRICS PARA USAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS fix_lots_metrics;

CREATE VIEW fix_lots_metrics AS
WITH
-- Usar vista base para costos directos en lugar de recalcular
lot_base AS (
  SELECT
    f.project_id, l.field_id, l.current_crop_id,
    -- Usar vista base para área sembrada
    CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END AS seeded_area_lot,
    -- Usar vista base para área cosechada
    CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END AS harvested_area_lot,
    COALESCE(l.tons, 0) AS tons_lot,
    -- Usar vista base para costos directos
    COALESCE(bdc.direct_cost, 0) AS direct_cost_lot
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
  WHERE l.deleted_at IS NULL
)
SELECT
  b.project_id, b.field_id, b.current_crop_id,
  SUM(b.seeded_area_lot) AS seeded_area,
  SUM(b.harvested_area_lot) AS harvested_area,
  CASE WHEN SUM(b.harvested_area_lot) > 0 THEN SUM(b.tons_lot) / SUM(b.harvested_area_lot) ELSE 0 END AS yield_tn_per_ha,
  CASE WHEN SUM(b.seeded_area_lot) > 0 THEN SUM(b.direct_cost_lot) / SUM(b.seeded_area_lot) ELSE 0 END AS cost_per_ha
FROM lot_base b
GROUP BY b.project_id, b.field_id, b.current_crop_id;

-- ========================================
-- 3. ACTUALIZAR DASHBOARD_OPERATING_RESULT_VIEW PARA USAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS dashboard_operating_result_view_v2;

CREATE VIEW dashboard_operating_result_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para resultado operativo
  SUM(bor.operating_result_per_ha * l.hectares) AS operating_result_usd,
  SUM(l.hectares) AS total_hectares,
  CASE 
    WHEN SUM(l.hectares) > 0 
    THEN SUM(bor.operating_result_per_ha * l.hectares) / SUM(l.hectares) 
    ELSE 0 
  END AS operating_result_per_ha
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 4. ACTUALIZAR DASHBOARD_MANAGEMENT_BALANCE_VIEW PARA USAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS dashboard_management_balance_view_v2;

CREATE VIEW dashboard_management_balance_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  -- Usar vista base para costos directos
  SUM(bdc.direct_cost) AS executed_costs_usd,
  p.admin_cost AS budget_cost_usd,
  -- Usar vista base para resultado operativo
  SUM(bor.operating_result_per_ha * l.hectares) AS operating_result_usd,
  SUM(l.hectares) AS total_hectares
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

-- ========================================
-- 5. RECREAR DASHBOARD_CROP_INCIDENCE_VIEW PARA USAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2;

CREATE VIEW dashboard_crop_incidence_view_v2 AS
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  l.current_crop_id,
  cc.name AS crop_name,
  SUM(l.hectares) AS crop_hectares,
  SUM(l.hectares) / NULLIF(SUM(SUM(l.hectares)) OVER (PARTITION BY p.id), 0) * 100 AS crop_incidence_pct
FROM projects p
JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
LEFT JOIN crops cc ON cc.id = l.current_crop_id AND cc.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, l.current_crop_id, cc.name;
