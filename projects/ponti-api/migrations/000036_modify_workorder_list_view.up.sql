CREATE OR REPLACE VIEW workorder_list_view AS
SELECT
  w.id,
  w.number,
  w.project_id,           
  w.field_id,    
  p.name              AS project_name,
  f.name              AS field_name,
  l.name              AS lot_name,
  w.date,
  c.name              AS crop_name,
  lb.name             AS labor_name,
  cat_lb.name         AS labor_category_name,
  t.name              AS type_name,
  w.contractor,
  w.effective_area    AS surface_ha,
  s.name              AS supply_name,
  wi.total_used       AS consumption,
  cat.name            AS category_name,
  wi.final_dose       AS dose,
  COALESCE((wi.final_dose * s.price), 0)                     AS cost_per_ha,
  s.price                                       AS unit_price,
  COALESCE(((wi.final_dose * s.price) * w.effective_area), 0) AS total_cost
FROM workorders w
JOIN projects p       ON p.id   = w.project_id
JOIN fields f         ON f.id   = w.field_id
JOIN lots l           ON l.id   = w.lot_id
JOIN crops c          ON c.id   = w.crop_id
JOIN labors lb        ON lb.id  = w.labor_id
JOIN categories cat_lb ON cat_lb.id = lb.category_id
LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
LEFT JOIN supplies s        ON s.id        = wi.supply_id
LEFT JOIN types t           ON t.id        = s.type_id
LEFT JOIN categories cat    ON cat.id      = s.category_id
WHERE w.deleted_at IS NULL;