-- ========================================
-- MIGRACIÓN 000071: ACTUALIZAR VISTAS PARA USAR VISTAS BASE (ROLLBACK)
-- Entidad: views (Actualizar vistas existentes)
-- Funcionalidad: Revertir actualizaciones de vistas
-- ========================================

-- Restaurar vistas originales (desde migraciones anteriores)
-- Nota: Estas vistas se restaurarán a su estado anterior a la migración 000071

-- Restaurar fix_lot_list desde migración 000068
DROP VIEW IF EXISTS fix_lot_list;
-- (La vista original se restaurará desde la migración 000068)

-- Restaurar fix_lots_metrics desde migración 000067
DROP VIEW IF EXISTS fix_lots_metrics;
-- (La vista original se restaurará desde la migración 000067)

-- Restaurar dashboard_operating_result_view desde migración 000069
DROP VIEW IF EXISTS dashboard_operating_result_view;
-- (La vista original se restaurará desde la migración 000069)

-- Restaurar dashboard_management_balance_view desde migración 000069
DROP VIEW IF EXISTS dashboard_management_balance_view;
-- (La vista original se restaurará desde la migración 000069)
