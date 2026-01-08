-- ========================================
-- MIGRACIÓN 000100: ADD SSOT FUNCTIONS FOR DASHBOARD (UP)
-- ========================================
-- 
-- Propósito: Crear funciones SSOT para dashboard (costos directos y resultado operativo)
-- Funciones: direct_costs_total_for_project + operating_result_total_for_project
-- Fórmula: SUM(direct_cost_usd) desde v3_workorder_metrics (sin movimientos internos)
-- Fecha: 2025-01-01
-- Autor: Sistema
-- 
-- Nota: Funciones reutilizables para dashboard en múltiples vistas

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA COSTOS DIRECTOS TOTALES
-- -------------------------------------------------------------------
-- Esta función calcula los costos directos totales usando v3_workorder_metrics (sin movimientos internos)
-- Fórmula: SUM(direct_cost_usd) desde v3_workorder_metrics
CREATE OR REPLACE FUNCTION v3_calc.direct_costs_total_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos: desde v3_workorder_metrics (ya corregidos, sin movimientos internos)
    (SELECT COALESCE(SUM(wm.direct_cost_usd), 0)::double precision
     FROM public.v3_workorder_metrics wm
     WHERE wm.project_id = p_project_id)
  , 0)::double precision
$$;

-- -------------------------------------------------------------------
-- FUNCIÓN SSOT PARA RESULTADO OPERATIVO TOTAL (CORREGIDA)
-- -------------------------------------------------------------------
-- Esta función calcula el resultado operativo total usando la misma lógica que el dashboard
-- Fórmula: Ingresos - Costos directos - Admin cost - Arriendo
CREATE OR REPLACE FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_calc.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados (usar función corregida)
    COALESCE(v3_calc.direct_costs_total_for_project(p_project_id), 0)::double precision
    -
    -- Costo administrativo total (admin_cost * total_hectares)
    (SELECT COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision
     FROM public.projects p
     LEFT JOIN (
       SELECT project_id, SUM(hectares) as total_hectares
       FROM public.v3_lot_metrics
       GROUP BY project_id
     ) ph ON ph.project_id = p.id
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
    -
    -- Arriendo total (lease_type_value * total_hectares)
    (SELECT COALESCE(f.lease_type_value * ph.total_hectares, 0)::double precision
     FROM public.fields f
     LEFT JOIN (
       SELECT project_id, SUM(hectares) as total_hectares
       FROM public.v3_lot_metrics
       GROUP BY project_id
     ) ph ON ph.project_id = f.project_id
     WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
     LIMIT 1)
  , 0)::double precision
$$;
