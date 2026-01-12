-- Rollback de la migración 000098
-- Restaura la vista v3_workorder_list a su estado anterior (migración 000097)

DROP VIEW IF EXISTS public.v3_workorder_list;

-- Restaurar la vista anterior (basada en la migración 000097)
CREATE VIEW public.v3_workorder_list AS
WITH workorder_surface AS (
  -- Obtener superficie única por workorder
  SELECT 
    w.id,
    w.effective_area AS surface_ha
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
-- UNION de labores e insumos
labor_and_supplies AS (
  -- Filas de LABORES (solo para workorders CON insumos)
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
    'Labor'::character varying(250) AS type_name,
    w.contractor,
    ws.surface_ha,
    '-'::character varying(100) AS supply_name,
    ws.surface_ha AS consumption,  -- Consumo = Superficie
    cat_lb.name AS category_name,
    (ws.surface_ha / ws.surface_ha)::numeric(18,6) AS dose_per_ha,  -- Dosis = 1
    lb.price AS unit_price,
    lb.price::numeric AS supply_cost_per_ha,  -- Costo $/HA = precio labor
    (ws.surface_ha * lb.price)::numeric AS supply_total_cost  -- Total costo = superficie × precio
  FROM public.workorders w
  JOIN workorder_surface ws ON ws.id = w.id
  JOIN public.projects   p ON p.id = w.project_id   AND p.deleted_at IS NULL
  JOIN public.fields     f ON f.id = w.field_id     AND f.deleted_at IS NULL
  JOIN public.lots       l ON l.id = w.lot_id       AND l.deleted_at IS NULL
  JOIN public.crops      c ON c.id = w.crop_id      AND c.deleted_at IS NULL
  JOIN public.labors     lb ON lb.id = w.labor_id   AND lb.deleted_at IS NULL
  JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
  -- Solo workorders que TENGAN insumos asociados
  WHERE w.deleted_at IS NULL
    AND EXISTS (
      SELECT 1 FROM public.workorder_items wi 
      WHERE wi.workorder_id = w.id 
        AND wi.deleted_at IS NULL
    )

  UNION ALL

  -- Filas de INSUMOS (existente)
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
    ws.surface_ha,
    s.name  AS supply_name,
    wi.total_used AS consumption,
    cat.name AS category_name,
    (wi.total_used / ws.surface_ha)::numeric(18,6) AS dose_per_ha,  -- Dosis = consumo/superficie
    s.price AS unit_price,
    (s.price * (wi.total_used / ws.surface_ha))::numeric AS supply_cost_per_ha,  -- Costo $/HA = precio × dosis
    (wi.total_used * s.price)::numeric AS supply_total_cost  -- Total costo = consumo × precio
  FROM public.workorders w
  JOIN workorder_surface ws ON ws.id = w.id
  JOIN public.projects   p ON p.id = w.project_id   AND p.deleted_at IS NULL
  JOIN public.fields     f ON f.id = w.field_id     AND f.deleted_at IS NULL
  JOIN public.lots       l ON l.id = w.lot_id       AND l.deleted_at IS NULL
  JOIN public.crops      c ON c.id = w.crop_id      AND c.deleted_at IS NULL
  JOIN public.labors     lb ON lb.id = w.labor_id   AND lb.deleted_at IS NULL
  JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
  LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
)
SELECT * FROM labor_and_supplies
ORDER BY id, type_name;
