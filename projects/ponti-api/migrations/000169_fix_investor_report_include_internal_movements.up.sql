-- ========================================
-- MIGRACIÓN 000169: FIX Informe de Aportes - Quitar filtro quantity > 0 (UP)
-- ========================================
-- 
-- Propósito: Sincronizar Informe de Aportes con Dashboard para incluir cantidades negativas
-- Problema: Control 7 falla después de movimientos internos
--           Dashboard: 11,113.14 USD (incluye cantidades negativas)
--           Aportes:   12,287.14 USD (excluye cantidades negativas con quantity > 0)
--           Diferencia: -1,174.00 USD
-- Solución: Quitar el filtro "sm.quantity > 0" de las categorías de insumos
--           para permitir cantidades negativas de movimientos internos de salida
-- Fecha: 2025-10-31
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- ACTUALIZAR: v3_report_investor_contribution_categories
-- ========================================
-- La vista ya incluye movimientos internos (migración 000168)
-- Solo necesitamos QUITAR el filtro "sm.quantity > 0"

CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  SELECT 
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM workorders w
      JOIN labors lab ON w.labor_id = lab.id
      JOIN categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0) AS seeded_area_ha
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
)
SELECT 
  lb.project_id,
  
  -- AGROQUÍMICOS - QUITADO: sm.quantity > 0
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      -- QUITADO: AND sm.quantity > 0
      AND cat.type_id = 2
      AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  -- FERTILIZANTES - QUITADO: sm.quantity > 0
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      -- QUITADO: AND sm.quantity > 0
      AND cat.type_id = 3
      AND cat.name = 'Fertilizantes'
  ), 0)::numeric AS fertilizers_total_usd,
  
  -- SEMILLAS - QUITADO: sm.quantity > 0
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      -- QUITADO: AND sm.quantity > 0
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0)::numeric AS seeds_total_usd,
  
  -- LABORES GENERALES (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name IN ('Pulverización', 'Otras Labores')
  ), 0)::numeric AS general_labors_total_usd,
  
  -- SIEMBRA (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Siembra'
  ), 0)::numeric AS sowing_total_usd,
  
  -- RIEGO (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Riego'
  ), 0)::numeric AS irrigation_total_usd,
  
  -- ARRIENDO CAPITALIZABLE (sin cambios)
  COALESCE((
    SELECT f.lease_type_value * SUM(lb2.seeded_area_ha)
    FROM public.fields f
    JOIN lot_base lb2 ON lb2.project_id = f.project_id
    WHERE f.project_id = lb.project_id
      AND f.deleted_at IS NULL
      AND f.lease_type_id IN (1, 3, 4)
    GROUP BY f.lease_type_value
    LIMIT 1
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- ADMINISTRACIÓN Y ESTRUCTURA (sin cambios)
  COALESCE((
    SELECT p.admin_cost * SUM(lb2.seeded_area_ha)
    FROM public.projects p
    JOIN lot_base lb2 ON lb2.project_id = p.id
    WHERE p.id = lb.project_id
      AND p.deleted_at IS NULL
    GROUP BY p.admin_cost
  ), 0)::numeric AS administration_total_usd,
  
  -- SUPERFICIE SEMBRADA TOTAL (sin cambios)
  COALESCE((
    SELECT SUM(seeded_area_ha)
    FROM lot_base
    WHERE project_id = lb.project_id
  ), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb
GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
'Categorías totales de aportes por proyecto. Sincronizado con Dashboard: incluye movimientos internos con cantidades negativas.';

COMMIT;

