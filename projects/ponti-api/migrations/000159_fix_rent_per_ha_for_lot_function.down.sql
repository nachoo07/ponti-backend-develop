-- ========================================
-- MIGRATION 000159: FIX RENT PER HA FOR LOT FUNCTION (DOWN)
-- ========================================
-- 
-- Purpose: Rollback de corrección de función v3_calc.rent_per_ha_for_lot()
-- Date: 2025-10-21
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- Restaurar función a versión simplificada (la que tenía el bug)
CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint)
 RETURNS double precision
 LANGUAGE plpgsql
AS $function$
DECLARE
    rent_value DOUBLE PRECISION;
BEGIN
    SELECT COALESCE(f.lease_type_value, 0) INTO rent_value
    FROM lots l
    JOIN fields f ON l.field_id = f.id
    WHERE l.id = p_lot_id;
    
    IF rent_value < 0 THEN
        RETURN 0;
    ELSE
        RETURN rent_value;
    END IF;
END;
$function$;

COMMENT ON FUNCTION v3_calc.rent_per_ha_for_lot(bigint) IS 
  'Versión con bug - solo devuelve lease_type_value';

