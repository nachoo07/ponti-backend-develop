-- =============================================================================
-- MIGRACIÓN 000309: v4_report.workorder_metrics - Paridad exacta con v3_workorder_metrics
-- =============================================================================
--
-- Propósito: Métricas agregadas por lote desde workorders
-- Fuente: 000117_create_v3_workorder_metrics.up.sql
--
-- Dependencias:
--   - v4_ssot.surface_for_lot, liters_for_lot, kilograms_for_lot
--   - v4_ssot.labor_cost_for_lot, supply_cost_for_lot
--   - v4_core.per_ha
--
-- FASE 1: Paridad exacta con v3_workorder_metrics
--

CREATE OR REPLACE VIEW v4_report.workorder_metrics AS
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
  v4_ssot.surface_for_lot(li.lot_id) AS surface_ha,
  
  -- Consumos de insumos
  v4_ssot.liters_for_lot(li.lot_id) AS liters,
  v4_ssot.kilograms_for_lot(li.lot_id) AS kilograms,
  
  -- Costos (usa funciones consolidadas de v4_ssot)
  v4_ssot.labor_cost_for_lot(li.lot_id) AS labor_cost_usd,
  v4_ssot.supply_cost_for_lot(li.lot_id) AS supplies_cost_usd,
  (v4_ssot.labor_cost_for_lot(li.lot_id) + 
   v4_ssot.supply_cost_for_lot(li.lot_id)) AS direct_cost_usd,
  
  -- Costo promedio por hectárea
  v4_core.per_ha(
    v4_ssot.labor_cost_for_lot(li.lot_id) + 
    v4_ssot.supply_cost_for_lot(li.lot_id),
    v4_ssot.surface_for_lot(li.lot_id)
  ) AS avg_cost_per_ha_usd,
  
  -- Consumos por hectárea
  v4_core.per_ha(
    v4_ssot.liters_for_lot(li.lot_id),
    v4_ssot.surface_for_lot(li.lot_id)
  ) AS liters_per_ha,
  v4_core.per_ha(
    v4_ssot.kilograms_for_lot(li.lot_id),
    v4_ssot.surface_for_lot(li.lot_id)
  ) AS kilograms_per_ha
  
FROM lot_ids li;

COMMENT ON VIEW v4_report.workorder_metrics IS 
'Paridad exacta con v3_workorder_metrics (000117). FASE 1: usa v4_ssot wrappers.';
