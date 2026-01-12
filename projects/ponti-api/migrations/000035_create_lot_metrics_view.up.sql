CREATE OR REPLACE VIEW public.lot_metrics_view AS
WITH 
-- Cálculo de área sembrada por lote
sowing AS (
  SELECT w.lot_id, SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 1  -- ID de "Categoría 1" (usar la que existe)
  GROUP BY w.lot_id
),
-- Cálculo de área cosechada por lote (WO con labor "Cosecha")
harvest AS (
  SELECT w.lot_id, SUM(w.effective_area) AS harvested_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 2  -- ID de "Categoría 2" (usar la que existe)
  GROUP BY w.lot_id
),
-- Cálculo de costos directos por lote
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
base AS (
  -- We take lot data with corrected calculations
  SELECT
    l.id AS lot_id,
    f.project_id,
    l.field_id,
    l.previous_crop_id,
    l.current_crop_id,
    l.hectares,
    COALESCE(s.sowed_area, 0) AS sowed_area,
    COALESCE(h.harvested_area, 0) AS harvested_area,
    COALESCE(l.tons, 0) AS tons,
    -- Total cost per lot
    COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0) AS direct_cost_total,
    -- Cost per sown hectare
    CASE 
      WHEN COALESCE(s.sowed_area, 0) > 0 
      THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
      ELSE 0 
    END AS cost_usd_per_ha
  FROM lots l
  JOIN fields f ON l.field_id = f.id
  LEFT JOIN sowing s ON l.id = s.lot_id
  LEFT JOIN harvest h ON l.id = h.lot_id
  LEFT JOIN direct_costs dc ON l.id = dc.lot_id
  WHERE l.deleted_at IS NULL
    AND f.deleted_at IS NULL
),
rollup AS (
  SELECT
    b.project_id,
    b.field_id,
    b.previous_crop_id,
    b.current_crop_id,
    -- Sown area (sum of all sown areas)
    COALESCE(SUM(b.sowed_area), 0) AS seeded_area,
    -- Harvested area (sum of areas from WO with "Harvest" labor)
    COALESCE(SUM(b.harvested_area), 0) AS harvested_area,
    -- Total tons (sum of manually entered values)
    COALESCE(SUM(b.tons), 0) AS total_harvest,
    -- Yield (tons per harvested hectare)
    CASE 
      WHEN COALESCE(SUM(b.harvested_area), 0) > 0 
      THEN COALESCE(SUM(b.tons), 0) / SUM(b.harvested_area)
      ELSE 0 
    END AS yield_tn_per_ha,
    -- Total direct cost
    COALESCE(SUM(b.direct_cost_total), 0) AS total_direct_cost,
    -- Weighted average cost per sown hectare
    COALESCE(
      SUM(COALESCE(b.cost_usd_per_ha, 0) * b.sowed_area) / NULLIF(SUM(b.sowed_area), 0),
      0
    ) AS weighted_cost_per_ha,
    -- Total hectares of the field
    COALESCE(SUM(b.hectares), 0) AS total_hectares
  FROM base b
  GROUP BY b.project_id, b.field_id, b.previous_crop_id, b.current_crop_id
)
SELECT * FROM rollup;
