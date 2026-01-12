-- =========================================================
-- CORRECCIÓN: workorder_metrics_view con lógica correcta
--  - Litros: final_dose para Herbicidas(4), Fungicidas(6), Insecticidas(5), Coadyuvantes(2)
--  - Kilogramos: final_dose para Fertilizantes(8), Semilla(1), Curasemillas(3)
--  - Costo directo: labor_price + supply_price (sin multiplicar por área)
--  - Filtros: customer_id y campaign_id agregados
-- =========================================================
DROP VIEW IF EXISTS workorder_metrics_view;

CREATE VIEW workorder_metrics_view AS
SELECT
  w.project_id,
  w.field_id,
  p.customer_id,
  p.campaign_id,
  SUM(w.effective_area) AS surface_ha,
  -- litros: final_dose para Herbicidas(4), Fungicidas(6), Insecticidas(5), Coadyuvantes(2)
  SUM(COALESCE(wi.final_dose,0))
    FILTER (WHERE s.category_id IN (2,4,5,6)) AS liters,
  -- kilogramos: final_dose para Fertilizantes(8), Semilla(1), Curasemillas(3)
  SUM(COALESCE(wi.final_dose,0))
    FILTER (WHERE s.category_id IN (1,3,8)) AS kilograms,
  -- costo directo: suma de precios únicos de labor + suma de precios únicos de insumos
  (SELECT SUM(price) FROM labors WHERE id IN (SELECT DISTINCT labor_id FROM workorders WHERE project_id = w.project_id AND field_id = w.field_id AND deleted_at IS NULL)) +
  (SELECT SUM(price) FROM supplies WHERE id IN (SELECT DISTINCT supply_id FROM workorder_items wi2 JOIN workorders w2 ON wi2.workorder_id = w2.id WHERE w2.project_id = w.project_id AND w2.field_id = w.field_id AND w2.deleted_at IS NULL AND wi2.deleted_at IS NULL)) AS direct_cost
FROM workorders w
  JOIN projects p ON p.id = w.project_id
  LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
  LEFT JOIN supplies s ON s.id = wi.supply_id
WHERE w.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND (wi.deleted_at IS NULL OR wi.deleted_at IS NULL)
  AND (s.deleted_at IS NULL OR s.deleted_at IS NULL)
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
GROUP BY w.project_id, w.field_id, p.customer_id, p.campaign_id;
