-- ========================================
-- MIGRACIÓN 000125: FIX workorder number in operational indicators (UP)
-- ========================================
-- 
-- Propósito: Corregir funciones para retornar número visible de orden en lugar del ID
-- Bug: operational_indicators mostraba ID (4, 16) en lugar de número visible (1, 6)
-- Fix: Cambiar SELECT id → SELECT number::bigint en funciones SSOT
-- Fecha: 2025-10-05
-- Autor: Sistema

BEGIN;

-- ========================================
-- PASO 1: DROP vista que depende de las funciones
-- ========================================

DROP VIEW IF EXISTS public.v3_dashboard_operational_indicators;

-- ========================================
-- PASO 2: DROP funciones antiguas
-- ========================================

DROP FUNCTION IF EXISTS v3_dashboard_ssot.first_workorder_id_for_project(bigint);
DROP FUNCTION IF EXISTS v3_dashboard_ssot.last_workorder_id_for_project(bigint);

-- ========================================
-- PASO 3: CREATE funciones nuevas (retornan number en lugar de id)
-- ========================================

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.first_workorder_number_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT number::bigint  -- Convertir VARCHAR a bigint
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.first_workorder_number_for_project IS 'Retorna número visible de primera orden de trabajo del proyecto';

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.last_workorder_number_for_project(p_project_id bigint)
RETURNS bigint
LANGUAGE sql STABLE AS $$
  SELECT number::bigint  -- Convertir VARCHAR a bigint
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.last_workorder_number_for_project IS 'Retorna número visible de última orden de trabajo del proyecto';

-- ========================================
-- PASO 4: RECREAR vista usando las nuevas funciones
-- ========================================

CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  
  -- Fechas operativas
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  
  -- NÚMEROS de workorders (no IDs)
  v3_dashboard_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  
  -- Fecha último arqueo
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
  
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_operational_indicators IS 'Módulo: Fechas e indicadores operativos por proyecto (NÚMEROS visibles, no IDs)';

COMMIT;
