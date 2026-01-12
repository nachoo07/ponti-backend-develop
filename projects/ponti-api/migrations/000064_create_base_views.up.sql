-- ========================================
-- MIGRACIÓN 000070: CREAR VISTAS BASE REUTILIZABLES
-- Entidad: base_views (Vistas base para evitar duplicación)
-- Funcionalidad: Crear vistas base que contengan la lógica común de cálculos
-- ========================================

-- ========================================
-- 1. VISTA BASE PARA COSTOS DIRECTOS (LABOR + SUPPLIES) - APLICANDO PRINCIPIOS DRY/SSOT
-- ========================================
DROP VIEW IF EXISTS base_direct_costs_view;

CREATE VIEW base_direct_costs_view AS
WITH labor_costs AS (
  SELECT
    w.project_id,
    w.field_id,
    w.lot_id,
    -- Cálculo directo de costo labor (se encapsulará en migración 000068)
    SUM(lb.price * w.effective_area) AS labor_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL 
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
  GROUP BY w.project_id, w.field_id, w.lot_id
),
supply_costs AS (
  SELECT
    w.project_id,
    w.field_id,
    w.lot_id,
    -- Cálculo directo de costo supply (se encapsulará en migración 000068)
    SUM(wi.final_dose * s.price * w.effective_area) AS supply_cost
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id
  JOIN supplies s ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL 
    AND w.effective_area > 0
    AND wi.final_dose > 0
    AND s.price IS NOT NULL
  GROUP BY w.project_id, w.field_id, w.lot_id
)
SELECT
  COALESCE(lc.project_id, sc.project_id) AS project_id,
  COALESCE(lc.field_id, sc.field_id) AS field_id,
  COALESCE(lc.lot_id, sc.lot_id) AS lot_id,
  COALESCE(lc.labor_cost, 0) AS labor_cost,
  COALESCE(sc.supply_cost, 0) AS supply_cost,
  COALESCE(lc.labor_cost, 0) + COALESCE(sc.supply_cost, 0) AS direct_cost
FROM labor_costs lc
FULL OUTER JOIN supply_costs sc ON lc.project_id = sc.project_id 
  AND lc.field_id = sc.field_id 
  AND lc.lot_id = sc.lot_id;

-- ========================================
-- 2. VISTA BASE PARA CÁLCULOS DE RENDIMIENTO - APLICANDO PRINCIPIOS DRY/SSOT
-- ========================================
DROP VIEW IF EXISTS base_yield_calculations_view;

CREATE VIEW base_yield_calculations_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.current_crop_id,
  l.hectares,
  l.tons,
  -- Cálculo directo de rendimiento (se encapsulará en migración 000068)
  COALESCE(l.tons / NULLIF(l.hectares, 0), 0) AS yield_tn_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
WHERE l.deleted_at IS NULL;

-- ========================================
-- 3. VISTA BASE PARA CÁLCULOS DE INGRESO NETO
-- ========================================
DROP VIEW IF EXISTS base_income_net_view;

CREATE VIEW base_income_net_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.current_crop_id,
  l.hectares,
  l.tons,
  cc.net_price AS net_price_usd,
  (l.tons * cc.net_price) AS income_net_total,
  CASE 
    WHEN l.hectares > 0 
    THEN (l.tons * cc.net_price) / l.hectares 
    ELSE 0 
  END AS income_net_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id 
  AND cc.crop_id = l.current_crop_id
  AND cc.deleted_at IS NULL
WHERE l.deleted_at IS NULL;

-- ========================================
-- 4. VISTA BASE PARA CÁLCULOS DE COSTOS ADMINISTRATIVOS
-- ========================================
DROP VIEW IF EXISTS base_admin_costs_view;

CREATE VIEW base_admin_costs_view AS
WITH project_total_hectares AS (
  SELECT
    p.id AS project_id,
    COALESCE(SUM(l.hectares), 1) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
)
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  p.admin_cost,
  pth.total_hectares,
  CASE 
    WHEN l.hectares > 0 
    THEN p.admin_cost / pth.total_hectares 
    ELSE 0 
  END AS admin_cost_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
JOIN project_total_hectares pth ON pth.project_id = p.id
WHERE l.deleted_at IS NULL;

-- ========================================
-- 5. VISTA BASE PARA CÁLCULOS DE ARRIENDO
-- ========================================
DROP VIEW IF EXISTS base_lease_calculations_view;

CREATE VIEW base_lease_calculations_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  f.lease_type_id,
  f.lease_type_percent,
  f.lease_type_value,
  -- Ingreso neto por hectárea (desde base_income_net_view)
  bin.income_net_per_ha,
  -- Costo por hectárea (desde base_direct_costs_view)
  bdc.direct_cost / NULLIF(l.hectares, 0) AS cost_per_ha,
  -- Costo administrativo por hectárea (desde base_admin_costs_view)
  bac.admin_cost_per_ha,
  -- Función para convertir porcentaje a decimal
  CASE
    WHEN f.lease_type_id = 1 THEN -- % INGRESO NETO
      (COALESCE(f.lease_type_percent, 0) / 100.0) * bin.income_net_per_ha
    WHEN f.lease_type_id = 2 THEN -- % UTILIDAD
      (COALESCE(f.lease_type_percent, 0) / 100.0) *
      (bin.income_net_per_ha - (bdc.direct_cost / NULLIF(l.hectares, 0)) - bac.admin_cost_per_ha)
    WHEN f.lease_type_id = 3 THEN -- ARRIENDO FIJO
      COALESCE(f.lease_type_value, 0)
    WHEN f.lease_type_id = 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
      COALESCE(f.lease_type_value, 0) + 
      ((COALESCE(f.lease_type_percent, 0) / 100.0) * bin.income_net_per_ha)
    ELSE 0
  END AS rent_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
WHERE l.deleted_at IS NULL;

-- ========================================
-- 6. VISTA BASE PARA CÁLCULOS DE ACTIVO TOTAL
-- ========================================
DROP VIEW IF EXISTS base_active_total_view;

CREATE VIEW base_active_total_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  -- Costo por hectárea
  bdc.direct_cost / NULLIF(l.hectares, 0) AS cost_per_ha,
  -- Arriendo por hectárea
  blc.rent_per_ha,
  -- Costo administrativo por hectárea
  bac.admin_cost_per_ha,
  -- Total activo por hectárea
  (bdc.direct_cost / NULLIF(l.hectares, 0)) + blc.rent_per_ha + bac.admin_cost_per_ha AS active_total_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_lease_calculations_view blc ON blc.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
WHERE l.deleted_at IS NULL;

-- ========================================
-- 7. VISTA BASE PARA CÁLCULOS DE RESULTADO OPERATIVO
-- ========================================
DROP VIEW IF EXISTS base_operating_result_view;

CREATE VIEW base_operating_result_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  -- Ingreso neto por hectárea
  bin.income_net_per_ha,
  -- Activo total por hectárea
  bat.active_total_per_ha,
  -- Resultado operativo por hectárea
  bin.income_net_per_ha - bat.active_total_per_ha AS operating_result_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_active_total_view bat ON bat.lot_id = l.id
WHERE l.deleted_at IS NULL;
