-- Rollback: Restore duplicate functions to public schema
-- Revierte: Restaurar funciones duplicadas al esquema public

-- 1. Restaurar calculate_rent_per_ha en esquema public
-- (copiar desde v3_calc.calculate_rent_per_ha)
CREATE OR REPLACE FUNCTION public.calculate_rent_per_ha(lease_value DOUBLE PRECISION)
RETURNS DOUBLE PRECISION AS $$
BEGIN
    -- Llamar a la función v3_calc
    RETURN v3_calc.calculate_rent_per_ha(lease_value);
END;
$$ LANGUAGE plpgsql;

-- 2. Restaurar get_project_dollar_value en esquema public
-- (copiar desde v3_calc.get_project_dollar_value)
CREATE OR REPLACE FUNCTION public.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS DECIMAL AS $$
BEGIN
    -- Llamar a la función v3_calc
    RETURN v3_calc.get_project_dollar_value(p_project_id, p_month);
END;
$$ LANGUAGE plpgsql;

-- Comentarios explicativos
COMMENT ON FUNCTION public.calculate_rent_per_ha(DOUBLE PRECISION) IS 'Función wrapper que llama a v3_calc.calculate_rent_per_ha';
COMMENT ON FUNCTION public.get_project_dollar_value(BIGINT, VARCHAR) IS 'Función wrapper que llama a v3_calc.get_project_dollar_value';
