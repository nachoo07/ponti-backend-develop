-- ========================================
-- MIGRACIÓN 000115: CREATE v3_lot_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Crear esquema v3_lot_ssot con TODAS las funciones de lotes (CONSOLIDADO DRY)
-- Dependencias: Requiere v3_core_ssot (migración 000113)
-- Alcance: 25 funciones (consolidadas desde workorder_ssot y dashboard_ssot)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY:
-- - Agregadas: supply_cost_for_lot_base, supply_cost_for_lot_by_category
-- - Movidas desde workorder_ssot: surface_for_lot, liters_for_lot, kilograms_for_lot
-- - Elimina duplicados de: labor_cost_for_lot_wo, supply_cost_*_mb
-- 
-- Nota: Código en inglés, comentarios en español
-- Usa v3_core_ssot para operaciones matemáticas básicas

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_lot_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_lot_ssot;

COMMENT ON SCHEMA v3_lot_ssot IS 'Funciones SSOT de lots: transversales usadas por dashboard + exclusivas de lots';

-- ========================================
-- GRUPO 1: ACCESSORS DE LOTE (2 funciones) - TRANSVERSALES
-- ========================================
-- Propósito: Helpers básicos usados por otras funciones
-- Uso: Dashboard, Lots, Reports

CREATE OR REPLACE FUNCTION v3_lot_ssot.lot_hectares(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_lot_ssot.lot_tons(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- ========================================
-- GRUPO 2: COSTOS Y AGREGACIONES POR LOTE (8 funciones) - CONSOLIDADO DRY
-- ========================================
-- Propósito: Funciones consolidadas desde v3_workorder_ssot y v3_dashboard_ssot
-- Eliminadas: labor_cost_for_lot_wo, supply_cost_for_lot_wo, surface/liters/kilograms_for_lot
--             labor_cost_for_lot_mb, supply_cost_seeds/agrochemicals_for_lot_mb

-- 2.1: Costo de labor (SSOT único)
CREATE OR REPLACE FUNCTION v3_lot_ssot.labor_cost_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;

-- 2.2: Costo de insumos base (solo workorder_items, sin movimientos internos)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

-- 2.3: Costo de insumos por categoría (para management balance)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_by_category(p_lot_id bigint, p_category_name text)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.final_dose * s.price * w.effective_area), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id
  JOIN public.supplies s ON s.id = wi.supply_id
  JOIN public.categories c ON c.id = s.category_id
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.final_dose > 0
    AND s.price IS NOT NULL
$$;

-- 2.4: Superficie total trabajada en el lote
CREATE OR REPLACE FUNCTION v3_lot_ssot.surface_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(w.effective_area), 0)::numeric
  FROM public.workorders w
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;

-- 2.5: Litros de insumos aplicados en el lote
CREATE OR REPLACE FUNCTION v3_lot_ssot.liters_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

-- 2.6: Kilogramos de insumos aplicados en el lote
CREATE OR REPLACE FUNCTION v3_lot_ssot.kilograms_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;

-- 2.7: Costo de insumos completo (con movimientos internos) - MANTENER como estaba
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot(p_lot_id bigint) 
RETURNS double precision
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

-- ========================================
-- GRUPO 2: INGRESOS POR LOTE (3 funciones) - EN ORDEN DE DEPENDENCIAS
-- ========================================
-- Propósito: Calcular ingresos por lote

-- 2.1: Precio neto (no depende de otras funciones SSOT lot)
CREATE OR REPLACE FUNCTION v3_lot_ssot.net_price_usd_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

-- 2.2: Ingreso neto total (depende de net_price_usd_for_lot) - MOVIDO DESDE GRUPO 6
CREATE OR REPLACE FUNCTION v3_lot_ssot.income_net_total_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v3_lot_ssot.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- 2.3: Ingreso neto por ha (depende de income_net_total_for_lot)
CREATE OR REPLACE FUNCTION v3_lot_ssot.income_net_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.safe_div_dp(
           COALESCE(v3_lot_ssot.income_net_total_for_lot(p_lot_id), 0)::double precision,
           v3_lot_ssot.lot_hectares(p_lot_id)
         )
$$;

-- ========================================
-- GRUPO 2bis: COSTO DIRECTO POR LOTE (1 función) - MOVIDO DESDE GRUPO 6
-- ========================================
-- Propósito: Calcular costo directo total por lote
-- Nota: Debe estar ANTES de cost_per_ha_for_lot que lo usa

-- Costo directo total = labor + supplies
CREATE OR REPLACE FUNCTION v3_lot_ssot.direct_cost_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(v3_lot_ssot.labor_cost_for_lot(p_lot_id), 0)::double precision
       + COALESCE(v3_lot_ssot.supply_cost_for_lot(p_lot_id), 0)
$$;

-- ========================================
-- GRUPO 3: COSTOS POR HA POR LOTE (1 función)
-- ========================================
-- Propósito: Calcular costos normalizados por hectárea

CREATE OR REPLACE FUNCTION v3_lot_ssot.cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.safe_div_dp(
           COALESCE(v3_lot_ssot.direct_cost_for_lot(p_lot_id), 0)::double precision,
           v3_lot_ssot.lot_hectares(p_lot_id)
         )
