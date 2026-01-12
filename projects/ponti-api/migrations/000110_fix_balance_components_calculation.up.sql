-- ========================================
-- MIGRACIÓN 000110: FIX BALANCE COMPONENTS CALCULATION (UP)
-- ========================================
-- 
-- Propósito: Corregir SOLO semilla_cost, insumos_cost y labores_cost
-- Nota: Solo cambia 3 líneas (124, 126, 128) - resto igual que 000104

-- Eliminar vista existente si existe
DROP VIEW IF EXISTS public.v3_dashboard_management_balance CASCADE;

-- Crear vista con estructura corregida
CREATE VIEW public.v3_dashboard_management_balance AS
WITH lots_base AS (
  SELECT
    l.id         AS lot_id,
    f.project_id AS project_id,
    l.hectares   AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM lots_base
  GROUP BY project_id
)
SELECT
  p.id AS project_id,
  
  -- ========================================
  -- INGRESOS Y RESULTADO OPERATIVO
  -- ========================================
  -- Ingresos: suma de ingresos netos de todos los lotes
  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0) AS income_usd,
  -- Resultado operativo: usando función SSOT corregida
  v3_calc.operating_result_total_for_project(p.id) AS operating_result_usd,
  -- Porcentaje de margen operativo
  v3_calc.renta_pct(
    v3_calc.operating_result_total_for_project(p.id),
    (COALESCE(v3_calc.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct,
  
  -- ========================================
  -- COSTOS DIRECTOS (EJECUTADOS + INVERTIDOS + STOCK) - CORREGIDO EN 000110
  -- ========================================
  -- Ejecutados: desde v3_workorder_metrics (función corregida de 000100)
  v3_calc.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  -- Invertidos: supply_movements (Stock + Remito oficial) + labores ejecutadas
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) + COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS costos_directos_invertidos_usd,
  -- Stock: invertidos - ejecutados
  (COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) + COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0)) - 
  v3_calc.direct_costs_total_for_project(p.id) AS costos_directos_stock_usd,
  
  -- ========================================
  -- SEMILLAS (EJECUTADOS + INVERTIDOS + STOCK) - CORREGIDO EN 000110
  -- ========================================
  -- Ejecutados: suma de semillas desde workorder_items (unit_id = 2)
  COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_ejecutados_usd,
  -- Invertidos: suma de supply_movements de semillas (Stock + Remito oficial)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND c.name = 'Semilla'
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) AS semillas_invertidos_usd,
  -- Stock: invertidos - ejecutados
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND c.name = 'Semilla'
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) - COALESCE(SUM(v3_calc.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_stock_usd,
  
  -- ========================================
  -- AGROQUÍMICOS (EJECUTADOS + INVERTIDOS + STOCK) - CORREGIDO EN 000110
  -- ========================================
  -- Ejecutados: suma de agroquímicos desde workorder_items (unit_id = 1)
  COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_ejecutados_usd,
  -- Invertidos: suma de supply_movements de agroquímicos (Stock + Remito oficial)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND c.name != 'Semilla'
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) AS agroquimicos_invertidos_usd,
  -- Stock: invertidos - ejecutados
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND c.name != 'Semilla'
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) - COALESCE(SUM(v3_calc.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_stock_usd,
  
  -- ========================================
  -- LABORES (EJECUTADOS + INVERTIDOS, STOCK = 0)
  -- ========================================
  -- Ejecutados: suma de labores desde workorders
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_ejecutados_usd,
  -- Invertidos: igual a ejecutados (las labores se pagan cuando se ejecutan)
  COALESCE(SUM(v3_calc.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_invertidos_usd,
  
  -- ========================================
  -- ARRIENDO (EJECUTADOS + INVERTIDOS, STOCK = NULL)
  -- ========================================
  -- Ejecutados: 0 (solo aparece al final si es porcentual)
  0::double precision AS arriendo_ejecutados_usd,
  -- Invertidos: lease_type_value * total_hectares
  COALESCE((SELECT f.lease_type_value * ph.total_hectares
            FROM public.fields f
            JOIN project_hectares ph ON ph.project_id = f.project_id
            WHERE f.project_id = p.id AND f.deleted_at IS NULL
            LIMIT 1), 0)::double precision AS arriendo_invertidos_usd,
  
  -- ========================================
  -- ESTRUCTURA (EJECUTADOS + INVERTIDOS, STOCK = NULL)
  -- ========================================
  -- Ejecutados: 0 (solo aparece al final si es porcentual)
  0::double precision AS estructura_ejecutados_usd,
  -- Invertidos: admin_cost * total_hectares
  COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision AS estructura_invertidos_usd,
  
  -- ========================================
  -- COSTOS CALCULADOS (CORREGIDOS EN MIGRACIÓN 000110)
  -- ========================================
  -- semilla_cost: solo categoría "Semilla" (id=1)
  COALESCE((
    SELECT SUM(wi.final_dose * s.price * w.effective_area)
    FROM workorders w
    JOIN workorder_items wi ON wi.workorder_id = w.id
    JOIN supplies s ON s.id = wi.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE w.project_id = p.id 
      AND c.name = 'Semilla'
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND wi.final_dose > 0
      AND s.price IS NOT NULL
  ), 0) AS semilla_cost,
  -- insumos_cost: todas las categorías EXCEPTO "Semilla" (agroquímicos y otros insumos)
  COALESCE((
    SELECT SUM(wi.final_dose * s.price * w.effective_area)
    FROM workorders w
    JOIN workorder_items wi ON wi.workorder_id = w.id
    JOIN supplies s ON s.id = wi.supply_id
    JOIN categories c ON c.id = s.category_id
    WHERE w.project_id = p.id 
      AND c.name != 'Semilla'
      AND w.deleted_at IS NULL
      AND w.effective_area > 0
      AND wi.final_dose > 0
      AND s.price IS NOT NULL
  ), 0) AS insumos_cost,
  -- labores_cost: usar métrica de labores directamente
  COALESCE((
    SELECT SUM(lm.total_labor_cost)
    FROM v3_labor_metrics lm
    WHERE lm.project_id = p.id
  ), 0) AS labores_cost

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY 
  p.id, 
  p.admin_cost, 
  ph.total_hectares;

