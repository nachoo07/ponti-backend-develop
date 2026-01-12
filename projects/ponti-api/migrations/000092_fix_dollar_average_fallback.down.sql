-- ========================================
-- MIGRATION 000092: FIX dollar_average_for_month fallback (DOWN)
-- ========================================
--
-- Purpose: Revert dollar_average_for_month to original implementation
-- Date: 2025-01-27
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- Revertir función dollar_average_for_month a implementación original
CREATE OR REPLACE FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT average_value 
     FROM public.project_dollar_values 
     WHERE project_id = p_project_id 
       AND month = TO_CHAR(p_date, 'YYYY-MM')  -- Convierte fecha a formato YYYY-MM
       AND deleted_at IS NULL
     LIMIT 1), 0  -- Si no hay datos para ese mes, retorna 0
  )::numeric
$$;
