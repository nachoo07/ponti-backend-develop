-- ========================================
-- MIGRACIÓN 000058: FIX INDICADORES OPERATIVOS - ROLLBACK
-- ========================================
-- 
-- Objetivo: Eliminar vista de indicadores operativos
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Eliminar vista de indicadores operativos
DROP VIEW IF EXISTS dashboard_operational_indicators_view;
