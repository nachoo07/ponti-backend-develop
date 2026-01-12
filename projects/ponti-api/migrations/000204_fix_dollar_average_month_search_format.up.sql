-- ========================================
-- MIGRATION 000204: FIX dollar_average_for_month search format (UP)
-- ========================================
--
-- Purpose: Fix dollar_average_for_month to search by year and month name in Spanish
--          instead of YYYY-MM format, since project_dollar_values stores month
--          as Spanish names ('Enero', 'Febrero', etc.)
-- Date: 2025-11-18
-- Author: System
--
-- Note: Code in English, comments in Spanish.

-- Corregir función dollar_average_for_month para buscar por año y nombre de mes en español
CREATE OR REPLACE FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT average_value 
     FROM public.project_dollar_values 
     WHERE project_id = p_project_id 
       AND year = EXTRACT(YEAR FROM p_date)::bigint  -- Año de la fecha
       AND month = CASE EXTRACT(MONTH FROM p_date)    -- Convertir número de mes a nombre en español
         WHEN 1 THEN 'Enero'
         WHEN 2 THEN 'Febrero'
         WHEN 3 THEN 'Marzo'
         WHEN 4 THEN 'Abril'
         WHEN 5 THEN 'Mayo'
         WHEN 6 THEN 'Junio'
         WHEN 7 THEN 'Julio'
         WHEN 8 THEN 'Agosto'
         WHEN 9 THEN 'Septiembre'
         WHEN 10 THEN 'Octubre'
         WHEN 11 THEN 'Noviembre'
         WHEN 12 THEN 'Diciembre'
       END
       AND deleted_at IS NULL
     LIMIT 1), 
    get_default_fx_rate()  -- Usar valor por defecto como fallback
  )::numeric
$$;

