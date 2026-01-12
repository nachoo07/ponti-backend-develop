-- ========================================
-- MIGRACIÓN 000113: CREATE v3_core_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Crear esquema v3_core_ssot con funciones matemáticas puras
-- Alcance: Funciones INMUTABLES y ESTABLES sin dependencias de otros esquemas SSOT
-- Total: 30 funciones (matemáticas, conversiones, FX, cálculos agrícolas básicos)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español
-- Nota: Funciones de lote (*_for_lot) movidas a v3_lot_ssot (000114)

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_core_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_core_ssot;

COMMENT ON SCHEMA v3_core_ssot IS 'Funciones SSOT transversales: matemáticas, conversiones, helpers básicos';

-- ========================================
-- GRUPO 1: OPERACIONES MATEMÁTICAS SEGURAS (7 funciones)
-- ========================================
-- Propósito: Prevenir errores de división por cero y manejar nulos consistentemente

CREATE OR REPLACE FUNCTION v3_core_ssot.coalesce0(numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.coalesce0(double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.safe_div(numeric, numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.safe_div_dp(double precision, double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.percentage(numeric, numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.percentage_capped(numeric, numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT LEAST(v3_core_ssot.safe_div($1, $2) * 100, 100)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.percentage_rounded(numeric, numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div($1, $2) * 100
$$;

-- ========================================
-- GRUPO 2: CONVERSIONES POR HECTÁREA (5 funciones)
-- ========================================
-- Propósito: Normalizar valores por hectárea (costos, dosis, rendimientos)

CREATE OR REPLACE FUNCTION v3_core_ssot.per_ha(numeric, numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.per_ha_dp(double precision, double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div_dp($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.units_per_ha(units numeric, area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha(units, area)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.dose_per_ha(total_dose numeric, surface_ha numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div(total_dose, surface_ha)
$$;

CREATE OR REPLACE FUNCTION v3_core_ssot.norm_dose(dose numeric, area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;

-- ========================================
-- GRUPO 3: MANEJO DE FECHAS (1 función)
-- ========================================
-- Propósito: Cálculos de fechas para campañas

CREATE OR REPLACE FUNCTION v3_core_ssot.calculate_campaign_closing_date(end_date date) 
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT CASE 
    WHEN end_date IS NULL THEN NULL
    ELSE end_date + (get_campaign_closure_days() || ' days')::INTERVAL
  END::date
$$;

-- ========================================
-- GRUPO 4: CONVERSIÓN DE MONEDAS (FX) (2 funciones)
-- ========================================
-- Propósito: Obtener valores de dólar para conversiones

CREATE OR REPLACE FUNCTION v3_core_ssot.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS NUMERIC AS $$
DECLARE
  dollar_value NUMERIC;
BEGIN
  -- Obtener valor del dólar para el proyecto y mes específico
  SELECT d.average_value INTO dollar_value
  FROM project_dollar_values d
  WHERE d.project_id = p_project_id 
    AND d.month = p_month
    AND d.deleted_at IS NULL
  LIMIT 1;
  
  -- Si no existe valor, retornar 1.0 como fallback
  RETURN COALESCE(dollar_value, 1.0);
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION v3_core_ssot.dollar_average_for_month(p_project_id bigint, p_date date) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT AVG(d.average_value)
     FROM project_dollar_values d
     WHERE d.project_id = p_project_id
       AND TO_CHAR(p_date, 'YYYY-MM') = d.month
       AND d.deleted_at IS NULL),
    1.0
  )
$$;

-- ========================================
-- GRUPO 5: CÁLCULOS AGRÍCOLAS BÁSICOS (7 funciones)
-- ========================================
-- Propósito: Cálculos agrícolas básicos aplicables a múltiples módulos

-- Área sembrada: si hay fecha de siembra, cuenta la superficie
CREATE OR REPLACE FUNCTION v3_core_ssot.seeded_area(sowing_date date, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Área cosechada: si hay toneladas > 0, cuenta la superficie
CREATE OR REPLACE FUNCTION v3_core_ssot.harvested_area(tons numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;

-- Rendimiento tn/ha (base hectáreas declaradas)
CREATE OR REPLACE FUNCTION v3_core_ssot.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div(COALESCE(tons,0), COALESCE(hectares,0))
$$;

-- Rendimiento tn/ha (base área cosechada)
CREATE OR REPLACE FUNCTION v3_core_ssot.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.safe_div(COALESCE(tons,0), COALESCE(harvested_area,0))
$$;

-- Costo de labor
CREATE OR REPLACE FUNCTION v3_core_ssot.labor_cost(labor_price numeric, effective_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;

-- Costo de insumo
CREATE OR REPLACE FUNCTION v3_core_ssot.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;

-- Costo por hectárea
CREATE OR REPLACE FUNCTION v3_core_ssot.cost_per_ha(total_cost numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha(total_cost, hectares)
$$;

-- ========================================
-- GRUPO 6: INGRESOS BÁSICOS (2 funciones)
-- ========================================
-- Propósito: Cálculos de ingresos básicos

-- Ingreso neto total
CREATE OR REPLACE FUNCTION v3_core_ssot.income_net_total(tons numeric, net_price_usd numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;

-- Ingreso neto por hectárea
CREATE OR REPLACE FUNCTION v3_core_ssot.income_net_per_ha(income_net_total numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha(income_net_total, hectares)
$$;

-- ========================================
-- GRUPO 7: RENTA BÁSICA (3 funciones)
-- ========================================
-- Propósito: Cálculos de renta según tipo de arriendo

-- Renta por ha según tipo de arriendo (integer)
CREATE OR REPLACE FUNCTION v3_core_ssot.rent_per_ha(
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

-- Renta por ha según tipo de arriendo (bigint overload)
CREATE OR REPLACE FUNCTION v3_core_ssot.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.rent_per_ha(
    lease_type_id::integer, 
    lease_type_percent, 
    lease_type_value,
    income_net_per_ha, 
    cost_per_ha, 
    admin_cost_per_ha
  )
$$;

-- Cálculo simple de renta por ha (validación de negativos)
CREATE OR REPLACE FUNCTION v3_core_ssot.calculate_rent_per_ha(lease_value DOUBLE PRECISION)
RETURNS DOUBLE PRECISION AS $$
BEGIN
  IF lease_value < 0 THEN
    RETURN 0;
  ELSE
    RETURN lease_value;
  END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ========================================
-- GRUPO 8: ACTIVOS Y RESULTADOS BÁSICOS (3 funciones)
-- ========================================
-- Propósito: Cálculos de activos y resultados operativos

-- Activo total por ha = costo directo/ha + renta/ha + admin/ha
CREATE OR REPLACE FUNCTION v3_core_ssot.active_total_per_ha(
  direct_cost_per_ha double precision,
  rent_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;

-- Resultado operativo por ha = ingreso neto/ha - activo total/ha
CREATE OR REPLACE FUNCTION v3_core_ssot.operating_result_per_ha(
  income_net_per_ha double precision,
  active_total_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;

-- Precio de indiferencia (USD / tn) = invertido_por_ha / (tn/ha)
CREATE OR REPLACE FUNCTION v3_core_ssot.indifference_price_usd_tn(
  total_invested_per_ha double precision, 
  yield_tn_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.per_ha_dp(total_invested_per_ha, yield_tn_per_ha)
$$;

-- ========================================
-- NOTA: FUNCIONES DE LOTE MOVIDAS A v3_lot_ssot (migración 000114)
-- ========================================
-- Todas las funciones *_for_lot y funciones relacionadas con lotes específicos
-- están ahora en v3_lot_ssot para evitar dependencias circulares.

COMMIT;

-- Comentarios sobre los grupos de funciones
COMMENT ON FUNCTION v3_core_ssot.percentage(numeric, numeric) IS 'Calcula porcentaje seguro (sin división por cero)';
COMMENT ON FUNCTION v3_core_ssot.per_ha(numeric, numeric) IS 'Normaliza valor por hectárea';
COMMENT ON FUNCTION v3_core_ssot.calculate_campaign_closing_date(date) IS 'Calcula fecha de cierre de campaña';
COMMENT ON FUNCTION v3_core_ssot.seeded_area(date, numeric) IS 'Calcula área sembrada basada en fecha de siembra';
COMMENT ON FUNCTION v3_core_ssot.harvested_area(numeric, numeric) IS 'Calcula área cosechada basada en toneladas';

