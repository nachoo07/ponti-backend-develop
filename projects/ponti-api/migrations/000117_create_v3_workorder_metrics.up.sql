-- ========================================
-- MIGRACIÓN 000117: CREATE v3_workorder_metrics (UP)
-- ========================================
-- 
-- Propósito: Vista v3_workorder_metrics usando funciones SSOT consolidadas
-- Dependencias: Requiere v3_core_ssot (000113) y v3_lot_ssot (000115)
-- Cambios: Usa v3_lot_ssot.* (consolidación DRY, elimina duplicados de v3_workorder_ssot)
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- CONSOLIDACIÓN DRY:
-- - Cambiado: v3_workorder_ssot.labor_cost_for_lot_wo → v3_lot_ssot.labor_cost_for_lot
-- - Cambiado: v3_workorder_ssot.supply_cost_for_lot_wo → v3_lot_ssot.supply_cost_for_lot_base
-- - Cambiado: v3_workorder_ssot.surface_for_lot → v3_lot_ssot.surface_for_lot
-- - Cambiado: v3_workorder_ssot.liters_for_lot → v3_lot_ssot.liters_for_lot
-- - Cambiado: v3_workorder_ssot.kilograms_for_lot → v3_lot_ssot.kilograms_for_lot
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- ELIMINAR VISTA ANTIGUA
-- ========================================
DROP VIEW IF EXISTS public.v3_workorder_metrics CASCADE;

-- ========================================
-- CREAR VISTA v3_workorder_metrics
-- ========================================
-- Propósito: Métricas agregadas por lote desde workorders
-- Cálculo ÚNICO que se reutiliza en dashboard

CREATE OR REPLACE VIEW public.v3_workorder_metrics AS
WITH lot_ids AS (
  -- Obtener todos los lotes que tienen workorders
  SELECT DISTINCT
    w.project_id,
    w.field_id,
    w.lot_id
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
)
SELECT
  li.project_id,
  li.field_id,
  li.lot_id,
  
  -- Superficie trabajada (suma de effective_area de workorders)
  v3_lot_ssot.surface_for_lot(li.lot_id) AS surface_ha,
  
  -- Consumos de insumos
  v3_lot_ssot.liters_for_lot(li.lot_id) AS liters,
  v3_lot_ssot.kilograms_for_lot(li.lot_id) AS kilograms,
  
  -- Costos (usa funciones consolidadas de v3_lot_ssot)
  v3_lot_ssot.labor_cost_for_lot(li.lot_id) AS labor_cost_usd,
  v3_lot_ssot.supply_cost_for_lot_base(li.lot_id) AS supplies_cost_usd,
  (v3_lot_ssot.labor_cost_for_lot(li.lot_id) + 
   v3_lot_ssot.supply_cost_for_lot_base(li.lot_id)) AS direct_cost_usd,
  
  -- Costo promedio por hectárea
  v3_core_ssot.cost_per_ha(
    v3_lot_ssot.labor_cost_for_lot(li.lot_id) + 
    v3_lot_ssot.supply_cost_for_lot_base(li.lot_id),
    v3_lot_ssot.surface_for_lot(li.lot_id)
  ) AS avg_cost_per_ha_usd,
  
  -- Consumos por hectárea
  v3_core_ssot.per_ha(
    v3_lot_ssot.liters_for_lot(li.lot_id),
    v3_lot_ssot.surface_for_lot(li.lot_id)
  ) AS liters_per_ha,
  v3_core_ssot.per_ha(
    v3_lot_ssot.kilograms_for_lot(li.lot_id),
    v3_lot_ssot.surface_for_lot(li.lot_id)
  ) AS kilograms_per_ha
  
FROM lot_ids li;

COMMIT;

-- Comentario sobre la vista
COMMENT ON VIEW public.v3_workorder_metrics IS 'Métricas agregadas por lote desde workorders usando funciones SSOT';
