-- ========================================
-- MIGRATION 000085: FIX v3_workorder_list VIEW (DOWN)
-- ========================================
-- 
-- Purpose: Revert v3_workorder_list to original version
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Revertir a la versión original de v3_workorder_list
CREATE OR REPLACE VIEW public.v3_workorder_list AS
WITH workorder_surface AS (
  -- Obtener superficie única por workorder
  SELECT 
    w.id,
    w.effective_area AS surface_ha
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
)
SELECT
  w.id,
  w.number,
  w.project_id,
  w.field_id,
  p.name  AS project_name,
  f.name  AS field_name,
  l.name  AS lot_name,
  w.date,
  c.name  AS crop_name,
  lb.name AS labor_name,
  cat_lb.name AS labor_category_name,
  t.name  AS type_name,
  w.contractor,
  -- Usar superficie única de la CTE, no de la tabla workorders directamente
  ws.surface_ha,
  s.name                                     AS supply_name,
  wi.total_used                              AS consumption,
  cat.name                                   AS category_name,
  wi.final_dose                              AS dose_per_ha,
  s.price                                    AS unit_price,
  -- costo por ha del insumo
  CASE WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
       THEN v3_calc.cost_per_ha(
              (wi.final_dose::double precision * s.price)::numeric,
              1 -- costo ya es por ha, se documenta con 1 ha
            )::numeric
       ELSE 0
  END                                                               AS supply_cost_per_ha,
  -- costo total del insumo para la WO (usa la función SSOT)
  v3_calc.supply_cost(wi.final_dose::double precision,
                      s.price::numeric,
                      ws.surface_ha)::numeric              AS supply_total_cost
FROM public.workorders w
JOIN workorder_surface ws ON ws.id = w.id
JOIN public.projects   p ON p.id = w.project_id   AND p.deleted_at IS NULL
JOIN public.fields     f ON f.id = w.field_id     AND f.deleted_at IS NULL
JOIN public.lots       l ON l.id = w.lot_id       AND l.deleted_at IS NULL
JOIN public.crops      c ON c.id = w.crop_id      AND c.deleted_at IS NULL
JOIN public.labors     lb ON lb.id = w.labor_id   AND lb.deleted_at IS NULL
JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s        ON s.id = wi.supply_id     AND s.deleted_at IS NULL
LEFT JOIN public.types    t        ON t.id = s.type_id        AND t.deleted_at IS NULL
LEFT JOIN public.categories cat    ON cat.id = s.category_id  AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL;
