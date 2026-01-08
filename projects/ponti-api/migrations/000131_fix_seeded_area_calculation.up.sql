-- ========================================
-- MIGRACIÓN 000131: FIX CÁLCULO DE SUPERFICIE SEMBRADA (UP)
-- ========================================
-- 
-- Propósito: Corregir v3_lot_ssot.seeded_area_for_lot para que funcione sin sowing_date
-- Problema: La función original requería sowing_date, pero los lotes no tienen este campo
-- Solución: Usar workorders de siembra o hectares del lote como fallback
-- Fecha: 2025-10-11
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- REEMPLAZAR FUNCIÓN seeded_area_for_lot
-- ========================================
-- Nueva lógica:
-- 1. Si hay workorders de siembra → usar suma de effective_area
-- 2. Si no hay workorders → usar hectares del lote
-- 3. Fallback → 0

CREATE OR REPLACE FUNCTION v3_lot_ssot.seeded_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Opción 1: Si hay órdenes de trabajo de siembra, usar suma de effective_area
    (SELECT SUM(w.effective_area)
     FROM public.workorders w
     JOIN public.labors lab ON lab.id = w.labor_id
     JOIN public.categories cat ON cat.id = lab.category_id
     WHERE w.lot_id = p_lot_id 
       AND w.deleted_at IS NULL
       AND cat.name = 'Siembra'
       AND cat.type_id = 4
       AND w.effective_area > 0),
    -- Opción 2: Si no hay órdenes de siembra, usar hectares del lote
    (SELECT l.hectares
     FROM public.lots l
     WHERE l.id = p_lot_id AND l.deleted_at IS NULL),
    -- Fallback: 0
    0
  )
$$;

COMMIT;

-- Comentario sobre el cambio
COMMENT ON FUNCTION v3_lot_ssot.seeded_area_for_lot(bigint) IS 
  'Calcula área sembrada: prioriza workorders de siembra, fallback a hectares del lote';

