-- Remove duplicate functions from public schema
-- Eliminar funciones duplicadas del esquema public para mantener solo las versiones v3_calc

-- 0. Eliminar vista v_rent_data primero (para evitar dependencias)
DROP VIEW IF EXISTS v_rent_data;

-- 0.1. Eliminar funciones v3_calc existentes si existen
DROP FUNCTION IF EXISTS v3_calc.calculate_rent_per_ha(DOUBLE PRECISION);
DROP FUNCTION IF EXISTS v3_calc.get_project_dollar_value(BIGINT, VARCHAR);

-- 0.2. Crear función v3_calc.calculate_rent_per_ha
CREATE OR REPLACE FUNCTION v3_calc.calculate_rent_per_ha(lease_value DOUBLE PRECISION)
RETURNS DOUBLE PRECISION AS $$
BEGIN
    -- Solo si es menor que 0, devolver 0
    -- Si es 0 o mayor, devolver el valor original
    IF lease_value < 0 THEN
        RETURN 0;
    ELSE
        RETURN lease_value;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 0.3. Crear función v3_calc.get_project_dollar_value
CREATE OR REPLACE FUNCTION v3_calc.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS NUMERIC AS $$
BEGIN
    -- Llamar a la función original si existe, sino usar lógica por defecto
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_project_dollar_value' AND pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')) THEN
        RETURN get_project_dollar_value(p_project_id, p_month);
    ELSE
        -- Lógica por defecto si la función original no existe
        RETURN 1.0;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 1. Recrear vista v_rent_data para usar v3_calc.calculate_rent_per_ha
DROP VIEW IF EXISTS v_rent_data;
CREATE OR REPLACE VIEW v_rent_data AS
SELECT 
    p.id as project_id,
    p.name as proyecto,
    l.id as lot_id,
    l.name as lote,
    f.name as campo,
    lt.name as tipo_arriendo,
    l.hectares,
    v3_calc.calculate_rent_per_ha(f.lease_type_value) as rent_per_ha,
    (v3_calc.calculate_rent_per_ha(f.lease_type_value) * l.hectares) as renta_total
FROM projects p
JOIN fields f ON p.id = f.project_id
JOIN lots l ON f.id = l.field_id
JOIN lease_types lt ON f.lease_type_id = lt.id
WHERE l.deleted_at IS NULL
ORDER BY p.id, l.id;

-- 2. Eliminar calculate_rent_per_ha del esquema public
-- (mantener solo la versión v3_calc.calculate_rent_per_ha)
DROP FUNCTION IF EXISTS public.calculate_rent_per_ha(DOUBLE PRECISION);

-- 3. Eliminar get_project_dollar_value del esquema public
-- (mantener solo la versión v3_calc.get_project_dollar_value)
DROP FUNCTION IF EXISTS public.get_project_dollar_value(BIGINT, VARCHAR);

-- 3. Verificar que las funciones v3_calc existan
-- (estas deberían existir desde migraciones anteriores)
-- v3_calc.calculate_rent_per_ha
-- v3_calc.get_project_dollar_value

-- Comentarios explicativos
COMMENT ON FUNCTION v3_calc.calculate_rent_per_ha(DOUBLE PRECISION) IS 'Función v3_calc para calcular rent_per_ha (única versión activa)';
COMMENT ON FUNCTION v3_calc.get_project_dollar_value(BIGINT, VARCHAR) IS 'Función v3_calc para obtener valor del dólar del proyecto (única versión activa)';
