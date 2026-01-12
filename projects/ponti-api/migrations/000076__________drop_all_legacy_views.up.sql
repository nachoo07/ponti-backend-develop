-- ========================================
-- MIGRATION 000076: DROP ALL LEGACY VIEWS 
-- ========================================
-- 
-- Purpose: Clean slate - remove all legacy views
-- Date: 2025-09-13
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- Eliminar todas las vistas legacy del dashboard
DROP VIEW IF EXISTS dashboard_contributions_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_contributions_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_costs_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_costs_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_harvest_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_harvest_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_sowing_progress_view CASCADE;
DROP VIEW IF EXISTS dashboard_sowing_progress_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view CASCADE;
DROP VIEW IF EXISTS dashboard_management_balance_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_operating_result_view CASCADE;
DROP VIEW IF EXISTS dashboard_operating_result_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_operational_indicators_view CASCADE;
DROP VIEW IF EXISTS dashboard_operational_indicators_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_crop_cost_incidence_view CASCADE;
DROP VIEW IF EXISTS dashboard_crop_incidence_view_v2 CASCADE;
DROP VIEW IF EXISTS dashboard_balance_management_view CASCADE;
DROP VIEW IF EXISTS dashboard_view CASCADE;

-- Eliminar vistas base (serán reemplazadas por funciones v3_calc)
DROP VIEW IF EXISTS base_admin_costs_view CASCADE;
DROP VIEW IF EXISTS base_direct_costs_view CASCADE;
DROP VIEW IF EXISTS base_income_net_view CASCADE;
DROP VIEW IF EXISTS base_lease_calculations_view CASCADE;
DROP VIEW IF EXISTS base_active_total_view CASCADE;
DROP VIEW IF EXISTS base_operating_result_view CASCADE;
DROP VIEW IF EXISTS base_yield_calculations_view CASCADE;

-- Eliminar vistas de workorders y labor
DROP VIEW IF EXISTS workorder_metrics_view CASCADE;
DROP VIEW IF EXISTS workorder_metrics_view_v2 CASCADE;
DROP VIEW IF EXISTS workorder_list_view CASCADE;
DROP VIEW IF EXISTS labor_cards_cube_view CASCADE;
DROP VIEW IF EXISTS labor_cards_cube_view_v2 CASCADE;
DROP VIEW IF EXISTS labor_metrics_view CASCADE;

-- Eliminar vistas de lotes
DROP VIEW IF EXISTS lot_metrics_view CASCADE;
DROP VIEW IF EXISTS lot_table_view CASCADE;
DROP VIEW IF EXISTS fix_labors_list CASCADE;
DROP VIEW IF EXISTS fix_lot_list CASCADE;
DROP VIEW IF EXISTS fix_lots_metrics CASCADE;

-- Eliminar vistas de reportes
DROP VIEW IF EXISTS report_field_crop_metrics_view_v2 CASCADE;
DROP VIEW IF EXISTS investor_contribution_data_view CASCADE;

-- Eliminar vistas auxiliares
DROP VIEW IF EXISTS views_fixes CASCADE;

-- Eliminar vistas adicionales que se crearon en migraciones anteriores
DROP VIEW IF EXISTS dashboard_crop_incidence_view CASCADE;

-- Eliminar cualquier vista que contenga 'lot' en el nombre (por si acaso)
DO $$
DECLARE
    view_name text;
BEGIN
    FOR view_name IN 
        SELECT viewname 
        FROM pg_views 
        WHERE viewname LIKE '%lot%' 
        AND viewname NOT LIKE 'v3_%'
        AND viewname NOT LIKE 'pg_%'
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || view_name || ' CASCADE';
    END LOOP;
END $$;

-- Eliminar cualquier vista que contenga 'dashboard' en el nombre (por si acaso)
DO $$
DECLARE
    view_name text;
BEGIN
    FOR view_name IN 
        SELECT viewname 
        FROM pg_views 
        WHERE viewname LIKE '%dashboard%' 
        AND viewname NOT LIKE 'v3_%'
        AND viewname NOT LIKE 'pg_%'
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || view_name || ' CASCADE';
    END LOOP;
END $$;

-- Eliminar cualquier vista que contenga 'workorder' en el nombre (por si acaso)
DO $$
DECLARE
    view_name text;
BEGIN
    FOR view_name IN 
        SELECT viewname 
        FROM pg_views 
        WHERE viewname LIKE '%workorder%' 
        AND viewname NOT LIKE 'v3_%'
        AND viewname NOT LIKE 'pg_%'
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || view_name || ' CASCADE';
    END LOOP;
END $$;

-- Eliminar cualquier vista que contenga 'labor' en el nombre (por si acaso)
DO $$
DECLARE
    view_name text;
BEGIN
    FOR view_name IN 
        SELECT viewname 
        FROM pg_views 
        WHERE viewname LIKE '%labor%' 
        AND viewname NOT LIKE 'v3_%'
        AND viewname NOT LIKE 'pg_%'
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || view_name || ' CASCADE';
    END LOOP;
END $$;

-- Eliminar cualquier vista que contenga 'report' en el nombre (por si acaso)
DO $$
DECLARE
    view_name text;
BEGIN
    FOR view_name IN 
        SELECT viewname 
        FROM pg_views 
        WHERE viewname LIKE '%report%' 
        AND viewname NOT LIKE 'v3_%'
        AND viewname NOT LIKE 'pg_%'
    LOOP
        EXECUTE 'DROP VIEW IF EXISTS ' || view_name || ' CASCADE';
    END LOOP;
END $$;

COMMIT;