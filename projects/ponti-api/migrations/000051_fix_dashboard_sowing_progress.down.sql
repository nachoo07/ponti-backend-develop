-- Migración 000050: Revertir corrección del avance de siembra
-- Enfoque: Eliminar la vista dashboard_sowing_progress_view

DROP VIEW IF EXISTS dashboard_sowing_progress_view;
