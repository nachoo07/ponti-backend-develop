-- VISTA “cubo” con las 3 cards
CREATE OR REPLACE VIEW labor_cards_cube_view AS
SELECT
  w.project_id,
  w.field_id,
  -- etiqueta de nivel (útil al debug/analytics)
  CASE
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=0 THEN 'project+field'
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=1 THEN 'project'
    WHEN GROUPING(w.project_id)=1 AND GROUPING(w.field_id)=0 THEN 'field'
    ELSE 'global'
  END AS level,

  -- Card 1: Superficie total (Ha)
  SUM(w.effective_area) AS surface_ha,

  -- Card 3: Suma de los precios de la labor (SIN multiplicar por Ha)
  SUM(lb.price) AS total_labor_cost,

  -- Card 2: Costo promedio PONDERADO (USD/Ha) = Σ(price*Ha)/Σ(Ha)
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
  AND lb.price IS NOT NULL                 -- el dinero viene siempre de labors en la BDD

GROUP BY GROUPING SETS (
  (w.project_id, w.field_id),              -- project + field
  (w.project_id),                          -- solo project
  (w.field_id),                            -- solo field
  ()                                       -- global
);

/* 
todas las combinaciones y filtros:
por project_id + field_id
por project_id (todas las fields)
por field_id (en todos los projects)
global (sin IDs)

surface_ha → suma de hectáreas trabajadas.
labor_cost_per_ha → promedio ponderado del costo por hectárea = (Σ precio×ha) / (Σ ha).
total_labor_cost → suma directa de los precios de las labores. 
*/