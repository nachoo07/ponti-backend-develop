-- ========================================
-- MIGRACIÓN 000118: CREATE v3_workorder_list (UP)
-- ========================================
-- 
-- Propósito: Recrear vista v3_workorder_list usando funciones SSOT
-- Dependencias: Requiere v3_core_ssot (000113) y v3_workorder_ssot (000114)
-- Cambios: Reemplaza referencias v3_calc.* por v3_core_ssot.*
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- ELIMINAR VISTA ANTIGUA
-- ========================================
DROP VIEW IF EXISTS public.v3_workorder_list CASCADE;

-- ========================================
-- CREAR VISTA v3_workorder_list
-- ========================================
-- Propósito: Listado detallado de workorders con items de insumos

CREATE OR REPLACE VIEW public.v3_workorder_list AS
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
  
  -- Superficie trabajada
  w.effective_area AS surface_ha,
  
  -- Datos del insumo
  s.name AS supply_name,
  wi.total_used AS consumption,
  cat.name AS category_name,
  wi.final_dose AS dose_per_ha,
  s.price AS unit_price,
  
  -- Costo por ha del insumo (usa v3_core_ssot)
  CASE 
    WHEN wi.final_dose IS NOT NULL AND s.price IS NOT NULL
    THEN v3_core_ssot.cost_per_ha(
      (wi.final_dose::double precision * s.price)::numeric,
      1::numeric -- costo ya es por ha
    )
    ELSE 0
  END AS supply_cost_per_ha,
  
  -- Costo total del insumo para la WO (usa v3_core_ssot)
  v3_core_ssot.supply_cost(
    wi.final_dose::double precision,
    s.price::numeric,
    w.effective_area::numeric
  ) AS supply_total_cost
  
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
LEFT JOIN public.types t ON t.id = s.type_id AND t.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
WHERE w.deleted_at IS NULL;

COMMIT;

-- Comentario sobre la vista
COMMENT ON VIEW public.v3_workorder_list IS 'Listado detallado de workorders usando funciones SSOT de v3_core_ssot';
