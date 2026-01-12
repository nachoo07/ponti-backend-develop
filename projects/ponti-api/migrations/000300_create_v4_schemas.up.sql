-- =============================================================================
-- MIGRACIÓN 000300: Crear schemas v4
-- =============================================================================
--
-- Propósito: Crear los schemas necesarios para la arquitectura v4
-- Fecha: 2025-01-XX
-- Autor: Sistema
--

CREATE SCHEMA IF NOT EXISTS v4_core;
CREATE SCHEMA IF NOT EXISTS v4_ssot;
CREATE SCHEMA IF NOT EXISTS v4_calc;
CREATE SCHEMA IF NOT EXISTS v4_report;

COMMENT ON SCHEMA v4_core IS 'Funciones math puras sin acceso a tablas';
COMMENT ON SCHEMA v4_ssot IS 'FASE 1: Wrappers a v3. FASE 2: Reimplementación';
COMMENT ON SCHEMA v4_calc IS 'Vistas de cálculo 1-fila-por-entidad';
COMMENT ON SCHEMA v4_report IS 'Vistas con contrato para Go (paridad con v3)';
