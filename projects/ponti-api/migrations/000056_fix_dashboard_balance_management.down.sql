-- ========================================
-- MIGRACIÓN 000056: FIX BALANCE DE GESTIÓN - ROLLBACK
-- ========================================
-- 
-- Objetivo: Eliminar vista del balance de gestión
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Eliminar vista del balance de gestión
DROP VIEW IF EXISTS dashboard_balance_management_view;
