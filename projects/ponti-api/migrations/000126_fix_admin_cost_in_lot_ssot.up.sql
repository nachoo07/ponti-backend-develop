-- ========================================
-- MIGRACIÓN 000126: FIX admin_cost en v3_lot_ssot (UP)
-- ========================================
-- 
-- Propósito: Corregir admin_cost_per_ha_for_lot para usar project.admin_cost directamente
-- Referencia: Migración 000083 estableció que admin_cost debe usarse sin cálculos
-- Fecha: 2025-10-07
-- Autor: Sistema
-- 
-- Problema: La función v3_lot_ssot.admin_cost_per_ha_for_lot dividía admin_cost entre hectáreas
-- Solución: Retornar project.admin_cost directamente tal como está en la entidad proyecto
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- PASO 1: CORREGIR FUNCIONES SSOT
-- ========================================

-- 1.1: admin_cost_per_ha_for_lot (retorna project.admin_cost tal cual)
CREATE OR REPLACE FUNCTION v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Retorna admin_cost del proyecto tal cual
  SELECT COALESCE(p.admin_cost, 0)::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_lot_ssot.admin_cost_per_ha_for_lot(bigint) IS 
'Retorna el admin_cost del proyecto tal cual. Valor fijo de la entidad proyecto.';

-- 1.2: operating_result_total_for_project (mantener multiplicación por hectáreas para Dashboard)
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
    -- Estructura (admin) total = admin_cost × total_hectares
    (SELECT COALESCE(admin_cost * total_hectares, 0)::double precision FROM project_totals)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.operating_result_total_for_project(bigint) IS 
'Calcula el resultado operativo total del proyecto. Estructura = admin_cost × hectáreas.';

