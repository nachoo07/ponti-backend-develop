-- ========================================
-- MIGRACIÓN 000124: CREATE v3_dashboard_operational_indicators VIEW (UP)
-- ========================================
-- 
-- Propósito: Vista de indicadores operativos del dashboard (SOLO ensamblaje con SSOT)
-- Dependencias: Requiere v3_core_ssot (000113), v3_dashboard_ssot (000116)
-- Arquitectura: Vista que SOLO ensambla, NO calcula (usa funciones SSOT)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY FASE 2:
-- - Vista ya usaba solo funciones SSOT (sin cálculos inline)
-- - Movida a migración separada para unificar criterio con otras vistas
-- - NO hay hardcodeos ni cálculos inline
-- 
-- Nota: Vistas SOLO ensamblan, NO calculan (usan funciones SSOT)

BEGIN;

-- ========================================
-- CREAR VISTA: v3_dashboard_operational_indicators
-- ========================================
-- Propósito: Fechas e indicadores operativos por proyecto

CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  
  -- Fechas operativas
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  
  -- IDs de workorders
  v3_dashboard_ssot.first_workorder_id_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_id_for_project(p.id) AS last_workorder_id,
  
  -- Fecha último arqueo
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
  
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_operational_indicators IS 'Módulo: Fechas e indicadores operativos por proyecto (SOLO ensamblaje SSOT)';

COMMIT;
