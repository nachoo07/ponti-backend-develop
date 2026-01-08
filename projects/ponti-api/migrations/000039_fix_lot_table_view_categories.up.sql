-- =======================
-- CORRECCIÓN COMPLETA DE VISTA LOT_TABLE_VIEW
-- =======================
-- Esta migración corrige la vista lot_table_view creada en la migración 000037
-- Incluye:
-- 1. Cambio de categorías incorrectas por las reales (9 y 13)
-- 2. Lógica exacta de cálculos de la 00037
-- 3. Optimización de índices

-- Primero eliminar la vista existente
DROP VIEW IF EXISTS lot_table_view;

-- Luego crear la vista completa con todas las correcciones
CREATE VIEW lot_table_view AS
WITH
-- =======================
-- CÁLCULO DE ÁREA SEMBRADA (optimizado)
-- =======================
sowing AS (
  SELECT 
    w.lot_id, 
    SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 9  -- ID correcto de "Siembra"
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE ÁREA COSECHADA (optimizado)
-- =======================
harvest AS (
  SELECT 
    w.lot_id,
    SUM(w.effective_area) AS harvested_area,
    MAX(w.date) AS last_harvest_date
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id = 13  -- ID correcto de "Cosecha"
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE COSTOS DIRECTOS (optimizado)
-- =======================
direct_costs AS (
  SELECT 
    w.lot_id,
    SUM(COALESCE(lb.price * w.effective_area, 0)) AS labor_cost,
    SUM(COALESCE(wi.final_dose * s.price * w.effective_area, 0)) AS supply_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE INGRESO NETO (optimizado) - PRECIO REAL DE COMERCIALIZACIÓN
-- =======================
income_net AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id 
    AND cc.crop_id = l.current_crop_id 
    AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.tons IS NOT NULL
    AND l.tons > 0
),

-- =======================
-- CÁLCULO DE ARRIENDO (optimizado) - LÓGICA COMPLEJA DE TIPOS DE ARRIENDO
-- =======================
rent_calculation AS (
  SELECT 
    l.id AS lot_id,
    CASE 
      WHEN f.lease_type_id = 1 THEN -- Fixed amount
        COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)
      WHEN f.lease_type_id = 2 THEN -- Percentage
        (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0)
      WHEN f.lease_type_id = 3 THEN -- Both
        (COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)) + 
        ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0))
      ELSE 0
    END AS rent_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN harvest h ON h.lot_id = l.id
  LEFT JOIN income_net in_net ON in_net.lot_id = l.id
  WHERE l.deleted_at IS NULL
),

-- =======================
-- CÁLCULO DE COSTO ADMINISTRATIVO (optimizado)
-- =======================
admin_cost AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(p.admin_cost, 0) * COALESCE(l.hectares, 0) AS admin_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.hectares IS NOT NULL
    AND l.hectares > 0
),

-- =======================
-- CÁLCULO DE FECHAS (optimizado)
-- =======================
lot_dates AS (
  SELECT 
    w.lot_id,
    MIN(CASE WHEN lb.category_id = 9 THEN w.date END) AS lot_sowing_date,
    MAX(CASE WHEN lb.category_id = 13 THEN w.date END) AS lot_harvest_date,
    COUNT(DISTINCT w.id) AS sequence
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND lb.category_id IN (9, 13) -- IDs correctos de Siembra y Cosecha
    AND w.date IS NOT NULL
  GROUP BY w.lot_id
)

