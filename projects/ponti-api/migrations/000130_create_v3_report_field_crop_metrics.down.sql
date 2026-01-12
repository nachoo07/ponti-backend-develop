-- ========================================
-- MIGRACIÓN 000130: DROP v3_report_field_crop VIEWS (DOWN)
-- ========================================

BEGIN;

-- Eliminar vista principal primero (depende de las otras)
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics CASCADE;

-- Eliminar las 5 vistas específicas
DROP VIEW IF EXISTS public.v3_report_field_crop_cultivos CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_labores CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_insumos CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_economicos CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_rentabilidad CASCADE;

-- Eliminar vista antigua si existe
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics_view CASCADE;

COMMIT;
