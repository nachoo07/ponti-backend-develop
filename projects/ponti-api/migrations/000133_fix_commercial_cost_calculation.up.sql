-- Migration: 000134_fix_commercial_cost_calculation
-- Description: Corrige la función commercial_cost_for_lot para calcular el valor en USD/TN
--              en lugar de devolver el porcentaje.
--              Fórmula: Precio Bruto × (% Comercial / 100)
-- Author: AI Assistant
-- Date: 2025-10-11

BEGIN;

-- =======================================================================================
-- CORRECCIÓN: Función commercial_cost_for_lot debe devolver USD/TN, no el porcentaje
-- =======================================================================================

-- ANTES: Devolvía cc.commercial_cost (el porcentaje: 2, 3, etc.)
-- AHORA: Calcula board_price * (commercial_cost / 100) para obtener el valor en USD/TN

CREATE OR REPLACE FUNCTION v3_report_ssot.commercial_cost_for_lot(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  -- Calcula el costo comercial en USD/TN = Precio Bruto × (% Comercial / 100)
  -- Ejemplo: Si precio bruto = 220 y comercial = 2%, entonces 220 × 0.02 = 4.4 USD/TN
  SELECT COALESCE(cc.board_price * (cc.commercial_cost / 100.0), 0)
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

-- Actualizar comentario de la función
COMMENT ON FUNCTION v3_report_ssot.commercial_cost_for_lot(bigint) IS 
  'Obtiene costo comercial calculado en USD/tn por lote. Calcula: board_price × (commercial_cost_pct / 100)';

COMMIT;

