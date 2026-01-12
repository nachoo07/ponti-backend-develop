-- ========================================
-- MIGRACIÓN 000131: FIX v3_report_field_crop_insumos - Add Fertilizantes y Otros Insumos (DOWN)
-- ========================================
-- 
-- Propósito: Revertir cambios de la migración 000131
-- Fecha: 2025-10-29
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- Revertir a la versión anterior de 000130
-- Las vistas se recrearán automáticamente cuando se ejecute el down de 000131
-- y el up de 000130 para volver al estado anterior

-- Eliminar vistas en orden inverso
DROP VIEW IF EXISTS public.v3_report_field_crop_metrics CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_rentabilidad CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_economicos CASCADE;
DROP VIEW IF EXISTS public.v3_report_field_crop_insumos CASCADE;

-- Mensaje de revert
-- Para volver al estado anterior, ejecutar: migrate down 1 && migrate up

COMMIT;

