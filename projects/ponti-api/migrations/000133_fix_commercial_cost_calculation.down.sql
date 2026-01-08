-- Migration: 000134_fix_commercial_cost_calculation (ROLLBACK)
-- Description: Revierte la corrección de commercial_cost_for_lot a su versión original
-- Author: AI Assistant
-- Date: 2025-10-11

BEGIN;

-- =======================================================================================
-- ROLLBACK: Restaurar función commercial_cost_for_lot a su versión original (incorrecta)
-- =======================================================================================

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

-- Restaurar comentario original
COMMENT ON FUNCTION v3_report_ssot.commercial_cost_for_lot(bigint) IS 'Obtiene costo comercial USD/tn por lote';

COMMIT;

