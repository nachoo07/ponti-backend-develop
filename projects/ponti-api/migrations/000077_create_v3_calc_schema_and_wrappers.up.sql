-- ========================================
-- MIGRATION 000077: CREATE v3_calc SCHEMA - SSOT/DRY CALCULATIONS (UP)
-- ========================================
-- 
-- Purpose: Single Source of Truth for all calculations (DRY principle)
-- Date: 2025-09-13
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- ============================================================================
-- SINGLE SOURCE OF TRUTH (SSOT) - v3_calc schema
-- 
-- DRY Principle Implementation:
--  ✓ Centralized calculations (no duplication across views/code)
--  ✓ Immutable functions (deterministic, cacheable)
--  ✓ Safe operations (division by zero protection)
--  ✓ Consistent business logic (same formula everywhere)
--  ✓ Composable functions (small, reusable pieces)
--  ✓ Type safety (proper numeric/double precision handling)
-- 
-- Best Practices Applied:
--  ✓ Schema separation (v3_calc namespace isolation)
--  ✓ Function naming convention (descriptive, consistent)
--  ✓ Parameter validation (COALESCE for null safety)
--  ✓ Performance optimization (STABLE vs IMMUTABLE)
--  ✓ Documentation (clear purpose for each function)
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS v3_calc;

