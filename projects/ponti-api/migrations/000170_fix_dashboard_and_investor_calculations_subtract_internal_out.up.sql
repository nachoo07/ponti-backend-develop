-- ========================================
-- MIGRACIÓN 000170: FIX Dashboard e Informe de Aportes - Restar movimientos internos de salida (UP)
-- ========================================
-- 
-- Propósito: Corregir cálculos para restar movimientos internos de salida (is_entry=FALSE)
-- Problema: Los movimientos internos con is_entry=FALSE NO se restan del proyecto origen
--           Ejemplo: Proyecto 11 transfiere 16 unidades (128 USD) pero no se resta
-- Solución: Modificar funciones para incluir movimientos internos con is_entry=FALSE como RESTA
-- Fecha: 2025-10-31
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- FIX 1: seeds_invested_for_project_mb (Dashboard)
-- ========================================

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Sumar entradas (is_entry=TRUE)
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name = 'Semilla'
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
    -
    -- Restar salidas de movimientos internos (is_entry=FALSE)
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name = 'Semilla'
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb IS 
'Calcula semillas invertidas para un proyecto (incluye entradas y resta salidas de movimientos internos). Sincronizado con Informe de Aportes.';

-- ========================================
-- FIX 2: agrochemicals_invested_for_project_mb (Dashboard)
-- ========================================

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Sumar entradas (is_entry=TRUE)
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
    -
    -- Restar salidas de movimientos internos (is_entry=FALSE)
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb IS 
'Calcula agroquímicos invertidos para un proyecto (incluye entradas y resta salidas de movimientos internos). Sincronizado con Informe de Aportes.';

-- ========================================
-- FIX 3: supply_movements_invested_total_for_project (Dashboard)
-- ========================================

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Sumar entradas (is_entry=TRUE)
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
    -
    -- Restar salidas de movimientos internos (is_entry=FALSE)
    COALESCE((SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = FALSE
       AND sm.movement_type = 'Movimiento interno'), 0)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project IS 
'Calcula total de movimientos de insumos invertidos para un proyecto (incluye entradas y resta salidas de movimientos internos). Usado en costos_directos_invertidos_usd.';

-- ========================================
-- FIX 4: v3_report_investor_contribution_categories (Informe de Aportes)
-- ========================================

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
  
  -- AGROQUÍMICOS - Incluir entradas y restar salidas
  (COALESCE((
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
  ), 0) - COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type = 'Movimiento interno'
      AND sm.is_entry = FALSE
      AND s.price IS NOT NULL
      AND cat.type_id = 2
      AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  ), 0))::numeric AS agrochemicals_total_usd,
  
  -- FERTILIZANTES - Incluir entradas y restar salidas
  (COALESCE((
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
  ), 0) - COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type = 'Movimiento interno'
      AND sm.is_entry = FALSE
      AND s.price IS NOT NULL
      AND cat.type_id = 3
      AND cat.name = 'Fertilizantes'
  ), 0))::numeric AS fertilizers_total_usd,
  
  -- SEMILLAS - Incluir entradas y restar salidas
  (COALESCE((
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
  ), 0) - COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type = 'Movimiento interno'
      AND sm.is_entry = FALSE
      AND s.price IS NOT NULL
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0))::numeric AS seeds_total_usd,
  
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
'Categorías totales de aportes por proyecto. CORREGIDO: Resta movimientos internos de salida (is_entry=FALSE).';

COMMIT;

