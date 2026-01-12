-- =============================================================================
-- MIGRACIÓN 000304: v4_calc.lot_base_income - Para cuando haya producción
-- =============================================================================
--
-- Propósito: Vista para producción/ingresos (usar cuando haya datos)
-- Fecha: 2025-01-XX
-- Autor: Sistema
--

CREATE OR REPLACE VIEW v4_calc.lot_base_income AS
SELECT
  l.id AS lot_id,
  COALESCE(l.tons, 0)::numeric AS production_tn,
  COALESCE(v3_report_ssot.board_price_for_lot(l.id), 0)::numeric AS board_price_usd_tn,
  COALESCE(v3_report_ssot.freight_cost_for_lot(l.id), 0)::numeric AS freight_usd_tn,
  COALESCE(v3_report_ssot.commercial_cost_for_lot(l.id), 0)::numeric AS commercial_usd_tn,
  COALESCE(v3_lot_ssot.net_price_usd_for_lot(l.id), 0)::numeric AS net_price_usd_tn
FROM public.lots l
WHERE l.deleted_at IS NULL;

COMMENT ON VIEW v4_calc.lot_base_income IS 
'Vista para producción/ingresos. Usar cuando haya datos de producción.';
