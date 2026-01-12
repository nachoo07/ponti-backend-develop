-- ========================================
-- MIGRACIÓN 000058: FIX DASHBOARD MANAGEMENT BALANCE VIEW (ROLLBACK)
-- ========================================
-- 
-- Objetivo: Revertir cambios de la migración 000058
-- Fecha: 2025-01-27
-- Autor: Sistema

-- Eliminar la vista corregida
DROP VIEW IF EXISTS dashboard_management_balance_view;

-- Recrear la vista anterior (si existe)
-- Nota: Esta vista se recreará automáticamente en migraciones anteriores
