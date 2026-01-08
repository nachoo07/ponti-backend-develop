-- ========================================
-- MIGRACIÓN 000059: FIX DASHBOARD CROP INCIDENCE VIEW
-- ========================================
-- 
-- Objetivo: Crear vista dashboard_crop_incidence_view con columnas correctas
-- Fecha: 2025-01-27
-- Autor: Sistema

-- Crear vista para la incidencia de cultivos con columnas correctas
DROP VIEW IF EXISTS dashboard_crop_incidence_view;
CREATE VIEW dashboard_crop_incidence_view AS
WITH 
-- Costos por cultivo y proyecto
crop_costs AS (
  SELECT 
    p.id AS project_id,
    c.id AS crop_id,
    c.name AS crop_name,
    COALESCE(SUM(l.hectares), 0) AS crop_hectares,
    COALESCE(SUM(l.hectares), 0) AS project_total_hectares,
    -- Calcular costos directos por cultivo
    COALESCE(SUM(
      CASE WHEN w.effective_area > 0 THEN lb.price * w.effective_area ELSE 0 END
    ), 0) AS crop_direct_costs_usd
  FROM projects p
  JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  JOIN crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  LEFT JOIN workorders w ON w.field_id = f.id AND w.deleted_at IS NULL
  LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, c.id, c.name
),
-- Calcular incidencia y costos por hectárea
crop_incidence AS (
  SELECT 
    cc.project_id,
    cc.crop_id,
    cc.crop_name,
    cc.crop_hectares,
    cc.project_total_hectares,
    -- Porcentaje de incidencia (hectáreas del cultivo / total del proyecto)
    CASE 
      WHEN cc.project_total_hectares > 0 
      THEN (cc.crop_hectares / cc.project_total_hectares) * 100
      ELSE 0 
    END AS incidence_pct,
    -- Costo por hectárea
    CASE 
      WHEN cc.crop_hectares > 0 
      THEN cc.crop_direct_costs_usd / cc.crop_hectares
      ELSE 0 
    END AS cost_per_ha_usd
  FROM crop_costs cc
)
SELECT
  ci.project_id,
  ci.crop_name,
  ci.crop_hectares,
  ci.incidence_pct,
  ci.cost_per_ha_usd
FROM crop_incidence ci
WHERE ci.crop_hectares > 0
ORDER BY ci.crop_name;
