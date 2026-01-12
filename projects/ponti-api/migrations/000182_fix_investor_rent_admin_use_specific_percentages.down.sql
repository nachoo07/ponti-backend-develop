-- ========================================
-- MIGRACIÓN 000182: FIX Investor Contributions - Use Specific Percentages (DOWN)
-- ========================================
--
-- Propósito: Revertir a la versión anterior (000181)
-- Fecha: 2025-11-08
-- Autor: Sistema

BEGIN;

-- Recrear la vista como estaba en 000181
-- (La versión anterior usaba project_investors.percentage para todo)
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

-- No recreamos aquí porque al hacer migrate down, automáticamente
-- se ejecutará la versión UP de la migración 000181

COMMIT;

