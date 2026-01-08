-- =======================
-- REVERTIR MIGRACIÓN 000038
-- =======================
-- Esta migración revierte los cambios de la migración 000038
-- Restaura la vista lot_table_view a su estado original (migración 000037)

-- Recrear la vista original sin las correcciones
DROP VIEW IF EXISTS lot_table_view;

CREATE VIEW lot_table_view AS
WITH
-- Área sembrada por lote
sowing AS (
  SELECT
    w.lot_id,
    SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 9  -- ID de "Siembra" (categoría que no existe)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- Área cosechada por lote
harvest AS (
  SELECT
    w.lot_id,
    SUM(w.effective_area) AS harvested_area,
    MAX(w.date) AS last_harvest_date
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 13  -- ID de "Cosecha" (categoría que no existe)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- Costos directos por lote
direct_costs AS (
  SELECT
    w.lot_id,
    SUM(COALESCE(lb.price * w.effective_area, 0)) AS labor_cost,
    SUM(COALESCE(wi.final_dose * s.price * w.effective_area, 0)) AS supply_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON w.id = wi.workorder_id
  LEFT JOIN supplies s ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND (wi.deleted_at IS NULL OR wi.id IS NULL)
    AND (s.deleted_at IS NULL OR s.id IS NULL)
  GROUP BY w.lot_id
),
-- Ingresos netos por lote
income_net AS (
  SELECT
    l.id AS lot_id,
    COALESCE(l.tons * 1000, 0) AS income_net  -- Usar valor fijo por tonelada
  FROM lots l
  WHERE l.deleted_at IS NULL
    AND l.tons IS NOT NULL
    AND l.tons > 0
),
-- Cálculo de renta por lote (simplificado)
rent_calculation AS (
  SELECT
    w.lot_id,
    SUM(COALESCE(w.effective_area * 100, 0)) AS rent_total  -- Usar valor fijo por hectárea
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 9  -- Solo para labores de siembra
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- Costos administrativos por lote (simplificado)
admin_cost AS (
  SELECT
    l.id AS lot_id,
    COALESCE(0, 0) AS admin_cost  -- Usar valor fijo
  FROM lots l
  WHERE l.deleted_at IS NULL
),
-- Fechas de siembra y cosecha por lote
lot_dates AS (
  SELECT
    w.lot_id,
    MIN(CASE WHEN lb.category_id = 9 THEN w.date END) AS lot_sowing_date,
    MAX(CASE WHEN lb.category_id = 13 THEN w.date END) AS lot_harvest_date,
    COUNT(DISTINCT w.id) AS sequence
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id IN (9, 13) -- IDs de Siembra y Cosecha (que no existen)
    AND w.date IS NOT NULL
  GROUP BY w.lot_id
)
SELECT
  l.id,
  f.project_id,
  l.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  pc.name AS previous_crop,
  l.previous_crop_id,
  cc.name AS current_crop,
  l.current_crop_id,
  l.variety,
  COALESCE(s.sowed_area, 0) AS sowed_area,
  l.season,
  COALESCE(l.tons, 0) AS tons,
  COALESCE(h.harvested_area, 0) AS harvested_area,
  h.last_harvest_date AS harvest_date,
  COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0) AS direct_cost_total,
  COALESCE(ac.admin_cost, 0) AS admin_cost,
  l.updated_at,
  ld.lot_sowing_date,
  ld.lot_harvest_date,
  ld.sequence,
  -- Cálculos derivados
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
    ELSE 0 
  END AS cost_usd_per_ha,
  CASE 
    WHEN COALESCE(h.harvested_area, 0) > 0 
    THEN COALESCE(l.tons, 0) / h.harvested_area
    ELSE 0 
  END AS yield_tn_per_ha,
  CASE 
    WHEN COALESCE(h.harvested_area, 0) > 0 
    THEN COALESCE(income.income_net, 0) / h.harvested_area
    ELSE 0 
  END AS income_net_per_ha,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(rc.rent_total, 0) / s.sowed_area
    ELSE 0 
  END AS rent_per_ha,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0) + COALESCE(ac.admin_cost, 0)) / s.sowed_area
    ELSE 0 
  END AS active_total_per_ha,
  CASE 
    WHEN COALESCE(h.harvested_area, 0) > 0 
    THEN (COALESCE(income.income_net, 0) - COALESCE(dc.labor_cost, 0) - COALESCE(dc.supply_cost, 0) - COALESCE(ac.admin_cost, 0)) / h.harvested_area
    ELSE 0 
  END AS operating_result_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id
JOIN projects p ON p.id = f.project_id
LEFT JOIN crops pc ON pc.id = l.previous_crop_id
LEFT JOIN crops cc ON cc.id = l.current_crop_id
LEFT JOIN sowing s ON l.id = s.lot_id
LEFT JOIN harvest h ON l.id = h.lot_id
LEFT JOIN direct_costs dc ON l.id = dc.lot_id
LEFT JOIN income_net income ON l.id = income.lot_id
LEFT JOIN rent_calculation rc ON l.id = rc.lot_id
LEFT JOIN admin_cost ac ON l.id = ac.lot_id
LEFT JOIN lot_dates ld ON l.id = ld.lot_id
WHERE l.deleted_at IS NULL
  AND f.deleted_at IS NULL
  AND p.deleted_at IS NULL
ORDER BY l.id;
