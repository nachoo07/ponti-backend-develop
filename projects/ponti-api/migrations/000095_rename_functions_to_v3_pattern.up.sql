-- Rename functions to follow v3 pattern
-- Todas las funciones creadas después de la migración 76 deben comenzar con v3

-- 1. Renombrar get_project_dollar_value a v3_calc.get_project_dollar_value
-- Primero crear la función en el esquema v3_calc
CREATE OR REPLACE FUNCTION v3_calc.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS DOUBLE PRECISION AS $$
BEGIN
    -- Llamar a la función original
    RETURN get_project_dollar_value(p_project_id, p_month);
END;
$$ LANGUAGE plpgsql;

-- 2. Renombrar calculate_rent_per_ha a v3_calc.calculate_rent_per_ha
-- Primero crear la función en el esquema v3_calc
CREATE OR REPLACE FUNCTION v3_calc.calculate_rent_per_ha(lease_value DOUBLE PRECISION)
RETURNS DOUBLE PRECISION AS $$
BEGIN
    -- Llamar a la función original
    RETURN calculate_rent_per_ha(lease_value);
END;
$$ LANGUAGE plpgsql;

-- 3. Verificar que calculate_campaign_closing_date ya esté en v3_calc
-- (Esta función ya debería estar en el esquema v3_calc según la migración 077)

-- Comentarios explicativos
COMMENT ON FUNCTION v3_calc.get_project_dollar_value(BIGINT, VARCHAR) IS 'Función renombrada para seguir patrón v3';
COMMENT ON FUNCTION v3_calc.calculate_rent_per_ha(DOUBLE PRECISION) IS 'Función renombrada para seguir patrón v3';
