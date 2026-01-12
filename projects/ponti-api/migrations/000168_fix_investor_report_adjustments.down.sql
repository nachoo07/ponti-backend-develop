-- ========================================
-- MIGRACIÓN 000168: FIX INVESTOR REPORT ADJUSTMENTS (DOWN)
-- ========================================
-- 
-- Rollback: Restaurar las vistas a la versión anterior (migración 000167)
--

BEGIN;

-- Eliminar las vistas en orden inverso
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;

-- Nota: Las vistas se recrearán automáticamente al aplicar la migración 000167
-- No es necesario recrearlas aquí, ya que el sistema de migraciones
-- las restaurará al estado anterior.

COMMIT;

