-- ========================================
-- MIGRACIÓN 000205: RECREATE v3_dashboard_metrics after 000202 (DOWN)
-- ========================================
--
-- Propósito: Eliminar v3_dashboard_metrics creada en migración 000205
-- Fecha: 2025-11-18
-- Autor: Sistema
--
-- Note: Código en inglés, comentarios en español

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_metrics;

COMMIT;

