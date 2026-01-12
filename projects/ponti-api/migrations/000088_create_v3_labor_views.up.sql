-- ========================================
-- MIGRATION 000088: CREATE v3_labor_views (UP)
-- ========================================
-- 
-- Purpose: Create labor metrics and list views grouped by entity
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- =============================================================================
-- FUNCIÓN SSOT: Obtener promedio del dólar por mes específico
-- =============================================================================
-- Esta función es la Single Source of Truth para obtener el promedio del dólar
-- del mes específico en que se realizó cada labor. Cada labor tiene una fecha
-- (w.date) que corresponde a un mes determinado, y necesitamos el promedio
-- del dólar exacto de ese mes para cálculos precisos.
CREATE OR REPLACE FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT average_value 
     FROM public.project_dollar_values 
     WHERE project_id = p_project_id 
       AND month = TO_CHAR(p_date, 'YYYY-MM')  -- Convierte fecha a formato YYYY-MM
       AND deleted_at IS NULL
     LIMIT 1), 0  -- Si no hay datos para ese mes, retorna 0
  )::numeric
$$;

-- Eliminar vista existente si existe para evitar conflictos
DROP VIEW IF EXISTS public.v3_labor_list;

-- -------------------------------------------------------------------
-- VIEW: public.v3_labor_list
-- Objetivo (ES): Listado detallado de workorders con datos relevantes
-- Resultado (ES): Una fila por workorder con nombres (proyecto, campo,
--                 lote, cultivo), categoría de labor, precios y costos.
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_list AS
/* ---------------------------------------------------------------
   SELECT único (no usa CTE) porque es un listado denormalizado.
   Qué hace (ES):
   - Expone atributos de WO y sus relaciones (project, field, lot,
     crop, labor, categoría, investor).
   - Calcula el costo total de la labor por WO.
   - NUEVO: Incluye el promedio del dólar del mes específico de cada labor.
   Fórmulas (ES):
   - total_labor_cost (USD) = v3_calc.labor_cost(lb.price, w.effective_area)
                              (típicamente price_per_ha * area_ha)
   - dollar_average_month = v3_calc.dollar_average_for_month(w.project_id, w.date)
                           (promedio USD del mes específico de la labor)
   Filtros clave:
   - w.deleted_at IS NULL
   - w.effective_area > 0
   - lb.price IS NOT NULL
--------------------------------------------------------------- */
SELECT
  -- Identificación WO
  w.id                  AS workorder_id,       -- ID de la orden
  w.number              AS workorder_number,   -- Número visible de la orden
  w.date,                                      -- Fecha de ejecución

  -- Proyecto
  w.project_id,
  p.name                AS project_name,

  -- Campo
  w.field_id,
  f.name                AS field_name,

  -- Lote (opcional: puede no existir)
  w.lot_id,
  l.name                AS lot_name,

  -- Cultivo (opcional)
  w.crop_id,
  c.name                AS crop_name,

  -- Labor
  w.labor_id,
  lb.name               AS labor_name,

  -- Categoría de la labor (normalizada a categories)
  cat_lb.id             AS labor_category_id,
  cat_lb.name           AS labor_category_name,

  -- Proveedor/contratista
  w.contractor,
  lb.contractor_name,

  -- Superficie y costos
  w.effective_area      AS surface_ha,         -- Área trabajada (ha)
  lb.price              AS cost_per_ha,        -- Precio de labor (USD/ha)
  v3_calc.labor_cost(lb.price::numeric, w.effective_area::numeric) AS total_labor_cost, -- USD (price*area)
  
  -- NUEVO: Promedio del dólar del mes específico de la labor
  -- Cada labor se realiza en una fecha específica (w.date) que corresponde 
  -- a un mes determinado. Esta función trae el promedio del dólar exacto
  -- de ese mes específico desde project_dollar_values.
  v3_calc.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month, -- Promedio USD del mes

  -- Inversionista (opcional)
  w.investor_id,
  i.name                AS investor_name

FROM public.workorders w
JOIN public.projects   p  ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields     f  ON f.id = w.field_id   AND f.deleted_at IS NULL
LEFT JOIN public.lots  l  ON l.id = w.lot_id     AND l.deleted_at IS NULL
LEFT JOIN public.crops c  ON c.id = w.crop_id    AND c.deleted_at IS NULL
JOIN public.labors     lb ON lb.id = w.labor_id  AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat_lb ON cat_lb.id = lb.category_id AND cat_lb.deleted_at IS NULL
LEFT JOIN public.investors  i      ON i.id = w.investor_id        AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;