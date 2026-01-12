-- ========================================
-- MIGRATION 000154: REMOVE ROUND FROM PERCENTAGE FUNCTION (UP)
-- ========================================
-- 
-- Purpose: Eliminar ROUND() de la función percentage_rounded para mantener precisión completa
-- Date: 2025-01-27
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.
-- REGLA CRÍTICA: NUNCA usar ROUND() en migraciones SQL, mantener precisión completa

-- ========================================
-- FUNCIÓN MEJORADA PARA CÁLCULO DE PORCENTAJES SIN REDONDEO
-- ========================================
-- Nota: Eliminar redondeo para mantener precisión completa
CREATE OR REPLACE FUNCTION v3_calc.percentage_rounded(numeric, numeric) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2) * 100
$$;
