-- ========================================
-- MIGRACIÓN 000174: FIX Investor Report Duplicate Rows (DOWN)
-- ========================================
--
-- Propósito: Revertir corrección de duplicación
-- Fecha: 2025-11-03
-- Autor: Sistema

BEGIN;

-- Restaurar la vista a la versión de la migración 000173
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- Nota: Para revertir completamente, se debe ejecutar la migración 000173 nuevamente
-- o restaurar desde un backup de la base de datos

COMMIT;

