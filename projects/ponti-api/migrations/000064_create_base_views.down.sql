-- ========================================
-- MIGRACIÓN 000070: CREAR VISTAS BASE REUTILIZABLES (ROLLBACK)
-- Entidad: base_views (Vistas base para evitar duplicación)
-- Funcionalidad: Eliminar vistas base creadas
-- ========================================

-- Eliminar todas las vistas base creadas
DROP VIEW IF EXISTS base_operating_result_view;
DROP VIEW IF EXISTS base_active_total_view;
DROP VIEW IF EXISTS base_lease_calculations_view;
DROP VIEW IF EXISTS base_admin_costs_view;
DROP VIEW IF EXISTS base_income_net_view;
DROP VIEW IF EXISTS base_yield_calculations_view;
DROP VIEW IF EXISTS base_direct_costs_view;
