-- ========================================
-- MIGRATION 000136: FIX INVESTOR REPORT RENT CALCULATION (DOWN)
-- ========================================
-- 
-- Purpose: Rollback de corrección de cálculo de arriendo
-- Date: 2025-10-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Eliminar función SSOT creada
DROP FUNCTION IF EXISTS v3_lot_ssot.rent_fixed_only_for_lot(bigint) CASCADE;

-- Restaurar vista original (de migración 000135)
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_project_base CASCADE;

CREATE OR REPLACE VIEW public.v3_report_investor_project_base AS
SELECT
  -- Identificación del proyecto
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  
  -- Superficie Total
  COALESCE(SUM(l.hectares), 0)::numeric AS surface_total_ha,
  
  -- Arriendo (versión original - usa rent_per_ha_for_lot que calcula TODOS los tipos)
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS lease_fixed_total_usd,
  
  -- Indicador si el arriendo es fijo
  CASE 
    WHEN COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0) > 0 
    THEN true 
    ELSE false 
  END AS lease_is_fixed,
  
  -- Arriendo por hectárea
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS lease_per_ha_usd,
  
  -- Administración y Estructura
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_total_usd,
  
  -- Administración por hectárea
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS admin_per_ha_usd

FROM public.projects p
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL

WHERE p.deleted_at IS NULL

GROUP BY 
  p.id, 
  p.name, 
  p.customer_id, 
  c.name, 
  p.campaign_id, 
  cam.name;

COMMENT ON VIEW public.v3_report_investor_project_base IS 
  'Vista 1/4 para informe de Aportes por Inversor. Contiene datos generales del proyecto: superficie, arriendo fijo y administración. Usa funciones SSOT para cálculos consistentes.';

