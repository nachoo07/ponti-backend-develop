-- Rollback: Remove rent_per_ha negative values fix

-- Eliminar vista
DROP VIEW IF EXISTS v_rent_data;

-- Eliminar función wrapper
DROP FUNCTION IF EXISTS v3_calc.rent_per_ha_for_lot_fixed(BIGINT);

-- Eliminar función helper
DROP FUNCTION IF EXISTS calculate_rent_per_ha(DOUBLE PRECISION);

-- Nota: No revertimos los valores actualizados porque no sabemos cuáles eran los originales
-- Si necesitas revertir, deberías tener un backup de los datos originales
-- Tampoco podemos revertir la función original porque no tenemos el código original