-- =======================
-- SELECT PRINCIPAL OPTIMIZADO
-- =======================
SELECT 
  l.id,
  f.project_id,
  l.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.name AS lot_name,
  pc.name AS previous_crop,
  l.previous_crop_id,
  cc.name AS current_crop,
  l.current_crop_id,
  l.variety,
  l.hectares,
  COALESCE(s.sowed_area, 0) AS sowed_area,
  l.season,
  COALESCE(l.tons, 0) AS tons,
  COALESCE(h.harvested_area, 0) AS harvested_area,
  h.last_harvest_date AS harvest_date,
  
  -- =======================
  -- COSTOS DIRECTOS
  -- =======================
  COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0) AS direct_cost_total,
  
  -- Costo por hectárea (USD/HA)
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
    ELSE 0 
  END AS cost_usd_per_ha,
  
  -- =======================
  -- INGRESOS
  -- =======================
  COALESCE(in_net.income_net_total, 0) AS income_net_total,
  
  -- Ingreso neto por hectárea
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(in_net.income_net_total, 0) / s.sowed_area
    ELSE 0 
  END AS income_net_per_ha,
  
  -- =======================
  -- RENDIMIENTO
  -- =======================
  -- Rendimiento (toneladas por hectárea cosechada)
  CASE 
    WHEN COALESCE(h.harvested_area, 0) > 0 
    THEN COALESCE(l.tons, 0) / h.harvested_area
    ELSE 0 
  END AS yield_tn_per_ha,
  
  -- =======================
  -- ARRIENDO
  -- =======================
  COALESCE(rc.rent_total, 0) AS rent_total,
  
  -- Arriendo por hectárea
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(rc.rent_total, 0) / s.sowed_area
    ELSE 0 
  END AS rent_per_ha,
  
  -- =======================
  -- COSTOS ADMINISTRATIVOS
  -- =======================
  COALESCE(ac.admin_total, 0) AS admin_total,
  
  -- Costo administrativo por hectárea
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN COALESCE(ac.admin_total, 0) / s.sowed_area
    ELSE 0 
  END AS admin_cost_per_ha,
  
  -- =======================
  -- TOTALES ACTIVOS
  -- =======================
  -- Total activo (costos + arriendo + admin)
  (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
  COALESCE(rc.rent_total, 0) + 
  COALESCE(ac.admin_total, 0) AS active_total,
  
  -- Total activo por hectárea
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
          COALESCE(rc.rent_total, 0) + 
          COALESCE(ac.admin_total, 0)) / s.sowed_area
    ELSE 0 
  END AS active_total_per_ha,
  
  -- =======================
  -- RESULTADO OPERATIVO
  -- =======================
  -- Resultado operativo total
  COALESCE(in_net.income_net_total, 0) - 
  ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
   COALESCE(rc.rent_total, 0) + 
   COALESCE(ac.admin_total, 0)) AS operating_result,
  
  -- Resultado operativo por hectárea
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(in_net.income_net_total, 0) - 
          ((COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) + 
           COALESCE(rc.rent_total, 0) + 
           COALESCE(ac.admin_total, 0))) / s.sowed_area
    ELSE 0 
  END AS operating_result_per_ha,
  
  -- =======================
  -- FECHAS Y SECUENCIA
  -- =======================
  ld.lot_sowing_date,
  ld.lot_harvest_date,
  COALESCE(ld.sequence, 0) AS sequence,
  l.updated_at

FROM lots l
JOIN fields f ON l.field_id = f.id AND f.deleted_at IS NULL
JOIN projects p ON f.project_id = p.id AND p.deleted_at IS NULL
LEFT JOIN crops pc ON l.previous_crop_id = pc.id AND pc.deleted_at IS NULL
LEFT JOIN crops cc ON l.current_crop_id = cc.id AND cc.deleted_at IS NULL
LEFT JOIN sowing s ON l.id = s.lot_id
LEFT JOIN harvest h ON l.id = h.lot_id
LEFT JOIN direct_costs dc ON l.id = dc.lot_id
LEFT JOIN income_net in_net ON l.id = in_net.lot_id
LEFT JOIN rent_calculation rc ON l.id = rc.lot_id
LEFT JOIN admin_cost ac ON l.id = ac.lot_id
LEFT JOIN lot_dates ld ON l.id = ld.lot_id
WHERE l.deleted_at IS NULL
  AND l.hectares IS NOT NULL
  AND l.hectares > 0;

-- =======================
-- ÍNDICES OPTIMIZADOS PARA CLOUD SQL (GCP)
-- =======================

-- Índices parciales para soft-delete (estándar en GCP)
CREATE INDEX IF NOT EXISTS idx_lot_table_workorders_notdel 
  ON workorders(lot_id, effective_area, date) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_labors_notdel 
  ON labors(id, category_id, price) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_workorder_items_notdel 
  ON workorder_items(workorder_id, supply_id, final_dose) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_supplies_notdel 
  ON supplies(id, price) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_lots_notdel 
  ON lots(id, field_id, current_crop_id, previous_crop_id, tons, hectares) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_fields_notdel 
  ON fields(id, project_id, lease_type_id, lease_type_value, lease_type_percent) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_projects_notdel 
  ON projects(id, admin_cost) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_crops_notdel 
  ON crops(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lot_table_crop_commercializations_notdel 
  ON crop_commercializations(project_id, crop_id, net_price) 
  WHERE deleted_at IS NULL;

-- Índices compuestos para JOINs frecuentes
CREATE INDEX IF NOT EXISTS idx_lot_table_workorders_composite 
  ON workorders(lot_id, labor_id, effective_area, date) 
  WHERE deleted_at IS NULL AND effective_area > 0;

CREATE INDEX IF NOT EXISTS idx_lot_table_lots_composite 
  ON lots(field_id, current_crop_id, previous_crop_id, tons, hectares) 
  WHERE deleted_at IS NULL AND hectares > 0;

-- Índices para categorías específicas (corregidos a 9 y 13)
CREATE INDEX IF NOT EXISTS idx_lot_table_labors_sowing 
  ON labors(id, category_id) 
  WHERE deleted_at IS NULL AND category_id = 9;

CREATE INDEX IF NOT EXISTS idx_lot_table_labors_harvest 
  ON labors(id, category_id) 
  WHERE deleted_at IS NULL AND category_id = 13;
