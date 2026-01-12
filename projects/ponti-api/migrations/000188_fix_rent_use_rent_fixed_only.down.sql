-- ========================================
-- MIGRATION 000188: FIX RENT CALCULATION IN INVESTOR CONTRIBUTION (DOWN)
-- ========================================
-- 
-- Purpose: Revertir cambio quirúrgico - volver a v3_calc.rent_per_ha_for_lot()
-- Date: 2025-11-08
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- Volver a la versión original con v3_calc.rent_per_ha_for_lot()
CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id
      JOIN public.categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
)
SELECT
  lb.project_id,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla')
  ), 0)::numeric AS seeds_total_usd,
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name IN ('Pulverización', 'Otras Labores')
  ), 0)::numeric AS general_labors_total_usd,
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.name = 'Siembra'
      AND cat.type_id = 4
  ), 0)::numeric AS sowing_total_usd,
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.name = 'Riego'
      AND cat.type_id = 4
  ), 0)::numeric AS irrigation_total_usd,
  -- ROLLBACK: Volver a v3_calc.rent_per_ha_for_lot()
  COALESCE(SUM(
    v3_calc.rent_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS rent_capitalizable_total_usd,
  COALESCE(SUM(
    v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS administration_total_usd,
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha
FROM lot_base lb
GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
  'Vista 2/4 para informe de Aportes por Inversor. Contiene todas las categorías de aportes. Usa funciones SSOT.';

COMMIT;

