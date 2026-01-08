-- =====================================================
-- 000063: SHARED - Índices de Soporte para Cálculos (ROLLBACK)
-- =====================================================
-- Entidad: shared (Compartido)
-- Funcionalidad: Eliminar índices de soporte para cálculos
-- =====================================================

-- Eliminar índices de workorders
DROP INDEX IF EXISTS idx_workorder_items_workorder_id;
DROP INDEX IF EXISTS idx_workorder_items_supply_id;

-- Eliminar índices de labors
DROP INDEX IF EXISTS idx_labors_project_id;
DROP INDEX IF EXISTS idx_labors_field_id;
DROP INDEX IF EXISTS idx_labors_labor_category_id;

-- Eliminar índices de lots
DROP INDEX IF EXISTS idx_lots_project_id;
DROP INDEX IF EXISTS idx_lots_field_id;
DROP INDEX IF EXISTS idx_lots_current_crop_id;

-- Eliminar índices de commercialization
DROP INDEX IF EXISTS idx_crop_commercializations_project_id;
DROP INDEX IF EXISTS idx_crop_commercializations_crop_id;

-- Eliminar índices de project
DROP INDEX IF EXISTS idx_projects_id;
DROP INDEX IF EXISTS idx_project_dollar_values_project_id;
