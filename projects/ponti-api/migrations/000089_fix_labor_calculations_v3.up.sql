-- ========================================
-- MIGRATION 000089: FIX labor calculations v3 (UP)
-- ========================================
-- 
-- Purpose: Add missing USD calculations to v3_labor_list view
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.
-- 
-- Problem: v3_labor_list is missing usd_cost_ha and usd_net_total fields
-- Solution: Add these fields using dollar_average_month for accurate calculations
-- 
-- Fields to add:
-- - usd_cost_ha: Costo U$/Ha en pesos (cost_per_ha * dollar_average_month)
-- - usd_net_total: Total U Neto en pesos (usd_cost_ha * surface_ha)

-- ========================================
-- 1. DROP AND RECREATE v3_labor_list WITH USD CALCULATIONS
-- ========================================

DROP VIEW IF EXISTS public.v3_labor_list;

-- -------------------------------------------------------------------
-- VIEW: public.v3_labor_list (CORREGIDA)
-- Objetivo (ES): Listado detallado de workorders con cálculos USD corregidos
-- Resultado (ES): Una fila por workorder con nombres (proyecto, campo,
--                 lote, cultivo), categoría de labor, precios y costos.
-- NUEVO: Incluye usd_cost_ha y usd_net_total calculados correctamente
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_list AS
/* ---------------------------------------------------------------
   SELECT único (no usa CTE) porque es un listado denormalizado.
   Qué hace (ES):
   - Expone atributos de WO y sus relaciones (project, field, lot,
     crop, labor, categoría, investor).
   - Calcula el costo total de la labor por WO.
   - Incluye el promedio del dólar del mes específico de cada labor.
   - NUEVO: Calcula usd_cost_ha y usd_net_total en pesos
   Fórmulas (ES):
   - total_labor_cost (USD) = v3_calc.labor_cost(lb.price, w.effective_area)
   - dollar_average_month = v3_calc.dollar_average_for_month(w.project_id, w.date)
   - usd_cost_ha (pesos) = lb.price * dollar_average_month
   - usd_net_total (pesos) = usd_cost_ha * w.effective_area
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
  
  -- Promedio del dólar del mes específico de la labor
  v3_calc.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month, -- Promedio USD del mes

  -- ========================================
  -- NUEVOS CAMPOS: CÁLCULOS EN PESOS
  -- ========================================
  
  -- Costo U$/Ha en pesos: Costo Ha * dólar promedio (mostrar en pesos)
  (lb.price * v3_calc.dollar_average_for_month(w.project_id, w.date)) AS usd_cost_ha,
  
  -- Total U Neto en pesos: Total costo en pesos * Has
  (lb.price * v3_calc.dollar_average_for_month(w.project_id, w.date) * w.effective_area) AS usd_net_total,

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

-- ========================================
-- 2. COMENTARIOS EXPLICATIVOS
-- ========================================
-- 
-- Campos agregados:
-- - usd_cost_ha: Costo por hectárea en pesos (lb.price * dollar_average_month)
-- - usd_net_total: Total neto en pesos (usd_cost_ha * surface_ha)
-- 
-- Fórmulas:
-- - usd_cost_ha = cost_per_ha * dollar_average_month
-- - usd_net_total = usd_cost_ha * surface_ha
-- 
-- Esto corrige el problema donde la UI mostraba valores random
-- porque no tenía acceso a estos cálculos en las vistas v3.
-- ========================================