$$;

-- ========================================
-- GRUPO 3bis: FUNCIONES TRANSVERSALES ANTICIPADAS (2 funciones)
-- ========================================
-- Propósito: Funciones que necesitan estar ANTES de rent_per_ha_for_lot_fixed
-- Movidas desde GRUPO 6 por orden de dependencias

-- Costo administrativo por ha (distribuido por proyecto)
CREATE OR REPLACE FUNCTION v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT CASE WHEN t.total_hectares > 0
              THEN COALESCE(p.admin_cost, 0)::double precision / t.total_hectares
              ELSE 0 END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    -- Calcular total_hectares inline para evitar dependencia circular
    SELECT COALESCE(
      (SELECT SUM(l2.hectares)
       FROM public.lots l2
       JOIN public.fields f2 ON f2.id = l2.field_id AND f2.deleted_at IS NULL
       WHERE f2.project_id = f.project_id AND l2.deleted_at IS NULL), 
      0
    )::double precision AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- Renta por ha según tipo de arriendo
CREATE OR REPLACE FUNCTION v3_lot_ssot.rent_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v3_lot_ssot.income_net_per_ha_for_lot(p_lot_id),
           v3_lot_ssot.cost_per_ha_for_lot(p_lot_id),
           v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- ========================================
-- GRUPO 3ter: RENTA FIXED (1 función)
-- ========================================
-- Propósito: Wrapper de rent_per_ha_for_lot con validación de negativos

CREATE OR REPLACE FUNCTION v3_lot_ssot.rent_per_ha_for_lot_fixed(p_lot_id BIGINT)
RETURNS DOUBLE PRECISION AS $$
DECLARE
  calculated_rent DOUBLE PRECISION;
BEGIN
  calculated_rent := v3_lot_ssot.rent_per_ha_for_lot(p_lot_id);
  
  -- Si el valor es negativo, retornar 0
  IF calculated_rent < 0 THEN
    RETURN 0;
  ELSE
    RETURN calculated_rent;
  END IF;
END;
$$ LANGUAGE plpgsql STABLE;

-- ========================================
-- GRUPO 4: RESULTADOS POR LOTE (3 funciones)
-- ========================================
-- Propósito: Calcular activos y resultados operativos por lote

CREATE OR REPLACE FUNCTION v3_lot_ssot.active_total_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.active_total_per_ha(
           v3_lot_ssot.cost_per_ha_for_lot(p_lot_id),
           v3_lot_ssot.rent_per_ha_for_lot(p_lot_id),
           v3_lot_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v3_lot_ssot.operating_result_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.operating_result_per_ha(
           v3_lot_ssot.income_net_per_ha_for_lot(p_lot_id),
           v3_lot_ssot.active_total_per_ha_for_lot(p_lot_id)
         )
$$;

CREATE OR REPLACE FUNCTION v3_lot_ssot.yield_tn_per_ha_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.per_ha_dp(
           v3_lot_ssot.lot_tons(p_lot_id),
           v3_lot_ssot.lot_hectares(p_lot_id)
         )
$$;

-- ========================================
-- GRUPO 5: ÁREAS POR LOTE (2 funciones)
-- ========================================
-- Propósito: Calcular áreas sembradas y cosechadas

CREATE OR REPLACE FUNCTION v3_lot_ssot.seeded_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_lot_ssot.harvested_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.harvested_area(
           v3_lot_ssot.lot_tons(p_lot_id)::numeric,
           v3_lot_ssot.lot_hectares(p_lot_id)::numeric
         )
$$;

-- ========================================
-- GRUPO 6: FUNCIONES TRANSVERSALES RESTANTES (2 funciones) - TRANSVERSALES
-- ========================================
-- Propósito: Funciones de lote usadas por Dashboard y otros módulos
-- Uso: Dashboard, Reports, Workorders
-- Nota: Funciones movidas:
--   - direct_cost_for_lot e income_net_total_for_lot → GRUPO 2/2bis
--   - admin_cost_per_ha_for_lot y rent_per_ha_for_lot → GRUPO 3bis

-- Porcentaje de rentabilidad
CREATE OR REPLACE FUNCTION v3_lot_ssot.renta_pct(
  operating_result_total_usd double precision, 
  total_costs_usd double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;

-- Helper: suma de costos directos
CREATE OR REPLACE FUNCTION v3_lot_ssot.direct_cost_usd(
  p_labor_cost_usd numeric,
  p_supply_cost_usd numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0)
$$;

COMMIT;

-- Comentarios sobre funciones clave
COMMENT ON FUNCTION v3_lot_ssot.labor_cost_for_lot(bigint) IS 'Calcula costo de labores por lote';
COMMENT ON FUNCTION v3_lot_ssot.supply_cost_for_lot(bigint) IS 'Calcula costo de insumos por lote';
COMMENT ON FUNCTION v3_lot_ssot.yield_tn_per_ha_for_lot(bigint) IS 'Calcula rendimiento en toneladas por hectárea';

