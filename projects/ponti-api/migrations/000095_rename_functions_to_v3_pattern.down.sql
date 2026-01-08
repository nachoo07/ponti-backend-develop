-- Rollback: Remove v3 function renames

-- Eliminar funciones v3_calc
DROP FUNCTION IF EXISTS v3_calc.get_project_dollar_value(BIGINT, VARCHAR);
DROP FUNCTION IF EXISTS v3_calc.calculate_rent_per_ha(DOUBLE PRECISION);

-- Nota: Las funciones originales se mantienen intactas