-- 1.3: admin_cost_total_for_project (mantener multiplicación por hectáreas para Dashboard)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Retorna admin_cost × total_hectares para el Dashboard (costo total del proyecto)
  SELECT COALESCE(
    (SELECT p.admin_cost * v3_dashboard_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.admin_cost_total_for_project(bigint) IS 
'Retorna admin_cost × hectáreas del proyecto (costo total de estructura para Dashboard).';

-- ========================================
-- PASO 2: ACTUALIZAR VISTA v3_lot_metrics
-- ========================================
-- admin_cost_per_ha_usd = project.admin_cost (tal cual)
-- admin_total_usd = project.admin_cost (tal cual)

DROP VIEW IF EXISTS public.v3_lot_metrics CASCADE;

CREATE OR REPLACE VIEW public.v3_lot_metrics AS
/* --------------------------------------------------------------------
   CTE: base - Datos básicos del lote
   - Calcula áreas sembradas y cosechadas desde workorders
   - sowed_area_ha = suma de effective_area de workorders de siembra (category_id = 9)
   - harvested_area_ha = suma de effective_area de workorders de cosecha (category_id = 13)
-------------------------------------------------------------------- */
WITH base AS (
  SELECT
    f.project_id,
    l.id              AS lot_id,
    l.name            AS lot_name,
    l.hectares,
    l.tons,
    l.sowing_date,
    -- Área sembrada desde workorders de siembra
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0)::numeric AS sowed_area_ha,
    -- Área cosechada desde workorders de cosecha
    COALESCE(SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END), 0)::numeric AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
  GROUP BY f.project_id, l.id, l.name, l.hectares, l.tons, l.sowing_date
),
/* --------------------------------------------------------------------
   CTE: workorder_costs - Costos desde v3_workorder_metrics (SSOT único)
   - ELIMINA cálculos duplicados de v3_lot_ssot
   - USA directamente v3_workorder_metrics que ya calcula todo correctamente
-------------------------------------------------------------------- */
workorder_costs AS (
  SELECT
    lot_id,
    COALESCE(labor_cost_usd, 0)::numeric AS labor_cost_usd,
    COALESCE(supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
    COALESCE(direct_cost_usd, 0)::numeric AS direct_cost_usd
  FROM v3_workorder_metrics
),
/* --------------------------------------------------------------------
   CTE: project_total_direct_cost
   - Calcula el costo directo total del proyecto desde v3_workorder_metrics
   - ÚNICA fuente de verdad para costos directos
-------------------------------------------------------------------- */
project_total_direct_cost AS (
  SELECT
    p.id AS project_id,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares,
    COALESCE(SUM(wc.direct_cost_usd), 0)::numeric AS total_direct_cost
  FROM public.projects p
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorder_costs wc ON wc.lot_id = l.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
)
/* --------------------------------------------------------------------
   SELECT PRINCIPAL: Ensamblado final de todas las métricas
   - Combina datos de base, workorder_costs, y cálculos adicionales
-------------------------------------------------------------------- */
SELECT
  -- ############# IDENTIFICADORES ##############
  b.project_id,
  b.lot_id,
  b.lot_name,
  b.hectares,

  -- ############# ÁREAS ##############
  b.sowed_area_ha,
  b.harvested_area_ha,

  -- ############# RENDIMIENTO ##############
  v3_lot_ssot.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
  b.tons,

  -- ############# FECHAS ##############
  b.sowing_date,

  -- ############# COSTOS DIRECTOS (desde v3_workorder_metrics) ##############
  COALESCE(wc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(wc.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  COALESCE(wc.direct_cost_usd, 0)::numeric AS direct_cost_usd,

  -- ############# INGRESOS (desde v3_lot_ssot) ##############
  COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric AS income_net_total_usd,

  -- ############# COSTOS POR HECTÁREA ##############
  -- Ingreso neto por ha
  v3_core_ssot.per_ha(
    COALESCE(v3_lot_ssot.income_net_total_for_lot(b.lot_id), 0)::numeric,
    b.hectares::numeric
  ) AS income_net_per_ha_usd,
  
  -- Arriendo por ha
  COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0)::numeric AS rent_per_ha_usd,
  
  -- Admin por ha (retorna project.admin_cost tal cual)
  COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_cost_per_ha_usd,
  
  -- Activo total por ha
  COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0)::numeric AS active_total_per_ha_usd,
  
  -- Resultado operativo por ha
  COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0)::numeric AS operating_result_per_ha_usd,

  -- ############# TOTALES ##############
  -- Arriendo total
  (COALESCE(v3_lot_ssot.rent_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS rent_total_usd,
  
  -- Admin total (retorna project.admin_cost tal cual)
  COALESCE(v3_lot_ssot.admin_cost_per_ha_for_lot(b.lot_id), 0)::numeric AS admin_total_usd,
  
  -- Activo total
  (COALESCE(v3_lot_ssot.active_total_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS active_total_usd,
  
  -- Resultado operativo total
  (COALESCE(v3_lot_ssot.operating_result_per_ha_for_lot(b.lot_id), 0) * b.hectares::numeric)::numeric AS operating_result_total_usd,

  -- ############# COSTO DIRECTO POR HA (PROMEDIO DEL PROYECTO) ##############
  -- Fórmula: total_direct_cost_proyecto / hectareas_proyecto
  -- Usa v3_workorder_metrics como ÚNICA fuente
  v3_core_ssot.cost_per_ha(
    COALESCE(ptdc.total_direct_cost, 0)::numeric,
    COALESCE(ptdc.total_hectares, 0)::numeric
  ) AS direct_cost_per_ha_usd,

  -- ############# SUPERFICIE TOTAL DEL PROYECTO ##############
  COALESCE(ptdc.total_hectares, 0)::numeric AS project_total_hectares

FROM base b
LEFT JOIN workorder_costs wc ON wc.lot_id = b.lot_id
LEFT JOIN project_total_direct_cost ptdc ON ptdc.project_id = b.project_id;

COMMENT ON VIEW public.v3_lot_metrics IS 'Métricas agregadas por lote usando v3_workorder_metrics como SSOT de costos directos';

-- ========================================
-- PASO 3: RECREAR VISTA v3_lot_list
-- ========================================
-- Esta vista depende de v3_lot_metrics, por eso fue eliminada por CASCADE
CREATE OR REPLACE VIEW public.v3_lot_list AS
WITH base AS (
  SELECT
    f.project_id,
    p.name AS project_name,
    f.id AS field_id,
    f.name AS field_name,
    l.id AS lot_id,
    l.name AS lot_name,
    l.variety,
    l.season,
    l.previous_crop_id,
    prev_crop.name AS previous_crop,
    l.current_crop_id,
    curr_crop.name AS current_crop,
    l.hectares,
    l.updated_at,
    l.sowing_date,
    l.tons
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops prev_crop ON prev_crop.id = l.previous_crop_id AND prev_crop.deleted_at IS NULL
  LEFT JOIN public.crops curr_crop ON curr_crop.id = l.current_crop_id AND curr_crop.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
wo_dates AS (
  SELECT
    w.lot_id,
    MIN(w.date) AS raw_sowing_date
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
  GROUP BY w.lot_id
),
lot_metrics_data AS (
  -- Obtener métricas desde v3_lot_metrics (áreas, costos, etc)
  SELECT
    project_id,
    lot_id,
    sowed_area_ha,
    harvested_area_ha,
    yield_tn_per_ha,
    direct_cost_per_ha_usd,
    direct_cost_usd,
    income_net_total_usd,
    income_net_per_ha_usd,
    rent_per_ha_usd,
    admin_cost_per_ha_usd,
    active_total_per_ha_usd,
    operating_result_per_ha_usd,
    rent_total_usd,
    admin_total_usd,
    active_total_usd,
    operating_result_total_usd
  FROM v3_lot_metrics
)
SELECT
  b.project_id,
  b.project_name,
  b.field_id,
  b.field_name,
  b.lot_id AS id,
  b.lot_name,
  b.variety,
  b.season,
  b.previous_crop_id,
  b.previous_crop,
  b.current_crop_id,
  b.current_crop,
  b.hectares,
  b.updated_at,
  
  -- Áreas (desde v3_lot_metrics)
  lm.sowed_area_ha,
  lm.harvested_area_ha,
  
  -- Rendimiento (desde v3_lot_metrics)
  lm.yield_tn_per_ha,
  
  -- Costo por ha (por cultivo, usa v3_dashboard_ssot para coincidir con Dashboard)
  v3_dashboard_ssot.cost_per_ha_for_crop_ssot(b.project_id, b.current_crop_id)::numeric AS cost_usd_per_ha,
  
  -- Ingresos y otros costos por ha (desde v3_lot_metrics)
  lm.income_net_per_ha_usd,
  lm.rent_per_ha_usd,
  lm.admin_cost_per_ha_usd,
  lm.active_total_per_ha_usd,
  lm.operating_result_per_ha_usd,
  
  -- Totales por lote (desde v3_lot_metrics)
  lm.income_net_total_usd,
  lm.direct_cost_usd AS direct_cost_total_usd,
  lm.rent_total_usd,
  lm.admin_total_usd,
  lm.active_total_usd,
  lm.operating_result_total_usd,
  
  -- Fechas
  b.sowing_date AS lot_sowing_date,
  NULL::date AS lot_harvest_date,
  b.tons,
  wd.raw_sowing_date
  
FROM base b
LEFT JOIN wo_dates wd ON wd.lot_id = b.lot_id
LEFT JOIN lot_metrics_data lm ON lm.lot_id = b.lot_id;

COMMENT ON VIEW public.v3_lot_list IS 'Listado de lotes - cost_usd_per_ha calculado por cultivo (coincide con Dashboard)';

-- ========================================
-- PASO 4: RECREAR VISTA v3_dashboard_metrics
-- ========================================
-- Esta vista también depende de v3_lot_metrics, fue eliminada por CASCADE
-- FIX: Corregir admin_cost (debe usarse tal cual, sin multiplicar por hectáreas)
CREATE OR REPLACE VIEW public.v3_dashboard_metrics AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    hectares,
    sowed_area_ha,
    harvested_area_ha,
    direct_cost_per_ha_usd,
    lot_id
  FROM public.v3_lot_metrics
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- ========================================
  -- CARD 1: AVANCE DE SIEMBRA
  -- ========================================
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- ========================================
  -- CARD 2: AVANCE DE COSECHA
  -- ========================================
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- ========================================
  -- CARD 3: AVANCE DE COSTOS
  -- ========================================
  -- Costo ejecutado (promedio ponderado por ha sembrada)
  COALESCE(
    SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
    0
  )::double precision AS executed_costs_usd,
  
  -- ⚠️ PRESUPUESTO HARDCODEADO TEMPORAL ⚠️
  -- TODO: Definir fórmula correcta para calcular el presupuesto dinámicamente
  -- Por ahora: admin_cost * 10 (valor temporal hasta definir cálculo)
  (p.admin_cost * 10)::double precision AS budget_cost_usd,
  
  -- Porcentaje de progreso
  v3_core_ssot.percentage(
    COALESCE(
      SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
      0
    )::numeric,
    (p.admin_cost * 10)::numeric
  ) AS costs_progress_pct,
  
  -- ========================================
  -- CARD 4: RESULTADO OPERATIVO
  -- ========================================
  -- Ingresos (suma de ingresos netos por lote)
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(lm.lot_id)), 0) AS operating_result_income_usd,
  
  -- Resultado operativo
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  
  -- Total costos (usa función SSOT, NO repetir cálculo)
  v3_dashboard_ssot.total_costs_for_project(p.id) AS operating_result_total_costs_usd,
  
  -- Porcentaje de rentabilidad (usa función SSOT para costos)
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  
  -- ========================================
  -- CAMPOS ADICIONALES (para compatibilidad)
  -- ========================================
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_metrics IS 'Métricas consolidadas del dashboard (5 cards en 1 vista) - admin_cost usado tal cual';

COMMIT;
