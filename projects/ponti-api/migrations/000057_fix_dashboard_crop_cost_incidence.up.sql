-- ========================================
-- MIGRACIÓN 000056: FIX INCIDENCIA DE COSTOS POR CULTIVO
-- ========================================
-- 
-- Objetivo: Crear vista para la incidencia de costos por cultivo
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Crear vista para la incidencia de costos por cultivo
DROP VIEW IF EXISTS dashboard_crop_cost_incidence_view;
CREATE VIEW dashboard_crop_cost_incidence_view AS
WITH lot_costs AS (
  SELECT w.lot_id,
         SUM(lb.price*w.effective_area) + SUM(wi.total_used*s.price) AS direct_costs_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
  GROUP BY w.lot_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  l.current_crop_id AS crop_id,
  c.name AS crop_name,
  SUM(l.hectares) AS crop_hectares,
  SUM(SUM(l.hectares)) OVER (PARTITION BY p.id) AS project_total_hectares,
  (SUM(l.hectares)::numeric / NULLIF(SUM(SUM(l.hectares)) OVER (PARTITION BY p.id),0) * 100) AS incidence_pct,
  COALESCE(SUM(lc.direct_costs_usd),0) AS crop_direct_costs_usd,
  CASE WHEN SUM(l.hectares)>0
       THEN COALESCE(SUM(lc.direct_costs_usd),0)::numeric / SUM(l.hectares)
       ELSE 0 END AS cost_per_ha_usd
FROM projects p
JOIN fields f ON f.project_id=p.id
JOIN lots l ON l.field_id=f.id
JOIN crops c ON c.id=l.current_crop_id
LEFT JOIN lot_costs lc ON lc.lot_id=l.id
GROUP BY p.customer_id,p.id,p.campaign_id,l.current_crop_id,c.name;
