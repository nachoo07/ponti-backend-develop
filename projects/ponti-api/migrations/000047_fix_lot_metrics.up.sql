-- =======================
-- MIGRATION: lot_metrics_view (English only, 4 cards)
-- =======================
-- This migration creates the lot_metrics_view with English column names
-- and only the 4 required metrics (cards):
--  - seeded_area
--  - harvested_area
--  - yield_tn_per_ha
--  - cost_per_ha
-- =======================

DROP VIEW IF EXISTS public.lot_metrics_view;

CREATE VIEW public.lot_metrics_view AS
WITH
-- 1) Labor aggregation per lot
labor_agg AS (
  SELECT
    w.lot_id,
    -- Sowing (category_id = 9), Harvest (category_id = 13)
    SUM(w.effective_area) FILTER (WHERE lb.category_id = 9)  AS seeded_area_lot,
    SUM(w.effective_area) FILTER (WHERE lb.category_id = 13) AS harvested_area_lot,
    -- Labor cost (all labors)
    SUM(COALESCE(lb.price,0) * w.effective_area)             AS labor_cost_lot
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- 2) Supply aggregation per lot
supply_agg AS (
  SELECT
    w.lot_id,
    SUM(COALESCE(wi.final_dose,0) * COALESCE(s.price,0) * w.effective_area) AS supply_cost_lot
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s         ON s.id = wi.supply_id   AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- 3) Base per lot
lot_base AS (
  SELECT
    f.project_id,
    l.field_id,
    l.current_crop_id,
    COALESCE(la.seeded_area_lot, 0)     AS seeded_area_lot,
    COALESCE(la.harvested_area_lot, 0)  AS harvested_area_lot,
    COALESCE(l.tons, 0)                 AS tons_lot,
    COALESCE(la.labor_cost_lot, 0) + COALESCE(sa.supply_cost_lot, 0) AS direct_cost_lot
  FROM lots l
  JOIN fields f       ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN labor_agg  la ON la.lot_id = l.id
  LEFT JOIN supply_agg sa ON sa.lot_id = l.id
  WHERE l.deleted_at IS NULL
)

-- 4) Final rollup: only the 4 required metrics
SELECT
  b.project_id,
  b.field_id,
  b.current_crop_id,                              -- used to filter yield by crop
  SUM(b.seeded_area_lot)       AS seeded_area,    -- total sowing area
  SUM(b.harvested_area_lot)    AS harvested_area, -- total harvested area
  CASE WHEN SUM(b.harvested_area_lot) > 0
       THEN SUM(b.tons_lot) / SUM(b.harvested_area_lot)
       ELSE 0 END             AS yield_tn_per_ha, -- yield in tons per harvested hectare
  CASE WHEN SUM(b.seeded_area_lot) > 0
       THEN SUM(b.direct_cost_lot) / SUM(b.seeded_area_lot)
       ELSE 0 END             AS cost_per_ha      -- average direct cost per seeded hectare
FROM lot_base b
GROUP BY b.project_id, b.field_id, b.current_crop_id;