-- =============================================================================
-- LAYER 1: SAFE MATHEMATICAL OPERATIONS (Foundation)
-- =============================================================================
-- Purpose: Null-safe arithmetic operations that prevent division by zero
-- and handle edge cases consistently across all business calculations
CREATE OR REPLACE FUNCTION v3_calc.coalesce0(numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.coalesce0(double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_calc.safe_div_dp(double precision, double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION v3_calc.percentage_capped(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT LEAST(v3_calc.safe_div($1, $2) * 100, 100)
$$;

-- =============================================================================
-- LAYER 2: BUSINESS UNIT CONVERSIONS (Per-hectare calculations)
-- =============================================================================
-- Purpose: Standardized per-hectare conversions used across all agricultural metrics
CREATE OR REPLACE FUNCTION v3_calc.per_ha(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.per_ha_dp(double precision, double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div_dp($1, $2)
$$;

-- =============================================================================
-- LAYER 3: AGRICULTURAL DOMAIN CALCULATIONS (Business Logic)
-- =============================================================================
-- Purpose: Core agricultural business calculations following domain rules

-- Dosis normalizada (suma de dosis sobre superficie)
CREATE OR REPLACE FUNCTION v3_calc.dose_per_ha(total_dose numeric, surface_ha numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div(total_dose, surface_ha)
$$;

-- Área sembrada: si hay fecha de siembra, cuenta la superficie
CREATE OR REPLACE FUNCTION v3_calc.seeded_area(sowing_date date, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Área cosechada: si hay toneladas > 0, cuenta la superficie
CREATE OR REPLACE FUNCTION v3_calc.harvested_area(tons numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Rendimiento tn/ha (base hectáreas declaradas)
CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div( COALESCE(tons,0), COALESCE(hectares,0) )
$$;

-- Rendimiento tn/ha (base área cosechada)
CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div( COALESCE(tons,0), COALESCE(harvested_area,0) )
$$;

-- Costos
CREATE OR REPLACE FUNCTION v3_calc.labor_cost(labor_price numeric, effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha(total_cost numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha(total_cost, hectares)
$$;

-- Ingresos
CREATE OR REPLACE FUNCTION v3_calc.income_net_total(tons numeric, net_price_usd numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha(income_net_total numeric, hectares numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha(income_net_total, hectares)
$$;

-- Renta por ha (según tipo de arriendo)
-- 1: % sobre ingreso neto por ha
-- 2: % sobre (ingreso neto - costo directo/ha - admin/ha)
-- 3: valor fijo por ha
-- 4: valor fijo por ha + % sobre ingreso neto por ha
CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha(
  lease_type_id integer,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT
    CASE
      WHEN lease_type_id = 1 THEN COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0)
      WHEN lease_type_id = 2 THEN COALESCE(lease_type_percent,0)/100.0 *
                               (COALESCE(income_net_per_ha,0) - COALESCE(cost_per_ha,0) - COALESCE(admin_cost_per_ha,0))
      WHEN lease_type_id = 3 THEN COALESCE(lease_type_value,0)
      WHEN lease_type_id = 4 THEN COALESCE(lease_type_value,0) +
                               (COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0))
      ELSE 0
    END
$$;

-- Overload para lease_type_id bigint (compatibilidad con fields.lease_type_id bigserial)
CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.rent_per_ha(lease_type_id::integer, lease_type_percent, lease_type_value,
                          income_net_per_ha, cost_per_ha, admin_cost_per_ha)
$$;

-- Activo total por ha = costo directo/ha + renta/ha + admin/ha
CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha(
  direct_cost_per_ha double precision,
  rent_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;

-- Resultado operativo por ha = ingreso neto/ha - activo total/ha
CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha(
  income_net_per_ha double precision,
  active_total_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;

-- % de renta = resultado operativo / costos totales
CREATE OR REPLACE FUNCTION v3_calc.renta_pct(operating_result_total_usd double precision, total_costs_usd double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;

-- Precio de indiferencia (USD / tn) = invertido_por_ha / (tn/ha)
CREATE OR REPLACE FUNCTION v3_calc.indifference_price_usd_tn(total_invested_per_ha double precision, yield_tn_per_ha double precision) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha_dp(total_invested_per_ha, yield_tn_per_ha)
$$;

-- Unidades por ha (litros/ha, kg/ha, etc.)
CREATE OR REPLACE FUNCTION v3_calc.units_per_ha(units numeric, area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha(units, area)
$$;

-- Alias para la función existente de dosis normalizada
CREATE OR REPLACE FUNCTION v3_calc.norm_dose(dose numeric, area numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;

-- =============================================================================
-- LAYER 7: CAMPAIGN DATE CALCULATIONS (Business date logic)
-- =============================================================================
-- Purpose: Campaign-related date calculations for business operations

-- Función para calcular fecha de cierre de campaña
CREATE OR REPLACE FUNCTION v3_calc.calculate_campaign_closing_date(end_date date) RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT CASE 
    WHEN end_date IS NULL THEN NULL
    ELSE end_date + (get_campaign_closure_days() || ' days')::INTERVAL  -- Usar valor de app_parameters
  END::date
$$;

-- No se necesita wrapper público - las vistas v3 usan directamente v3_calc.*

-- =============================================================================
-- LAYER 4: LOT-LEVEL BUSINESS QUERIES (Data Access Layer)
-- =============================================================================
-- Purpose: Lot-specific calculations that read transactional tables
-- These replace base_* views with direct, optimized queries
-- Performance: STABLE functions (can be cached within transaction)

-- Basic lot data accessors (DRY: avoid repeating SELECT against public.lots)
CREATE OR REPLACE FUNCTION v3_calc.lot_hectares(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.lot_tons(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- COST CALCULATIONS (SSOT for all cost-related business logic)
CREATE OR REPLACE FUNCTION v3_calc.labor_cost_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

-- Supply cost calculation (aggregates all supply items for a lot)
-- FIX: Considera tanto workorder_items como movimientos internos de salida
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
    +
    -- Costos por movimientos internos de salida (insumos transferidos a otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.workorders w ON w.lot_id = p_lot_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND w.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = w.project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

-- Total direct cost (SSOT: labor + supply costs combined)
CREATE OR REPLACE FUNCTION v3_calc.direct_cost_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_calc.labor_cost_for_lot(p_lot_id), 0)::double precision
       + COALESCE(v3_calc.supply_cost_for_lot(p_lot_id), 0)
$$;

-- =============================================================================
-- SUPPLY COST FUNCTIONS FOR INTERNAL MOVEMENTS
-- =============================================================================
-- Función para calcular costos de supply para un proyecto considerando movimientos internos
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
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
    -- Costos por movimientos internos de salida (insumos transferidos a otros proyectos)
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

-- Función para calcular costos de supply recibidos por movimientos internos
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_received_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por movimientos internos de entrada (insumos recibidos de otros proyectos)
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

-- =============================================================================
-- LAYER 5: INCOME CALCULATIONS (Revenue and pricing logic)
-- =============================================================================
-- Net price lookup (gets current pricing for lot's crop in project)
CREATE OR REPLACE FUNCTION v3_calc.net_price_usd_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id    = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

-- Total net income for lot (tons × net_price)
CREATE OR REPLACE FUNCTION v3_calc.income_net_total_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v3_calc.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Income per hectare (applies per-ha conversion to total income)
CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.income_net_total_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

-- =============================================================================
-- LAYER 6: PROJECT-LEVEL AGGREGATIONS (Administrative calculations)
-- =============================================================================
-- Total hectares for project (used for admin cost proration)
CREATE OR REPLACE FUNCTION v3_calc.total_hectares_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::double precision
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;

-- Costo de administración por ha para un lote
CREATE OR REPLACE FUNCTION v3_calc.admin_cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::double precision / t.total_hectares
              ELSE 0 END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT v3_calc.total_hectares_for_project(f.project_id) AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Costo directo por ha para un lote
CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.direct_cost_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

-- Renta por ha para un lote (según reglas del field)
CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Activo total por ha para un lote
CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.active_total_per_ha(
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.rent_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;

-- Resultado operativo por ha para un lote
CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.operating_result_per_ha(
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.active_total_per_ha_for_lot(p_lot_id)
         )
$$;

-- Rendimiento tn/ha para un lote
CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.per_ha_dp(
           v3_calc.lot_tons(p_lot_id),
           v3_calc.lot_hectares(p_lot_id)
         )
$$;

-- Área sembrada (ha) para un lote
CREATE OR REPLACE FUNCTION v3_calc.seeded_area_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Área cosechada (ha) para un lote
CREATE OR REPLACE FUNCTION v3_calc.harvested_area_for_lot(p_lot_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.harvested_area(
           v3_calc.lot_tons(p_lot_id)::numeric,
           v3_calc.lot_hectares(p_lot_id)::numeric
         )
$$;

-- =============================================================================
-- FUNCIONES ADICIONALES PARA DASHBOARD FIXES
-- =============================================================================
-- Agregar estas funciones antes del COMMIT para completar la funcionalidad del dashboard

-- Función para calcular presupuesto total de costos por proyecto
CREATE OR REPLACE FUNCTION v3_calc.total_budget_cost_for_project(p_project_id bigint) RETURNS numeric
LANGUAGE sql STABLE AS $$
  -- Por ahora retornamos un valor placeholder hasta que se implemente
  -- el sistema de presupuestos. En el futuro esto debería consultar
  -- una tabla de presupuestos o calcular basado en labores/insumos planificados
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric  -- Placeholder: 10x admin_cost
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

-- Función para calcular total invertido por proyecto (costos directos + arriendo + estructura)
CREATE OR REPLACE FUNCTION v3_calc.total_invested_cost_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Arriendo invertido
    (SELECT COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Estructura invertida
    (SELECT COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Función para calcular resultado operativo correcto (ingresos - costos directos - admin)
CREATE OR REPLACE FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_calc.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costo administrativo total
    (SELECT COALESCE(p.admin_cost, 0)::double precision
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Función para calcular costos directos invertidos (solo labores + insumos)
-- FIX: Considera movimientos internos en el cálculo de costos invertidos
CREATE OR REPLACE FUNCTION v3_calc.direct_costs_invested_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Labores invertidas (ejecutadas + no ejecutadas)
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Insumos invertidos (usados + no usados) - usar datos de stocks
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    -- Insumos recibidos por movimientos internos
    v3_calc.supply_cost_received_for_project(p_project_id)
  , 0)::double precision
$$;

-- Función para calcular costos totales por cultivo
CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para el cultivo
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Función para calcular costos totales del proyecto
CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para todo el proyecto
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Función para calcular stock disponible por proyecto
CREATE OR REPLACE FUNCTION v3_calc.stock_value_for_project(p_project_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock disponible = insumos comprados - insumos consumidos
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id 
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Función para calcular costo por hectárea por cultivo
CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.per_ha_dp(
    v3_calc.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;

-- =============================================================================
-- FIX: ADMIN_COST TYPE CHANGE (BIGINT to NUMERIC for decimal.Decimal compatibility)
-- =============================================================================
-- Eliminar vistas que dependen de admin_cost antes del cambio de tipo
DROP VIEW IF EXISTS v3_dashboard CASCADE;
DROP VIEW IF EXISTS v3_dashboard_metrics CASCADE;
DROP VIEW IF EXISTS v3_dashboard_balance CASCADE;
DROP VIEW IF EXISTS v3_dashboard_management CASCADE;
DROP VIEW IF EXISTS v3_dashboard_crop_incidence CASCADE;
DROP VIEW IF EXISTS v3_dashboard_operational_indicators CASCADE;

-- Cambiar el tipo de admin_cost de BIGINT a NUMERIC(15,3) para compatibilidad con decimal.Decimal
ALTER TABLE projects 
ALTER COLUMN admin_cost TYPE NUMERIC(15,3) USING admin_cost::NUMERIC(15,3);

-- Comentario para documentar el cambio
COMMENT ON COLUMN projects.admin_cost IS 'Costo administrativo del proyecto en USD con 3 decimales de precisión';

COMMIT;


