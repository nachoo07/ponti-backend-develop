-- ========================================
-- MIGRATION 000137: FIX SUMMARY RESULTS DIRECT COSTS CALCULATION (DOWN)
-- ========================================
-- 
-- Purpose: Revertir cambios de cálculo de costos directos
-- Date: 2025-10-12
-- Author: System
-- 
-- Revierte a usar v3_calc.direct_cost_for_lot (con movimientos internos)
--
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_summary_results_view: resumen de resultados generales por cultivo (REVERTIDO)
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_summary_results_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    -- Superficie sembrada: suma de effective_area de workorders de tipo siembra
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
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
),
by_crop AS (
  SELECT
    lb.project_id,
    lb.current_crop_id,
    lb.crop_name::text AS crop_name,
    
    -- Superficie total por cultivo
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS surface_ha,
    
    -- CORREGIDO: Usar función SSOT v3_calc.income_net_total_for_lot
    -- Esta función calcula: tons × net_price (correcto)
    -- Antes: superficie × precio (INCORRECTO)
    COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric AS net_income_usd,
    
    -- Costos directos totales por cultivo (labores + insumos)
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric AS direct_costs_usd,
    
    -- Arriendo total por cultivo: superficie × costo de arriendo por ha (usando función SSOT)
    COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric AS rent_usd,
    
    -- Estructura total por cultivo: costo fijo × superficie (usando función SSOT)
    COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric AS structure_usd,
    
    -- Total invertido por cultivo: Costos directos + Arriendo + Estructura
    (
      COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric
      + COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric
      + COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric
    ) AS total_invested_usd,
    
    -- CORREGIDO: Resultado operativo total por cultivo: Ingreso Neto - Total Invertido
    -- Usa la función SSOT v3_calc.income_net_total_for_lot (correcto)
    (
      COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)::numeric
      - (
        COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric
        + COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric
        + COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric
      )
    ) AS operating_result_usd
    
  FROM lot_base lb
  WHERE lb.current_crop_id IS NOT NULL
  GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_usd,
    SUM(operating_result_usd)::numeric AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  bc.total_invested_usd,
  bc.operating_result_usd,
  
  -- Renta del cultivo (%)
  v3_calc.renta_pct(
    bc.operating_result_usd::double precision,
    bc.total_invested_usd::double precision
  )::numeric AS crop_return_pct,
  
  -- Totales del proyecto (para comparación)
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_usd AS total_invested_project_usd,
  pt.total_operating_result_usd,
  
  -- Renta total del proyecto (%)
  v3_calc.renta_pct(
    pt.total_operating_result_usd::double precision,
    pt.total_invested_usd::double precision
  )::numeric AS project_return_pct

FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id
ORDER BY bc.project_id, bc.current_crop_id;

