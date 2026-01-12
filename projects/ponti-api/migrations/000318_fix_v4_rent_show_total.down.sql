-- =============================================================================
-- MIGRACIÓN 000318 DOWN: Revertir a mostrar rent_fixed (FASE 1 paridad)
-- =============================================================================

-- Revertir ejecutando las migraciones anteriores
-- 000303, 000304, 000305, 000307, 000311, 000316

-- Por simplicidad, dropear las vistas y recrearlas requiere re-ejecutar migraciones
-- En producción, hacer migrate down hasta 000317 y luego up hasta 000317

DROP VIEW IF EXISTS v4_report.field_crop_metrics CASCADE;
DROP VIEW IF EXISTS v4_report.field_crop_economicos CASCADE;
DROP VIEW IF EXISTS v4_report.lot_list CASCADE;
DROP VIEW IF EXISTS v4_report.lot_metrics CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_income CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_costs CASCADE;

-- NOTA: Para restaurar completamente, ejecutar:
-- migrate down -all
-- migrate up to 000317
