-- ========================================
-- MIGRATION 000192: Align investor project base rent with rent_fixed_only (UP)
-- ========================================
--
-- Objetivo: Actualizar la vista base del informe de aportes para que utilice
--           v3_lot_ssot.rent_fixed_only_for_lot en todos los cálculos de arriendo.
--
-- Nota: Código en inglés, comentarios en español.

BEGIN;

CREATE OR REPLACE VIEW public.v3_report_investor_project_base AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  COALESCE(SUM(l.hectares), 0::double precision)::numeric AS surface_total_ha,
  COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS lease_fixed_total_usd,
  CASE
    WHEN COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) > 0::double precision THEN TRUE
    ELSE FALSE
  END AS lease_is_fixed,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS lease_per_ha_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS admin_total_usd,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS admin_per_ha_usd
FROM public.projects p
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY
  p.id,
  p.name,
  p.customer_id,
  c.name,
  p.campaign_id,
  cam.name;

COMMENT ON VIEW public.v3_report_investor_project_base IS
'Vista 1/4 para informe de Aportes por Inversor. Contiene datos generales del proyecto: superficie, arriendo fijo y administración. Usa funciones SSOT para cálculos consistentes.';

COMMIT;

