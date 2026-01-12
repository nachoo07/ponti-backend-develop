-- ========================================
-- MIGRACIÓN 000116: CREATE v3_dashboard_ssot SCHEMA (DOWN)
-- ========================================
-- 
-- Propósito: Eliminar esquema v3_dashboard_ssot y todas sus funciones
-- Fecha: 2025-10-04
-- Autor: Sistema

-- Eliminar esquema completo con todas sus funciones
DROP SCHEMA IF EXISTS v3_dashboard_ssot CASCADE;

