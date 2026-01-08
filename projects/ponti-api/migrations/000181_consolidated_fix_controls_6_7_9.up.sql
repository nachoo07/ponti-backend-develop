-- ========================================
-- MIGRACIÓN 000181: CONSOLIDACIÓN DE FIXES - Controles 6, 7, 9 (UP)
-- ========================================
--
-- Propósito: Consolidar todas las correcciones de las migraciones 000175-000179
--            en una sola migración para deploy en GCP
--
-- Problemas corregidos:
--   1. Control 6: Agroquímicos ($33,092 → $43,377)
--   2. Control 7: Arriendo ($118,900 → $162,691)
--   3. Control 9: Labores Pre-cosecha ($146,207 → $67,750)
--
-- Causa raíz:
--   - Control 6: Doble contabilización negativa en movimientos internos
--   - Control 7: LIMIT 1 en vez de sumar todos los lotes + tipo 4 mal calculado
--   - Control 9: Labores invertidos incluía cosecha (no es inversión inicial)
--
-- Solución:
--   - Crear funciones SSOT para arriendo por lote
--   - Eliminar doble resta de movimientos internos
--   - Separar labores pre-cosecha de cosecha
--   - Recrear vistas con cálculos correctos
--
-- Fecha: 2024-11-04
-- Autor: Sistema
-- Referencia: Migraciones 000175, 000176, 000177, 000178, 000179
--

BEGIN;

-- ========================================
-- PARTE 1: FIX LEASE FUNCTIONS (Control 7)
-- ========================================
-- Origen: Migración 000175

-- Función para calcular arriendo invertido (suma todos los lotes)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(v3_lot_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.lease_invested_for_project IS 
'Calcula arriendo invertido para un proyecto sumando todos los lotes. FIX 000181: Usa rent_per_ha_for_lot() por cada lote.';

-- Función para calcular arriendo ejecutado (solo tipos 3 y 4)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.lease_executed_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(
    CASE
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * l.hectares
      ELSE 0
    END
  ), 0)::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.lease_executed_for_project IS 
'Calcula arriendo ejecutado para un proyecto (solo tipos 3 y 4 con componente fijo). FIX 000181: Suma todos los lotes.';

-- ========================================
-- PARTE 2: FIX DASHBOARD LABORES (Control 9)
-- ========================================
-- Origen: Migración 000177

-- Función para labores PRE-COSECHA (excluye Cosecha)
CREATE OR REPLACE FUNCTION v3_lot_ssot.labor_cost_pre_harvest_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
    AND cat.type_id = 4
    AND cat.name != 'Cosecha'  -- EXCLUIR COSECHA
$$;

COMMENT ON FUNCTION v3_lot_ssot.labor_cost_pre_harvest_for_lot IS 
'Suma labores PRE-COSECHA (excluye Cosecha) para un lote. FIX 000181: Control 9.';

-- ========================================
-- PARTE 3: FIX AGROCHEMICALS FUNCTIONS (Control 6)
-- ========================================
-- Origen: Migración 000178

-- Corregir función de agroquímicos (eliminar doble resta)
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
       AND c.type_id = 2
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.agrochemicals_invested_for_project_mb IS 
'Calcula agroquímicos invertidos (incluye movimientos internos con is_entry=TRUE). FIX 000181: Elimina doble contabilización.';

-- Corregir función de semillas
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
       AND c.type_id = 1
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.seeds_invested_for_project_mb IS 
'Calcula semillas invertidas (incluye movimientos internos con is_entry=TRUE). FIX 000181: Elimina doble contabilización.';

-- Corregir función general de supply movements
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
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::double precision
$$;

COMMENT ON FUNCTION v3_dashboard_ssot.supply_movements_invested_total_for_project IS 
'Calcula total de movimientos invertidos (incluye movimientos internos con is_entry=TRUE). FIX 000181: Elimina doble contabilización.';

-- ========================================
-- PARTE 4: RECREAR DASHBOARD MANAGEMENT BALANCE
-- ========================================
-- Origen: Migraciones 000177 y 000178

DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
SELECT
  p.id AS project_id,
  
  -- Ingresos y Resultado
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  
  -- Costos Directos
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) + 
   COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) AS costos_directos_invertidos_usd,
  ((v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) + 
    COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) - 
   v3_dashboard_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
  
  -- Semillas
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) - 
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)) AS semillas_stock_usd,
  
  -- Agroquímicos
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) - 
   COALESCE(SUM(
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
   ), 0)) AS agroquimicos_stock_usd,
  
  -- Fertilizantes
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_ejecutados_usd,
  (SELECT COALESCE(SUM(sm.quantity * s.price), 0)
   FROM public.supply_movements sm
   JOIN public.supplies s ON s.id = sm.supply_id
   JOIN public.categories c ON s.category_id = c.id
   WHERE sm.project_id = p.id
     AND sm.deleted_at IS NULL
     AND s.deleted_at IS NULL
     AND sm.is_entry = TRUE
     AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
     AND c.type_id = 3) AS fertilizantes_invertidos_usd,
  ((SELECT COALESCE(SUM(sm.quantity * s.price), 0)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id
    JOIN public.categories c ON s.category_id = c.id
    WHERE sm.project_id = p.id
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.is_entry = TRUE
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND c.type_id = 3) - 
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0)) AS fertilizantes_stock_usd,
  
  -- Labores (FIX: Invertidos excluye cosecha)
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0) AS labores_invertidos_usd,
  
  -- Arriendo (FIX: Usa funciones SSOT que suman todos los lotes)
  v3_dashboard_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
  v3_dashboard_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
  
  -- Estructura
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  
  -- Costos calculados (para compatibilidad)
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semilla_cost,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS insumos_cost,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_cost
  
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

COMMENT ON VIEW public.v3_dashboard_management_balance IS 
'Balance de gestión con ejecutados/invertidos/stock. FIX 000181: Controles 6, 7, 9 corregidos.';

