-- =============================================================================
-- MIGRACIÓN 000317: FASE 2 - Corregir bug del dólar en v4_report.labor_list
-- =============================================================================
--
-- BUG IDENTIFICADO:
-- En v3_labor_list: usd_cost_ha = price × dollar → Esto es ARS, no USD
-- En Go: costHaARS = USDCostHa × dollar → Multiplica de nuevo
-- Resultado: doble multiplicación del dólar
--
-- CORRECCIÓN:
-- usd_cost_ha = price (el verdadero USD, sin multiplicar por dólar)
-- Así cuando Go hace: costHaARS = usd_cost_ha × dollar → Obtiene ARS correcto
--
-- NOTA: Este cambio SOLO aplica cuando REPORT_SCHEMA=v4_report
-- v3_labor_list mantiene el bug para compatibilidad
--

DROP VIEW IF EXISTS v4_report.labor_list;

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
  -- FASE 2 FIX: usd_cost_ha ahora es el verdadero USD (price sin multiplicar)
  -- Go hará: costHaARS = usd_cost_ha × dollar = price × dollar = ARS correcto
  lb.price::numeric AS usd_cost_ha,
  -- usd_net_total también corregido: price × area (USD total)
  (lb.price * w.effective_area)::numeric AS usd_net_total,
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
'FASE 2: Bug dólar corregido. usd_cost_ha = price (USD real). Go calcula ARS correctamente.';
