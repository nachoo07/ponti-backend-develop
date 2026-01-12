-- =========================================================
-- VIEW optimizada para Cloud SQL (PostgreSQL)
--  - Usa SUM ... FILTER para separar litros vs kilos
--  - Empuja filtros de soft-delete y área válida a los CTEs
--  - Devuelve métricas por (project_id, field_id)
-- =========================================================
DROP VIEW IF EXISTS workorder_metrics_view;

CREATE VIEW workorder_metrics_view AS
WITH
workorder_base AS (
  SELECT
    w.id              AS workorder_id,
    w.project_id,
    w.field_id,
    w.effective_area,
    (COALESCE(lb.price,0) * w.effective_area) AS labor_cost_per_wo
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
supply_agg AS (
  SELECT
    w.id         AS workorder_id,
    w.project_id,
    w.field_id,
    -- costo insumos
    SUM(COALESCE(wi.final_dose,0) * COALESCE(s.price,0) * w.effective_area) AS total_supplies_cost,
    -- litros: Herbicida(2), Fungicida(5)
    SUM(COALESCE(wi.final_dose,0) * w.effective_area)
      FILTER (WHERE s.category_id IN (2,5)) AS total_liters,
    -- kilogramos: Fertilizante(3), Semilla(4)
    SUM(COALESCE(wi.final_dose,0) * w.effective_area)
      FILTER (WHERE s.category_id IN (3,4)) AS total_kilograms
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id
  JOIN supplies s         ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL
    AND wi.deleted_at IS NULL
    AND s.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.id, w.project_id, w.field_id
),
field_metrics AS (
  SELECT
    wb.project_id,
    wb.field_id,
    SUM(wb.effective_area)                   AS total_surface_ha,
    SUM(wb.labor_cost_per_wo)                AS total_labor_cost,
    SUM(COALESCE(sa.total_supplies_cost,0))  AS total_supplies_cost,
    SUM(COALESCE(sa.total_liters,0))         AS total_liters,
    SUM(COALESCE(sa.total_kilograms,0))      AS total_kilograms,
    COUNT(DISTINCT wb.workorder_id)          AS total_workorders
  FROM workorder_base wb
  LEFT JOIN supply_agg sa ON sa.workorder_id = wb.workorder_id
  GROUP BY wb.project_id, wb.field_id
)
SELECT
  fm.project_id,
  fm.field_id,
  fm.total_surface_ha                                             AS surface_ha,
  COALESCE(fm.total_liters,0)                                     AS liters,
  COALESCE(fm.total_kilograms,0)                                  AS kilograms,
  (fm.total_labor_cost + fm.total_supplies_cost)                  AS direct_cost
FROM field_metrics fm
WHERE fm.total_surface_ha > 0;
