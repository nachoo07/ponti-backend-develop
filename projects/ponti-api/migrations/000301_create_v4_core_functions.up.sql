-- =============================================================================
-- MIGRACIÓN 000301: v4_core - Funciones math puras (SIN acceso a tablas)
-- =============================================================================
--
-- Propósito: Funciones matemáticas reutilizables
-- Fecha: 2025-01-XX
-- Autor: Sistema
--

CREATE OR REPLACE FUNCTION v4_core.safe_div(numerator numeric, denominator numeric)
RETURNS numeric LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN denominator IS NULL OR denominator = 0 THEN 0 ELSE numerator / denominator END;
$$;

CREATE OR REPLACE FUNCTION v4_core.per_ha(value numeric, area_ha numeric)
RETURNS numeric LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div(value, area_ha);
$$;

CREATE OR REPLACE FUNCTION v4_core.percentage(part numeric, total numeric)
RETURNS numeric LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN total IS NULL OR total = 0 THEN 0 ELSE (part / total) * 100 END;
$$;

COMMENT ON FUNCTION v4_core.safe_div IS 'División segura: retorna 0 si denominador es NULL o 0';
COMMENT ON FUNCTION v4_core.per_ha IS 'Calcula valor por hectárea usando safe_div';
COMMENT ON FUNCTION v4_core.percentage IS 'Calcula porcentaje (part/total * 100)';