-- ========================================
-- PARTE 5: RECREAR INVESTOR CONTRIBUTION CATEGORIES (Control 6, 7)
-- ========================================
-- Origen: Migraciones 000176 y 000179

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
  SUM(lb.seeded_area_ha)::numeric AS total_seeded_area_ha,
  
  -- AGROQUÍMICOS (FIX: solo suma entradas, sin doble resta)
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
  
  -- FERTILIZANTES (FIX: solo suma entradas, sin doble resta)
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
  
  -- SEMILLA (FIX: solo suma entradas, sin doble resta)
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
  
  -- LABORES GENERALES (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.project_id = lb.project_id
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name IN ('Pulverización', 'Otras Labores')
  ), 0)::numeric AS general_labors_total_usd,
  
  -- SIEMBRA (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.project_id = lb.project_id
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Siembra'
  ), 0)::numeric AS sowing_total_usd,
  
  -- RIEGO (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = lab.category_id
    WHERE w.project_id = lb.project_id
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Riego'
  ), 0)::numeric AS irrigation_total_usd,
  
  -- ARRIENDO CAPITALIZABLE (FIX: usa rent_per_ha_for_lot por cada lote)
  COALESCE((
    SELECT SUM(v3_lot_ssot.rent_per_ha_for_lot(l2.id) * l2.hectares)
    FROM public.lots l2
    JOIN public.fields f ON f.id = l2.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = lb.project_id 
      AND l2.deleted_at IS NULL
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- ADMINISTRACIÓN (sin cambios)
  COALESCE((
    SELECT p.admin_cost * SUM(v3_lot_ssot.seeded_area_for_lot(l2.id))
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l2 ON l2.field_id = f.id AND l2.deleted_at IS NULL
    WHERE p.id = lb.project_id
    GROUP BY p.admin_cost
  ), 0)::numeric AS administration_total_usd
  
FROM lot_base lb
GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
'Categorías de aporte por proyecto. FIX 000181: Controles 6 y 7 corregidos.';

-- ========================================
-- PARTE 6: CREAR INVESTOR CONTRIBUTION DATA VIEW
-- ========================================
-- Origen: Migración 000176 (esta vista FALTABA en 000179)

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH project_base AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    c.id AS customer_id,
    c.name AS customer_name,
    camp.id AS campaign_id,
    camp.name AS campaign_name,
    COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares,
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha,
    COALESCE(p.admin_cost, 0)::numeric AS admin_cost_per_ha,
    COALESCE(p.admin_cost * SUM(lb.seeded_area_ha), 0)::numeric AS administration_total_usd
  FROM public.projects p
  JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
  LEFT JOIN public.campaigns camp ON p.campaign_id = camp.id AND camp.deleted_at IS NULL
  LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN (
    SELECT
      l2.id AS lot_id,
      COALESCE(SUM(w.effective_area), 0) AS seeded_area_ha
    FROM public.lots l2
    LEFT JOIN public.workorders w ON w.lot_id = l2.id AND w.deleted_at IS NULL
    LEFT JOIN public.labors lab ON w.labor_id = lab.id
    LEFT JOIN public.categories cat ON lab.category_id = cat.id
    WHERE cat.name = 'Siembra' AND cat.type_id = 4
    GROUP BY l2.id
  ) lb ON lb.lot_id = l.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, p.name, c.id, c.name, camp.id, camp.name, p.admin_cost
),
investor_base AS (
  SELECT
    pb.project_id,
    i.id AS investor_id,
    i.name AS investor_name,
    COALESCE(pi.percentage, 0)::numeric AS share_pct_agreed
  FROM project_base pb
  JOIN public.project_investors pi ON pi.project_id = pb.project_id AND pi.deleted_at IS NULL
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
),
category_totals AS (
  SELECT
    pb.project_id,
    cc.agrochemicals_total_usd,
    cc.fertilizers_total_usd,
    cc.seeds_total_usd,
    cc.general_labors_total_usd,
    cc.sowing_total_usd,
    cc.irrigation_total_usd,
    cc.rent_capitalizable_total_usd,
    cc.administration_total_usd,
    cc.total_seeded_area_ha,
    (
      cc.agrochemicals_total_usd +
      cc.fertilizers_total_usd +
      cc.seeds_total_usd +
      cc.general_labors_total_usd +
      cc.sowing_total_usd +
      cc.irrigation_total_usd +
      cc.rent_capitalizable_total_usd +
      cc.administration_total_usd
    )::numeric AS total_contributions_usd
  FROM project_base pb
  JOIN public.v3_report_investor_contribution_categories cc ON cc.project_id = pb.project_id
),
investor_agrochemicals_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    SUM(sm.quantity * s.price)::numeric AS agrochemicals_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 2
    AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  GROUP BY sm.project_id, sm.investor_id
),
investor_fertilizers_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    SUM(sm.quantity * s.price)::numeric AS fertilizers_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 3
    AND cat.name = 'Fertilizantes'
  GROUP BY sm.project_id, sm.investor_id
),
investor_seeds_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    SUM(sm.quantity * s.price)::numeric AS seeds_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 1
    AND cat.name = 'Semilla'
  GROUP BY sm.project_id, sm.investor_id
),
investor_general_labors_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    SUM(lab.price * w.effective_area)::numeric AS general_labors_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name IN ('Pulverización', 'Otras Labores')
  GROUP BY w.project_id, w.investor_id
),
investor_sowing_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    SUM(lab.price * w.effective_area)::numeric AS sowing_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Siembra'
  GROUP BY w.project_id, w.investor_id
),
investor_irrigation_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    SUM(lab.price * w.effective_area)::numeric AS irrigation_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Riego'
  GROUP BY w.project_id, w.investor_id
),
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    COALESCE(agro.agrochemicals_real_usd, 0) AS agrochemicals_real_usd,
    COALESCE(fert.fertilizers_real_usd, 0) AS fertilizers_real_usd,
    COALESCE(seed.seeds_real_usd, 0) AS seeds_real_usd,
    COALESCE(glabor.general_labors_real_usd, 0) AS general_labors_real_usd,
    COALESCE(sow.sowing_real_usd, 0) AS sowing_real_usd,
    COALESCE(irrig.irrigation_real_usd, 0) AS irrigation_real_usd,
    (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100)::numeric AS rent_real_usd,
    (ct.administration_total_usd * ib.share_pct_agreed / 100)::numeric AS administration_real_usd,
    (
      COALESCE(agro.agrochemicals_real_usd, 0) +
      COALESCE(fert.fertilizers_real_usd, 0) +
      COALESCE(seed.seeds_real_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0) +
      (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100) +
      (ct.administration_total_usd * ib.share_pct_agreed / 100)
    )::numeric AS total_real_contribution_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN investor_agrochemicals_real agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_fertilizers_real fert ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_real seed ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors_real glabor ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
  LEFT JOIN investor_sowing_real sow ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation_real irrig ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
),
investor_harvest_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    SUM(lab.price * w.effective_area)::numeric AS harvest_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Cosecha'
  GROUP BY w.project_id, w.investor_id
),
harvest_totals AS (
  SELECT
    pb.project_id,
    COALESCE(SUM(hr.harvest_real_usd), 0)::numeric AS total_harvest_usd,
    CASE 
      WHEN pb.total_seeded_area_ha > 0 
      THEN COALESCE(SUM(hr.harvest_real_usd), 0) / pb.total_seeded_area_ha
      ELSE 0
    END::numeric AS total_harvest_usd_ha
  FROM project_base pb
  LEFT JOIN investor_harvest_real hr ON hr.project_id = pb.project_id
  GROUP BY pb.project_id, pb.total_seeded_area_ha
)
SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,

  -- SECCIÓN 1: CABECERA DE INVERSORES
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct', id.share_pct_agreed
      ) ORDER BY id.investor_id)
   FROM investor_base id
   WHERE id.project_id = pb.project_id
  ) AS investor_headers,

  -- SECCIÓN 2: DATOS GENERALES DEL PROYECTO
  jsonb_build_object(
    'surface_total_ha', COALESCE(pb.total_seeded_area_ha, 0),
    'lease_fixed_total_usd', COALESCE(ct.rent_capitalizable_total_usd, 0),
    'lease_is_fixed', true,
    'admin_per_ha_usd', CASE
      WHEN COALESCE(pb.total_seeded_area_ha, 0) > 0
      THEN COALESCE(pb.administration_total_usd, 0) / pb.total_seeded_area_ha
      ELSE 0
    END,
    'admin_total_usd', COALESCE(pb.administration_total_usd, 0)
  ) AS general_project_data,

  -- SECCIÓN 3: CATEGORÍAS DE APORTE
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'key', cat_data.key,
        'sort_index', cat_data.sort_index,
        'type', cat_data.type,
        'label', cat_data.label,
        'total_usd', cat_data.total_usd,
        'total_usd_ha', cat_data.total_usd_ha,
        'investors', cat_data.investors,
        'requires_manual_attribution', cat_data.requires_manual_attribution,
        'attribution_note', cat_data.attribution_note
      ) ORDER BY cat_data.sort_index
    )
    FROM (
      -- Seeds
      SELECT
        'seeds' AS key, 1 AS sort_index, 'pre_harvest' AS type, 'Semilla' AS label,
        ct.seeds_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.seeds_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.seeds_real_usd,
            'share_pct', CASE WHEN ct.seeds_total_usd > 0 THEN (irc.seeds_real_usd / ct.seeds_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Agrochemicals
      SELECT
        'agrochemicals' AS key, 2 AS sort_index, 'pre_harvest' AS type, 'Agroquímicos' AS label,
        ct.agrochemicals_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.agrochemicals_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.agrochemicals_real_usd,
            'share_pct', CASE WHEN ct.agrochemicals_total_usd > 0 THEN (irc.agrochemicals_real_usd / ct.agrochemicals_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Fertilizers
      SELECT
        'fertilizers' AS key, 3 AS sort_index, 'pre_harvest' AS type, 'Fertilizantes' AS label,
        ct.fertilizers_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.fertilizers_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.fertilizers_real_usd,
            'share_pct', CASE WHEN ct.fertilizers_total_usd > 0 THEN (irc.fertilizers_real_usd / ct.fertilizers_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Sowing
      SELECT
        'sowing' AS key, 4 AS sort_index, 'pre_harvest' AS type, 'Siembra' AS label,
        ct.sowing_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.sowing_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.sowing_real_usd,
            'share_pct', CASE WHEN ct.sowing_total_usd > 0 THEN (irc.sowing_real_usd / ct.sowing_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- General Labors
      SELECT
        'general_labors' AS key, 5 AS sort_index, 'pre_harvest' AS type, 'Labores Generales' AS label,
        ct.general_labors_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.general_labors_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.general_labors_real_usd,
            'share_pct', CASE WHEN ct.general_labors_total_usd > 0 THEN (irc.general_labors_real_usd / ct.general_labors_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Irrigation
      SELECT
        'irrigation' AS key, 6 AS sort_index, 'pre_harvest' AS type, 'Riego' AS label,
        ct.irrigation_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.irrigation_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', irc.investor_id,
            'investor_name', irc.investor_name,
            'amount_usd', irc.irrigation_real_usd,
            'share_pct', CASE WHEN ct.irrigation_total_usd > 0 THEN (irc.irrigation_real_usd / ct.irrigation_total_usd * 100) ELSE 0 END
          ) ORDER BY irc.investor_id)
          FROM investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Administration Structure
      SELECT
        'administration_structure' AS key, 7 AS sort_index, 'pre_harvest' AS type, 'Administración y Estructura' AS label,
        ct.administration_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.administration_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', ib.investor_id,
            'investor_name', ib.investor_name,
            'amount_usd', (ct.administration_total_usd * ib.share_pct_agreed / 100),
            'share_pct', ib.share_pct_agreed
          ) ORDER BY ib.investor_id)
          FROM investor_base ib
          WHERE ib.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id

      UNION ALL

      -- Capitalizable Lease
      SELECT
        'capitalizable_lease' AS key, 8 AS sort_index, 'pre_harvest' AS type, 'Arriendo Capitalizable' AS label,
        ct.rent_capitalizable_total_usd AS total_usd,
        CASE WHEN ct.total_seeded_area_ha > 0 THEN ct.rent_capitalizable_total_usd / ct.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', ib.investor_id,
            'investor_name', ib.investor_name,
            'amount_usd', (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100),
            'share_pct', ib.share_pct_agreed
          ) ORDER BY ib.investor_id)
          FROM investor_base ib
          WHERE ib.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM category_totals ct WHERE ct.project_id = pb.project_id
    ) cat_data
  ) AS contribution_categories,

  -- SECCIÓN 4: COMPARACIÓN ACORDADO vs REAL
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', irc.investor_id,
        'investor_name', irc.investor_name,
        'agreed_share_pct', irc.share_pct_agreed,
        'agreed_usd', (ct.total_contributions_usd * irc.share_pct_agreed / 100),
        'actual_usd', irc.total_real_contribution_usd,
        'adjustment_usd', (irc.total_real_contribution_usd - (ct.total_contributions_usd * irc.share_pct_agreed / 100))
      ) ORDER BY irc.investor_id
    )
    FROM investor_real_contributions irc
    JOIN category_totals ct ON ct.project_id = irc.project_id
    WHERE irc.project_id = pb.project_id
  ) AS investor_contribution_comparison,

  -- SECCIÓN 5: LIQUIDACIÓN DE COSECHA
  jsonb_build_object(
    'rows', jsonb_build_array(
      -- Fila "Cosecha" con valores reales
      jsonb_build_object(
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', COALESCE(ht.total_harvest_usd, 0),
        'total_us_ha', COALESCE(ht.total_harvest_usd_ha, 0),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', hr.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', hr.harvest_real_usd,
              'share_pct', CASE 
                WHEN ht.total_harvest_usd > 0 
                THEN (hr.harvest_real_usd / ht.total_harvest_usd * 100)
                ELSE 0 
              END
            ) ORDER BY hr.investor_id
          )
          FROM investor_harvest_real hr
          JOIN investor_base ib ON ib.investor_id = hr.investor_id AND ib.project_id = hr.project_id
          WHERE hr.project_id = pb.project_id
        ), '[]'::jsonb)
      ),
      -- Fila "Totales"
      jsonb_build_object(
        'key', 'totals',
        'type', 'totals',
        'total_usd', COALESCE(ht.total_harvest_usd, 0),
        'total_us_ha', COALESCE(ht.total_harvest_usd_ha, 0),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', hr.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', hr.harvest_real_usd,
              'share_pct', CASE 
                WHEN ht.total_harvest_usd > 0 
                THEN (hr.harvest_real_usd / ht.total_harvest_usd * 100)
                ELSE 0 
              END
            ) ORDER BY hr.investor_id
          )
          FROM investor_harvest_real hr
          JOIN investor_base ib ON ib.investor_id = hr.investor_id AND ib.project_id = hr.project_id
          WHERE hr.project_id = pb.project_id
        ), '[]'::jsonb)
      )
    ),
    -- Footer: Pago acordado por inversor
    'footer_payment_agreed', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', (ht.total_harvest_usd * ib.share_pct_agreed / 100),
          'share_pct', ib.share_pct_agreed
        ) ORDER BY ib.investor_id
      )
      FROM investor_base ib
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb),
    -- Footer: Ajuste de pago
    'footer_payment_adjustment', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', (
            COALESCE(hr.harvest_real_usd, 0) - 
            (ht.total_harvest_usd * ib.share_pct_agreed / 100)
          )
        ) ORDER BY ib.investor_id
      )
      FROM investor_base ib
      LEFT JOIN investor_harvest_real hr ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb)
  ) AS harvest_settlement

FROM project_base pb
JOIN category_totals ct ON ct.project_id = pb.project_id
LEFT JOIN harvest_totals ht ON ht.project_id = pb.project_id
ORDER BY pb.project_id;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
'Vista FINAL para informe de Aportes por Inversor. FIX 000181: Controles 6, 7, 9 corregidos.';

COMMIT;

