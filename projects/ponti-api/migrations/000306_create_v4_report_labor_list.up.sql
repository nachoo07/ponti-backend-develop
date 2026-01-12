-- =============================================================================
-- MIGRACIÓN 000306: v4_report.labor_list - Paridad exacta con v3_labor_list
-- =============================================================================
--
-- Propósito: Vista con paridad exacta con v3_labor_list (000089)
-- Fecha: 2025-01-XX
-- Autor: Sistema
--
-- FASE 1: Replica naming incorrecto (usd_cost_ha es ARS, no USD)
-- BUG CONOCIDO: usd_cost_ha = price × dollar = ARS, no USD
-- FIX EN FASE 2: Renombrar a cost_per_ha_ars
--

CREATE OR REPLACE VIEW v4_report.labor_list AS
SELECT
  w.id AS workorder_id,
  w.number AS workorder_number,
  w.date,
  w.project_id,
  p.name AS project_name,
  w.field_id,
  f.name AS field_name,
  w.lot_id,
  l.name AS lot_name,
  w.crop_id,
  c.name AS crop_name,
  w.labor_id,
  lb.name AS labor_name,
  lb.category_id AS labor_category_id,
  cat.name AS labor_category_name,
  w.contractor,
  lb.contractor_name,
  w.effective_area AS surface_ha,
  lb.price AS cost_per_ha,
  (lb.price * w.effective_area)::numeric AS total_labor_cost,
  v3_calc.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
  -- FASE 1: Replicar naming incorrecto de v3 para paridad
  -- BUG CONOCIDO: usd_cost_ha es realmente ARS (price × dollar)
  (lb.price * v3_calc.dollar_average_for_month(w.project_id, w.date))::numeric AS usd_cost_ha,
  (lb.price * v3_calc.dollar_average_for_month(w.project_id, w.date) * w.effective_area)::numeric AS usd_net_total,
  w.investor_id,
  i.name AS investor_name
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id AND p.deleted_at IS NULL
JOIN public.fields f ON f.id = w.field_id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.id = w.lot_id AND l.deleted_at IS NULL
LEFT JOIN public.crops c ON c.id = w.crop_id AND c.deleted_at IS NULL
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
LEFT JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
LEFT JOIN public.investors i ON i.id = w.investor_id AND i.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL;

COMMENT ON VIEW v4_report.labor_list IS 
'Paridad exacta con v3_labor_list (000089). FASE 1: Mantiene bug de naming (usd_cost_ha es ARS).';
