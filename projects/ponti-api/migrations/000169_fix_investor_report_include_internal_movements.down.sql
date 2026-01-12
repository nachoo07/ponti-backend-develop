-- ========================================
-- MIGRACIÓN 000169: FIX Informe de Aportes - Quitar filtro quantity > 0 (DOWN)
-- ========================================
-- Revertir a la versión anterior (con filtro quantity > 0)

BEGIN;

CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  SELECT 
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM workorders w
      JOIN labors lab ON w.labor_id = lab.id
      JOIN categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0) AS seeded_area_ha
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
)
SELECT 
  lb.project_id,
  
  -- AGROQUÍMICOS (con filtro quantity > 0)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0  -- RESTAURADO: filtro original
      AND cat.type_id = 2
      AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  -- FERTILIZANTES (con filtro quantity > 0)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0  -- RESTAURADO: filtro original
      AND cat.type_id = 3
      AND cat.name = 'Fertilizantes'
  ), 0)::numeric AS fertilizers_total_usd,
  
  -- SEMILLAS (con filtro quantity > 0)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0  -- RESTAURADO: filtro original
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0)::numeric AS seeds_total_usd,
  
  -- LABORES GENERALES (sin cambios)
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
  
  -- SIEMBRA (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Siembra'
  ), 0)::numeric AS sowing_total_usd,
  
  -- RIEGO (sin cambios)
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name = 'Riego'
  ), 0)::numeric AS irrigation_total_usd,
  
  -- ARRIENDO CAPITALIZABLE (usando v3_calc functions - actual)
  COALESCE(SUM(
    CASE 
      WHEN f.lease_type_id IN (3, 4) THEN v3_calc.rent_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
      ELSE 0
    END
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- ADMINISTRACIÓN (usando v3_calc functions - actual)
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares), 0)::numeric AS administration_total_usd,
  
  -- SUPERFICIE SEMBRADA TOTAL (sin cambios)
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb
JOIN fields f ON f.id = (SELECT field_id FROM lots WHERE id = lb.lot_id)
GROUP BY lb.project_id;

COMMIT;
