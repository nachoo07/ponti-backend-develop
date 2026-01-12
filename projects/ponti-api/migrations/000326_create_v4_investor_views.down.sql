-- ========================================
-- MIGRATION 000326: Create v4_report investor views with bug fixes (DOWN)
-- ========================================
--
-- Rollback: Eliminar vistas de investor en v4_report
--
-- Nota: Código en inglés, comentarios en español.

BEGIN;

-- Eliminar en orden inverso de dependencias
DROP VIEW IF EXISTS v4_report.investor_contribution_data;
DROP VIEW IF EXISTS v4_report.investor_distributions;
DROP VIEW IF EXISTS v4_report.investor_contribution_categories;
DROP VIEW IF EXISTS v4_report.investor_project_base;

COMMIT;
