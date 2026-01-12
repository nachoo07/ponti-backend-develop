-- ========================================
-- MIGRACIÓN 000068: REVERTIR PRINCIPIOS DRY, SSOT, VIEW COMPOSITION Y ENCAPSULATION
-- Entidad: views (Revertir principios de ingeniería de software)
-- Funcionalidad: Revertir la aplicación de principios DRY, SSOT, View Composition y Encapsulation
-- ========================================

-- ========================================
-- 1. ELIMINAR VISTAS DERIVADAS
-- ========================================
DROP VIEW IF EXISTS dashboard_harvest_progress_derived_view;
DROP VIEW IF EXISTS dashboard_sowing_progress_derived_view;
DROP VIEW IF EXISTS lot_metrics_derived_view;
DROP VIEW IF EXISTS workorder_metrics_derived_view;

-- ========================================
-- 2. ELIMINAR VISTAS BASE
-- ========================================
DROP VIEW IF EXISTS base_lot_calculations_view;
DROP VIEW IF EXISTS base_workorder_calculations_view;

-- ========================================
-- 3. ELIMINAR FUNCIONES DE NEGOCIO
-- ========================================
DROP FUNCTION IF EXISTS calculate_cost_per_ha(DECIMAL, DECIMAL);
DROP FUNCTION IF EXISTS calculate_yield(DECIMAL, DECIMAL);
DROP FUNCTION IF EXISTS calculate_harvested_area(DECIMAL, DECIMAL);
DROP FUNCTION IF EXISTS calculate_sowed_area(DATE, DECIMAL);
DROP FUNCTION IF EXISTS calculate_supply_cost(DECIMAL, DECIMAL, DECIMAL);
DROP FUNCTION IF EXISTS calculate_labor_cost(DECIMAL, DECIMAL);

-- ========================================
-- 4. ELIMINAR TABLA DE DOCUMENTACIÓN
-- ========================================
DROP TABLE IF EXISTS engineering_principles_documentation;
