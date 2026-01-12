-- ========================================
-- MIGRACIÓN 000129: CREATE v3_report_ssot SCHEMA (UP)
-- ========================================
-- 
-- Propósito: Crear esquema v3_report_ssot con funciones para reportes
-- Dependencias: Requiere v3_core_ssot (000113), v3_lot_ssot (000115)
-- Alcance: 3 funciones (precios de comercialización)
-- Fecha: 2025-10-09
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español
-- Usa v3_lot_ssot para reutilizar cálculos existentes

BEGIN;

-- ========================================
-- CREAR ESQUEMA v3_report_ssot
-- ========================================
CREATE SCHEMA IF NOT EXISTS v3_report_ssot;

COMMENT ON SCHEMA v3_report_ssot IS 'Funciones SSOT específicas de reportes: precios y cálculos exclusivos';

-- ========================================
-- GRUPO 1: PRECIOS DE COMERCIALIZACIÓN (3 funciones)
-- ========================================
-- Propósito: Obtener precios desde crop_commercializations por lote
-- Nota: net_price ya existe en v3_lot_ssot.net_price_usd_for_lot()

-- 1.1: Precio bruto (precio pizarra)
CREATE OR REPLACE FUNCTION v3_report_ssot.board_price_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.board_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

-- 1.2: Costo de flete
CREATE OR REPLACE FUNCTION v3_report_ssot.freight_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.freight_cost, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

-- 1.3: Costo comercial (gastos comerciales)
CREATE OR REPLACE FUNCTION v3_report_ssot.commercial_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(cc.commercial_cost, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;

COMMIT;

-- Comentarios sobre funciones
COMMENT ON FUNCTION v3_report_ssot.board_price_for_lot(bigint) IS 'Obtiene precio bruto (pizarra) USD/tn por lote';
COMMENT ON FUNCTION v3_report_ssot.freight_cost_for_lot(bigint) IS 'Obtiene costo de flete USD/tn por lote';
COMMENT ON FUNCTION v3_report_ssot.commercial_cost_for_lot(bigint) IS 'Obtiene costo comercial USD/tn por lote';
