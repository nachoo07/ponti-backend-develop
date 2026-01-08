-- =====================================================
-- 000063: SHARED - Índices de Soporte para Cálculos
-- =====================================================
-- Entidad: shared (Compartido)
-- Funcionalidad: Crear índices de soporte para optimizar cálculos
-- =====================================================

-- Índices para optimizar cálculos de workorders
CREATE INDEX IF NOT EXISTS idx_workorder_items_workorder_id 
ON workorder_items(workorder_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_items_supply_id 
ON workorder_items(supply_id) WHERE deleted_at IS NULL;

-- Índices para optimizar cálculos de labors
CREATE INDEX IF NOT EXISTS idx_labors_project_id 
ON labors(project_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_labors_category_id 
ON labors(category_id) WHERE deleted_at IS NULL;

-- Índices para optimizar cálculos de lots
CREATE INDEX IF NOT EXISTS idx_lots_field_id 
ON lots(field_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lots_current_crop_id 
ON lots(current_crop_id) WHERE deleted_at IS NULL;

-- Índices para optimizar cálculos de commercialization
CREATE INDEX IF NOT EXISTS idx_crop_commercializations_project_id 
ON crop_commercializations(project_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_crop_commercializations_crop_id 
ON crop_commercializations(crop_id) WHERE deleted_at IS NULL;

-- Índices para optimizar cálculos de project
CREATE INDEX IF NOT EXISTS idx_projects_id 
ON projects(id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_project_dollar_values_project_id 
ON project_dollar_values(project_id) WHERE deleted_at IS NULL;

-- Comentarios en español
COMMENT ON INDEX idx_workorder_items_workorder_id IS 'Índice para optimizar cálculos de workorder items por workorder_id';
COMMENT ON INDEX idx_workorder_items_supply_id IS 'Índice para optimizar cálculos de workorder items por supply_id';
COMMENT ON INDEX idx_labors_project_id IS 'Índice para optimizar cálculos de labors por project_id';
COMMENT ON INDEX idx_labors_category_id IS 'Índice para optimizar cálculos de labors por category_id';
COMMENT ON INDEX idx_lots_field_id IS 'Índice para optimizar cálculos de lots por field_id';
COMMENT ON INDEX idx_crop_commercializations_project_id IS 'Índice para optimizar cálculos de commercialization por project_id';
