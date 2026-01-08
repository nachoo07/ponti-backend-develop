-- ========================================
-- MIGRATION 000089: FIX labor calculations v3 (DOWN)
-- ========================================
-- 
-- Purpose: Revert v3_labor_list to previous version without USD calculations
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- 1. REVERT v3_labor_list TO PREVIOUS VERSION
-- ========================================

DROP VIEW IF EXISTS public.v3_labor_list;

-- Recreate the previous version without usd_cost_ha and usd_net_total
CREATE OR REPLACE VIEW public.v3_labor_list AS
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
-- Esta migración revierte la vista v3_labor_list a su versión anterior
-- sin los campos usd_cost_ha y usd_net_total.
-- 
-- La vista vuelve a tener solo:
-- - cost_per_ha (precio en USD/ha)
-- - total_labor_cost (costo total en USD)
-- - dollar_average_month (promedio del dólar del mes)
-- 
-- Sin los campos:
-- - usd_cost_ha (costo en pesos/ha)
-- - usd_net_total (total neto en pesos)
-- ========================================
