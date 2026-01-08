-- ========================================
-- MIGRACIÓN 000113: CREATE v3_core_ssot SCHEMA (DOWN)
-- ========================================
-- 
-- Propósito: Eliminar esquema v3_core_ssot y todas sus funciones
-- Fecha: 2025-10-04
-- Autor: Sistema

-- Eliminar esquema completo con todas sus funciones
DROP SCHEMA IF EXISTS v3_core_ssot CASCADE;

