-- Migración 000052: Corregir SOLO el avance de cosecha
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de cosecha
-- Partir de la migración 000050 (no tocar la 000051 ni 000052)

DROP VIEW IF EXISTS dashboard_harvest_progress_view;

CREATE OR REPLACE VIEW dashboard_harvest_progress_view AS
WITH harvest_workorders AS (
  SELECT DISTINCT w.lot_id
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE lb.category_id = 13  -- ID 13 = Cosecha (según seeds.sql)
    AND w.effective_area > 0
),
harvest_lots AS (
  SELECT 
    f.project_id,
    SUM(CASE WHEN hw.lot_id IS NOT NULL THEN l.hectares ELSE 0 END) AS harvested_hectares,
    SUM(l.hectares) AS total_hectares
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  LEFT JOIN harvest_workorders hw ON hw.lot_id = l.id
  GROUP BY f.project_id
)
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(hl.harvested_hectares, 0) AS harvest_hectares,
    COALESCE(hl.total_hectares, 0) AS harvest_total_hectares,
    CASE WHEN COALESCE(hl.total_hectares, 0) > 0
         THEN (COALESCE(hl.harvested_hectares, 0)::numeric / hl.total_hectares::numeric * 100)
    ELSE 0 END AS harvest_progress_pct
FROM projects p
LEFT JOIN harvest_lots hl ON hl.project_id = p.id;


