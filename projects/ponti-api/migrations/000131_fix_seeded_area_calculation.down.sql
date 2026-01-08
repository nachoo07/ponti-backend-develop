-- ========================================
-- MIGRACIÓN 000131: FIX CÁLCULO DE SUPERFICIE SEMBRADA (DOWN)
-- ========================================
-- 
-- Propósito: Revertir el cambio en v3_lot_ssot.seeded_area_for_lot
-- Restaurar la función original que usaba sowing_date

BEGIN;

-- ========================================
-- RESTAURAR FUNCIÓN ORIGINAL
-- ========================================
-- Lógica original: Si sowing_date IS NOT NULL → usar hectares, sino 0

CREATE OR REPLACE FUNCTION v3_lot_ssot.seeded_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

COMMIT;

-- Restaurar comentario original
COMMENT ON FUNCTION v3_lot_ssot.seeded_area_for_lot(bigint) IS 
  'Calcula área sembrada basada en fecha de siembra';

