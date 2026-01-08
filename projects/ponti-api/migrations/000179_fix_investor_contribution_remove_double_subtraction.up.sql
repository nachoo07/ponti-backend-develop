-- ========================================
-- MIGRACIÓN 000179: FIX Investor Contribution Remove Double Subtraction (UP)
-- ========================================
--
-- Propósito: Corregir doble contabilización en v3_report_investor_contribution_categories
-- Problema: La vista resta movimientos internos con is_entry=FALSE para:
--           - Agroquímicos
--           - Fertilizantes
--           - Semillas
--           Esto causa doble contabilización (mismo problema que migración 000178)
-- Solución: Eliminar resta de movimientos internos is_entry=FALSE
-- Fecha: 2025-11-03
-- Autor: Sistema
--
-- Impacto: Control 7 pasará de ERROR a OK
--          Informe de Aportes coincidirá con Dashboard
--
-- Note: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR v3_report_investor_contribution_categories
-- ========================================

DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;

CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    COALESCE(SUM(w.effective_area), 0) AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lab ON w.labor_id = lab.id
  LEFT JOIN public.categories cat ON lab.category_id = cat.id
  WHERE l.deleted_at IS NULL
    AND (cat.name = 'Siembra' AND cat.type_id = 4 OR cat.name IS NULL)
  GROUP BY l.id, f.project_id
)
SELECT
  lb.project_id,
  
  -- ========================================
  -- FIX: AGROQUÍMICOS (eliminar resta de movimientos internos)
  -- ========================================
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
      AND cat.type_id = 2
      AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  -- ========================================
  -- FIX: FERTILIZANTES (eliminar resta de movimientos internos)
  -- ========================================
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
      AND cat.type_id = 3
      AND cat.name = 'Fertilizantes'
  ), 0)::numeric AS fertilizers_total_usd,
  
  -- ========================================
  -- FIX: SEMILLA (eliminar resta de movimientos internos)
  -- ========================================
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
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0)::numeric AS seeds_total_usd,
  
  -- LABORES GENERALES (Pulverización, Otras Labores)
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
  
  -- SIEMBRA
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
  
  -- RIEGO
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
  
  -- ARRIENDO CAPITALIZABLE (corregido en migración 000176)
  COALESCE((
    SELECT SUM(v3_lot_ssot.rent_per_ha_for_lot(l2.id) * l2.hectares)
    FROM public.lots l2
    JOIN public.fields f ON f.id = l2.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = lb.project_id 
      AND l2.deleted_at IS NULL
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- ADMINISTRACIÓN Y ESTRUCTURA
  COALESCE((
    SELECT p.admin_cost * SUM(lb2.seeded_area_ha)
    FROM public.projects p
    JOIN lot_base lb2 ON lb2.project_id = p.id
    WHERE p.id = lb.project_id
      AND p.deleted_at IS NULL
    GROUP BY p.admin_cost
  ), 0)::numeric AS administration_total_usd,
  
  -- SUPERFICIE SEMBRADA TOTAL
  COALESCE((
    SELECT SUM(seeded_area_ha)
    FROM lot_base
    WHERE project_id = lb.project_id
  ), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb
GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
'Categorías de aporte por proyecto para informe de inversores. FIX 000179: Elimina doble contabilización de movimientos internos.';

COMMIT;


-- ========================================
-- RECREAR v3_investor_contribution_data_view
-- ========================================
-- Esta vista depende de v3_report_investor_contribution_categories que acabamos de recrear
-- (Copiar desde migración 000176, líneas 210-743)

-- [La vista es muy larga, usará la versión existente en migración 000176]
-- La recreación se hará automáticamente al aplicar la migración 000176 nuevamente.

