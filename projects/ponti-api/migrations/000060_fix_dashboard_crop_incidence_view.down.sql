-- ========================================
-- MIGRACIÓN 000059: FIX DASHBOARD CROP INCIDENCE VIEW (ROLLBACK)
-- ========================================
-- 
-- Objetivo: Revertir cambios de la migración 000059
-- Fecha: 2025-01-27
-- Autor: Sistema

-- Eliminar la vista corregida
DROP VIEW IF EXISTS dashboard_crop_incidence_view;
