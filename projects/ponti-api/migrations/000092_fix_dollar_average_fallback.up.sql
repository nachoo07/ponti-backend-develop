-- ========================================
-- MIGRATION 000092: FIX dollar_average_for_month fallback (UP)
-- ========================================
--
-- Purpose: Fix dollar_average_for_month to use get_default_fx_rate as fallback
--          when no project-specific data is available
-- Date: 2025-01-27
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- Corregir función dollar_average_for_month para usar fallback
CREATE OR REPLACE FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT average_value 
     FROM public.project_dollar_values 
     WHERE project_id = p_project_id 
       AND month = TO_CHAR(p_date, 'YYYY-MM')  -- Convierte fecha a formato YYYY-MM
       AND deleted_at IS NULL
     LIMIT 1), 
    get_default_fx_rate()  -- Usar valor por defecto como fallback
  )::numeric
$$;
