CREATE OR REPLACE VIEW labor_cards_cube_view AS
SELECT
  w.project_id,
  w.field_id,
  CASE
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=0 THEN 'project+field'
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=1 THEN 'project'
    WHEN GROUPING(w.project_id)=1 AND GROUPING(w.field_id)=0 THEN 'field'
    ELSE 'global'
  END AS level,
  SUM(w.effective_area) AS surface_ha,
  SUM(lb.price) AS total_labor_cost,
  CASE
    WHEN SUM(w.effective_area) > 0
      THEN SUM(lb.price * w.effective_area) / SUM(w.effective_area)
    ELSE 0
  END AS labor_cost_per_ha
FROM workorders w
JOIN labors lb ON lb.id = w.labor_id
WHERE
  w.deleted_at IS NULL
  AND lb.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL
GROUP BY GROUPING SETS (
  (w.project_id, w.field_id),
  (w.project_id),
  (w.field_id),
  ()
);