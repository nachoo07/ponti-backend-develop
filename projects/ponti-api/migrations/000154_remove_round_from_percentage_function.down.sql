-- ========================================
-- MIGRATION 000154: REMOVE ROUND FROM PERCENTAGE FUNCTION (DOWN)
-- ========================================
-- 
-- Purpose: Revertir eliminación de ROUND() en la función percentage_rounded
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- ========================================
-- REVERTIR FUNCIÓN DE PORCENTAJES CON REDONDEO
-- ========================================
-- Nota: Volver a usar redondeo a 3 decimales
CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT ROUND(v3_calc.safe_div($1, $2) * 100, 3)
$$;
