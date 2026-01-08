-- ========================================
-- MIGRACIÓN 000116: CREATE v3_dashboard_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Funciones SSOT específicas del dashboard (AMPLIADO por consolidación DRY fase 2)
-- Dependencias: Requiere v3_core_ssot (000113) y v3_lot_ssot (000115)
-- Alcance: 25 funciones (19 originales + 6 nuevas extraídas de vistas)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY FASE 1:
-- - Eliminadas: labor_cost_for_lot_mb (consolidada en v3_lot_ssot.labor_cost_for_lot)
-- - Eliminadas: supply_cost_seeds_for_lot_mb (ahora v3_lot_ssot.supply_cost_for_lot_by_category)
-- - Eliminadas: supply_cost_agrochemicals_for_lot_mb (ahora v3_lot_ssot.supply_cost_for_lot_by_category)
-- 
-- CONSOLIDACIÓN DRY FASE 2 (nuevas funciones extraídas):
-- - Agregadas GRUPO 7 (5 funciones): management_balance inline → SSOT functions
--   * supply_movements_invested_total_for_project
--   * lease_executed_for_project
--   * lease_invested_for_project
--   * admin_cost_total_for_project
--   * total_costs_for_project
-- - Agregadas GRUPO 8 (1 función): crop_incidence inline → SSOT function
--   * crop_incidence_for_project (table-returning)
-- 
-- Nota: Código en inglés, comentarios en español
-- Usa v3_core_ssot y v3_lot_ssot

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_dashboard_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_dashboard_ssot;

COMMENT ON SCHEMA v3_dashboard_ssot IS 'Funciones SSOT específicas del dashboard: cálculos por lote, proyecto, y agregaciones';

-- ========================================
-- NOTA: GRUPOS 1-6 (Funciones de lote) movidos a esquemas SSOT
-- ========================================
-- Funciones transversales de lote en v3_core_ssot (000113)
-- Funciones exclusivas de lote en v3_lot_ssot (000114)

