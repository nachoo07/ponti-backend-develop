-- =============================================================================
-- Migration: 000323_fix_v4_lot_list_missing_columns (DOWN)
-- Rollback: Volver al estado de 000318 (con columnas faltantes)
-- NOTA: Este rollback dejará las vistas rotas como estaban antes
-- =============================================================================

DROP VIEW IF EXISTS v4_report.lot_list CASCADE;
DROP VIEW IF EXISTS v4_report.lot_metrics CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_income CASCADE;
DROP VIEW IF EXISTS v4_calc.lot_base_costs CASCADE;

-- No recreamos las vistas rotas, simplemente las eliminamos
-- Para restaurar el estado anterior, ejecutar 000318 de nuevo
