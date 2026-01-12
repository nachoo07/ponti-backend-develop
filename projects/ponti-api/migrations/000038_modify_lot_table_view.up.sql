-- =======================
-- MIGRACIÓN: Agregar columna hectares a la vista lot_table_view
-- =======================
DROP VIEW IF EXISTS lot_table_view;

CREATE VIEW lot_table_view AS
WITH
-- =======================
-- CÁLCULO DE ÁREA SEMBRADA (optimizado)
-- =======================
sowing AS (
  SELECT 
    w.lot_id, 
    SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 1
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- =======================
-- CÁLCULO DE ÁREA COSECHADA (optimizado)
-- =======================
harvest AS (
  SELECT 
    w.lot_id,
    SUM(w.effective_area) AS harvested_area,
    MAX(w.date) AS last_harvest_date
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 2
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- =======================
-- CÁLCULO DE COSTOS DIRECTOS (optimizado)
-- =======================
direct_costs AS (
  SELECT 
    w.lot_id,
    SUM(COALESCE(lb.price * w.effective_area, 0)) AS labor_cost,
    SUM(COALESCE(wi.final_dose * s.price * w.effective_area, 0)) AS supply_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),
-- =======================
-- CÁLCULO DE INGRESO NETO (optimizado)
-- =======================
income_net AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id 
    AND cc.crop_id = l.current_crop_id 
    AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.tons IS NOT NULL
    AND l.tons > 0
),
-- =======================
-- CÁLCULO DE ARRIENDO (optimizado)
-- =======================
rent_calculation AS (
  SELECT 
    l.id AS lot_id,
    CASE 
      WHEN f.lease_type_id = 1 THEN 
        COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)
      WHEN f.lease_type_id = 2 THEN 
        (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0)
      WHEN f.lease_type_id = 3 THEN 
        (COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)) + 
        ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0))
      ELSE 0
    END AS rent_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN harvest h ON h.lot_id = l.id
  LEFT JOIN income_net in_net ON in_net.lot_id = l.id
  WHERE l.deleted_at IS NULL
),
-- =======================
-- CÁLCULO DE COSTO ADMINISTRATIVO (optimizado)
-- =======================
admin_cost AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(p.admin_cost, 0) * COALESCE(l.hectares, 0) AS admin_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.hectares IS NOT NULL
    AND l.hectares > 0
),
-- =======================
-- CÁLCULO DE FECHAS (optimizado)
-- =======================
lot_dates AS (
  SELECT 
    w.lot_id,
    MIN(CASE WHEN lb.category_id = 1 THEN w.date END) AS lot_sowing_date,
    MAX(CASE WHEN lb.category_id = 2 THEN w.date END) AS lot_harvest_date,
    COUNT(DISTINCT w.id) AS sequence
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id IN (1, 2)
    AND w.date IS NOT NULL
  GROUP BY w.lot_id
)
-- =======================
-- SELECT PRINCIPAL OPTIMIZADO
-- =======================
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
  l.hectares,
  COALESCE(s.sowed_area, 0) AS sowed_area,
  l.season,
  COALESCE(l.tons, 0) AS tons,
  COALESCE(h.harvested_area, 0) AS harvested_area,
  h.last_harvest_date AS harvest_date,
  COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0) AS direct_cost_total,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
    ELSE 0 
  END AS cost_usd_per_ha,
  COALESCE(in_net.income_net_total, 0) AS income_net_total,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(in_net.income_net_total, 0) / s.sowed_area
    ELSE 0 
  END AS income_net_per_ha,
  CASE 
    WHEN COALESCE(h.harvested_area, 0) > 0 
    THEN COALESCE(l.tons, 0) / h.harvested_area
    ELSE 0 
  END AS yield_tn_per_ha,
  COALESCE(rc.rent_total, 0) AS rent_total,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(rc.rent_total, 0) / s.sowed_area
    ELSE 0 
  END AS rent_per_ha,
  COALESCE(ac.admin_total, 0) AS admin_total,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(ac.admin_total, 0) / s.sowed_area
    ELSE 0 
  END AS admin_cost_per_ha,
  (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
  COALESCE(rc.rent_total, 0) + 
  COALESCE(ac.admin_total, 0) AS active_total,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
          COALESCE(rc.rent_total, 0) + 
          COALESCE(ac.admin_total, 0)) / s.sowed_area
    ELSE 0 
  END AS active_total_per_ha,
  COALESCE(in_net.income_net_total, 0) - 
  ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
   COALESCE(rc.rent_total, 0) + 
   COALESCE(ac.admin_total, 0)) AS operating_result,
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(in_net.income_net_total, 0) - 
          ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
           COALESCE(rc.rent_total, 0) + 
           COALESCE(ac.admin_total, 0))) / s.sowed_area
    ELSE 0 
  END AS operating_result_per_ha,
  ld.lot_sowing_date,
  ld.lot_harvest_date,
  COALESCE(ld.sequence, 0) AS sequence,
  l.updated_at
FROM lots l
JOIN fields f ON l.field_id = f.id AND f.deleted_at IS NULL
JOIN projects p ON f.project_id = p.id AND p.deleted_at IS NULL
LEFT JOIN crops pc ON l.previous_crop_id = pc.id AND pc.deleted_at IS NULL
LEFT JOIN crops cc ON l.current_crop_id = cc.id AND cc.deleted_at IS NULL
LEFT JOIN sowing s ON l.id = s.lot_id
LEFT JOIN harvest h ON l.id = h.lot_id
LEFT JOIN direct_costs dc ON l.id = dc.lot_id
LEFT JOIN income_net in_net ON l.id = in_net.lot_id
LEFT JOIN rent_calculation rc ON l.id = rc.lot_id
LEFT JOIN admin_cost ac ON l.id = ac.lot_id
LEFT JOIN lot_dates ld ON l.id = ld.lot_id
WHERE l.deleted_at IS NULL
  AND l.hectares IS NOT NULL
  AND l.hectares > 0;
