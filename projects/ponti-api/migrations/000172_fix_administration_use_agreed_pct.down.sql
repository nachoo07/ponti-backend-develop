-- ========================================
-- MIGRACIÓN 000172: FIX Administración y Arriendo usan % acordado (DOWN)
-- ========================================
--
-- Revierte los cambios de la migración 000172
-- Restaura la versión anterior de v3_investor_contribution_data_view (de migración 000171)
--
BEGIN;

-- Restaurar versión anterior de v3_investor_contribution_data_view
-- (Esta es la versión de la migración 000171, que calculaba share_pct basándose en aportes reales)

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- Nota: En producción, se recomienda aplicar la migración 000171.up.sql
-- en lugar de duplicar el código aquí

COMMIT;