-- ========================================
-- GRUPO 1: AGREGACIONES POR PROYECTO (4 funciones)
-- ========================================
-- Propósito: Calcular totales a nivel proyecto

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.total_hectares_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::double precision
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.direct_costs_total_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wm.direct_cost_usd), 0)::double precision
     FROM public.v3_workorder_metrics wm
     WHERE wm.project_id = p_project_id)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.operating_result_total_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::double precision as total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        CASE 
          WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * pt.total_hectares
          ELSE 0
        END, 
        0
      )::double precision as total_lease
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    CROSS JOIN project_totals pt
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    LIMIT 1
  )
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados (desde v3_workorder_metrics)
    v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
    -
    -- Arriendo total
    (SELECT total_lease FROM lease_cost)
    -
    -- Estructura (admin) total
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.total_invested_cost_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_lot_ssot.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Arriendo invertido
    (SELECT COALESCE(SUM(v3_lot_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Estructura invertida
    (SELECT COALESCE(SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

-- ========================================
-- GRUPO 2: MOVIMIENTOS INTERNOS (2 funciones)
-- ========================================
-- Propósito: Calcular costos de movimientos internos

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.supply_cost_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    -- Costos por movimientos internos de salida
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.supply_cost_received_for_project(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno entrada'
       AND sm.is_entry = true
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

-- ========================================
-- GRUPO 3: COSTOS POR CULTIVO (3 funciones)
-- ========================================
-- Propósito: Calcular costos agregados por cultivo
-- CORREGIDO: Usa v3_workorder_metrics (SSOT único) en lugar de v3_lot_ssot.direct_cost_for_lot

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.total_costs_for_crop(p_project_id bigint, p_crop_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wm.direct_cost_usd), 0)::double precision
     FROM v3_workorder_metrics wm
     JOIN public.lots l ON l.id = wm.lot_id
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.per_ha_dp(
    v3_dashboard_ssot.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.cost_per_ha_for_crop_ssot(p_project_id bigint, p_crop_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.cost_per_ha_for_crop(p_project_id, p_crop_id)
$$;

-- ========================================
-- GRUPO 4: MANAGEMENT BALANCE (4 funciones _mb) - REDUCIDO por consolidación DRY
-- ========================================
-- Propósito: Funciones específicas para balance de gestión
-- Nota: labor_cost_for_lot_mb, supply_cost_seeds_for_lot_mb, supply_cost_agrochemicals_for_lot_mb
--       eliminadas y consolidadas en v3_lot_ssot

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name = 'Semilla'
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.name != 'Semilla'
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.direct_costs_invested_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.stock_value_for_project_mb(p_project_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::double precision - v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
$$;

-- ========================================
-- GRUPO 5: INDICADORES OPERATIVOS (5 funciones)
-- ========================================
-- Propósito: Fechas e IDs de workorders para indicadores operativos

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.first_workorder_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MIN(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_workorder_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.first_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_workorder_id_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_stock_count_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(close_date)
  FROM public.stocks
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;

-- ========================================
-- GRUPO 6: FUNCIONES AUXILIARES (1 función)
-- ========================================
-- Propósito: Helpers para cálculos internos
-- Nota: direct_cost_usd movido a v3_core_ssot

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.direct_cost_per_ha_usd(
  p_labor_cost_usd numeric,
  p_supply_cost_usd numeric,
  p_sowed_area_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div(
    COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0),
    COALESCE(p_sowed_area_ha, 0)
  )
$$;

-- ========================================
-- GRUPO 7: FUNCIONES PARA MANAGEMENT BALANCE (5 funciones) - NUEVAS
-- ========================================
-- Propósito: Extraer cálculos inline de v3_dashboard_management_balance
-- Fecha: 2025-10-04 (Consolidación DRY fase 2)

-- 7.1: Total de supply movements invertidos
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::double precision
$$;

-- 7.2: Arriendo ejecutado (solo tipos 3 y 4)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_executed_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE((
    SELECT CASE 
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
      ELSE 0
    END
    FROM public.fields f
    WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
    LIMIT 1
  ), 0)::double precision
$$;

-- 7.3: Arriendo invertido (todos los tipos)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE((
    SELECT f.lease_type_value * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
    FROM public.fields f
    WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
    LIMIT 1
  ), 0)::double precision
$$;

-- 7.4: Admin cost total para el proyecto
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT p.admin_cost * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

-- 7.5: Total costs para el proyecto (directo + arriendo + admin)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.total_costs_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    v3_dashboard_ssot.direct_costs_total_for_project(p_project_id) + 
    v3_dashboard_ssot.lease_invested_for_project(p_project_id) + 
    v3_dashboard_ssot.admin_cost_total_for_project(p_project_id)
  , 0)::double precision
$$;

-- ========================================
-- GRUPO 8: FUNCIÓN PARA CROP INCIDENCE (1 función) - NUEVA
-- ========================================
-- Propósito: Encapsular lógica de ajuste de porcentajes por cultivo
-- Fecha: 2025-10-04 (Consolidación DRY fase 2)

-- 8.1: Incidencia de cultivos con porcentajes ajustados
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.crop_incidence_for_project(p_project_id bigint)
RETURNS TABLE(
  current_crop_id bigint,
  crop_name text,
  crop_hectares numeric,
  crop_incidence_pct numeric
)
LANGUAGE sql STABLE AS $$
  WITH lot_base AS (
    SELECT
      l.current_crop_id,
      c.name AS crop_name,
      l.hectares
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
    WHERE f.project_id = p_project_id 
      AND l.deleted_at IS NULL 
      AND l.hectares IS NOT NULL 
      AND l.hectares > 0
  ),
  project_totals AS (
    SELECT SUM(hectares)::numeric AS total_project_hectares
    FROM lot_base
  ),
  by_crop AS (
    SELECT 
      current_crop_id, 
      crop_name, 
      SUM(hectares)::numeric AS crop_hectares
    FROM lot_base
    WHERE current_crop_id IS NOT NULL
    GROUP BY current_crop_id, crop_name
  ),
  crop_percentages AS (
    SELECT
      bc.current_crop_id,
      bc.crop_name,
      bc.crop_hectares,
      pt.total_project_hectares,
      v3_core_ssot.percentage_rounded(bc.crop_hectares, pt.total_project_hectares) AS base_percentage,
      ROW_NUMBER() OVER (ORDER BY bc.crop_name) AS crop_order,
      COUNT(*) OVER () AS total_crops
    FROM by_crop bc
    CROSS JOIN project_totals pt
  ),
  project_sums AS (
    SELECT SUM(base_percentage) AS total_percentage
    FROM crop_percentages
  )
  SELECT
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    CASE 
      WHEN ps.total_percentage > 99.000 AND cp.crop_order = cp.total_crops THEN
        100.000 - COALESCE((
          SELECT SUM(base_percentage) 
          FROM crop_percentages cp2 
          WHERE cp2.crop_order < cp.crop_order
        ), 0)
      ELSE
        cp.base_percentage
    END AS crop_incidence_pct
  FROM crop_percentages cp
  CROSS JOIN project_sums ps
  ORDER BY cp.crop_name
$$;

COMMIT;

-- Comentarios sobre funciones clave
COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 'Calcula resultado operativo total del proyecto';
COMMENT ON FUNCTION v3_dashboard_ssot.first_workorder_date_for_project(bigint) IS 'Obtiene fecha de la primera orden de trabajo del proyecto';
COMMENT ON FUNCTION v3_dashboard_ssot.last_stock_count_date_for_project(bigint) IS 'Obtiene fecha del último arqueo de stock del proyecto';

