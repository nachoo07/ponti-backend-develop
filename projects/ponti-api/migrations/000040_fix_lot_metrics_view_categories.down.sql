-- =======================
-- REVERTIR CORRECCIÓN DE VISTA LOT_METRICS_VIEW
-- =======================
-- Esta migración revierte la corrección de categorías de la migración 000039
-- Restaura las categorías originales (incorrectas):
-- - lb.category_id = 1 (Semilla) para área sembrada (mantener)
-- - lb.category_id = 7 → lb.category_id = 2 (Cosecha) para área cosechada

-- Recrear vista con categorías originales (incorrectas)
CREATE OR REPLACE VIEW lot_metrics_view AS
WITH
seeding AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    SUM(w.effective_area) AS seeded_area
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  JOIN workorders w ON w.lot_id = l.id
  JOIN labors lb ON lb.id = w.labor_id
  WHERE p.deleted_at IS NULL
    AND f.deleted_at IS NULL
    AND l.deleted_at IS NULL
    AND w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 1  -- ID de "Semilla" (mantener)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY p.id, p.name
),
harvesting AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    SUM(w.effective_area) AS harvested_area
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  JOIN workorders w ON w.lot_id = l.id
  JOIN labors lb ON lb.id = w.labor_id
  WHERE p.deleted_at IS NULL
    AND f.deleted_at IS NULL
    AND l.deleted_at IS NULL
    AND w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 2  -- ID de "Cosecha" (categoría que no existe)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY p.id, p.name
),
yields AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    CASE 
      WHEN SUM(w.effective_area) > 0 THEN SUM(w.tons) / SUM(w.effective_area)
      ELSE 0
    END AS yield_tn_per_ha
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  JOIN workorders w ON w.lot_id = l.id
  JOIN labors lb ON lb.id = w.labor_id
  WHERE p.deleted_at IS NULL
    AND f.deleted_at IS NULL
    AND l.deleted_at IS NULL
    AND w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 2  -- ID de "Cosecha" (categoría que no existe)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND w.tons IS NOT NULL
    AND w.tons > 0
  GROUP BY p.id, p.name
),
costs AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    CASE 
      WHEN SUM(w.effective_area) > 0 THEN SUM(w.direct_cost_total) / SUM(w.effective_area)
      ELSE 0
    END AS cost_per_hectare
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  JOIN workorders w ON w.lot_id = l.id
  JOIN labors lb ON lb.id = w.labor_id
  WHERE p.deleted_at IS NULL
    AND f.deleted_at IS NULL
    AND l.deleted_at IS NULL
    AND w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 1  -- ID de "Semilla" (mantener)
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND w.direct_cost_total IS NOT NULL
    AND w.direct_cost_total > 0
  GROUP BY p.id, p.name
)
SELECT
  COALESCE(s.project_id, h.project_id, y.project_id, c.project_id) AS project_id,
  COALESCE(s.project_name, h.project_name, y.project_name, c.project_name) AS project_name,
  COALESCE(s.seeded_area, 0) AS seeded_area,
  COALESCE(h.harvested_area, 0) AS harvested_area,
  COALESCE(y.yield_tn_per_ha, 0) AS yield_tn_per_ha,
  COALESCE(c.cost_per_hectare, 0) AS cost_per_hectare
FROM seeding s
FULL OUTER JOIN harvesting h ON h.project_id = s.project_id
FULL OUTER JOIN yields y ON y.project_id = COALESCE(s.project_id, h.project_id)
FULL OUTER JOIN costs c ON c.project_id = COALESCE(s.project_id, h.project_id, y.project_id)
ORDER BY project_id;

-- Eliminar índices de la migración anterior
DROP INDEX IF EXISTS idx_lot_metrics_labors_harvest;

-- Recrear índice con categoría original (incorrecta)
CREATE INDEX IF NOT EXISTS idx_lot_metrics_labors_harvest
  ON labors(id, category_id)
  WHERE deleted_at IS NULL AND category_id = 2;
