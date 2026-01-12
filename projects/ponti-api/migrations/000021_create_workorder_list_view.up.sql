/* CREATE OR REPLACE VIEW workorder_list_view AS
WITH
-- =======================
-- BASE DE ÓRDENES DE TRABAJO (información principal)
-- =======================
workorder_base AS (
  SELECT
    w.id,
    w.number,
    w.project_id,           
    w.field_id,
    w.lot_id,
    w.labor_id,
    w.crop_id,
    w.effective_area,
    w.date,
    w.contractor
  FROM workorders w
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),

-- =======================
-- INFORMACIÓN DE PROYECTO, CAMPO Y LOTE
-- =======================
location_info AS (
  SELECT
    wb.id,
    wb.number,
    wb.project_id,
    wb.field_id,
    wb.lot_id,
    wb.labor_id,
    wb.crop_id,
    wb.effective_area,
    wb.date,
    wb.contractor,
    -- Información de proyecto
    p.name AS project_name,
    -- Información de campo
    f.name AS field_name,
    -- Información de lote
    l.name AS lot_name,
    -- Información de cultivo
    c.name AS crop_name,
    -- Información de labor
    lb.name AS labor_name
  FROM workorder_base wb
  JOIN projects p ON p.id = wb.project_id AND p.deleted_at IS NULL
  JOIN fields f ON f.id = wb.field_id AND f.deleted_at IS NULL
  JOIN lots l ON l.id = wb.lot_id AND l.deleted_at IS NULL
  JOIN crops c ON c.id = wb.crop_id AND c.deleted_at IS NULL
  JOIN labors lb ON lb.id = wb.labor_id AND lb.deleted_at IS NULL
),

-- =======================
-- INSUMOS Y COSTOS (agregados por workorder para evitar duplicados)
-- =======================
supply_details AS (
  SELECT
    w.id AS workorder_id,
    -- Información de insumo
    s.name AS supply_name,
    -- Consumo y dosis
    wi.total_used AS consumption,
    wi.final_dose AS dose,
    -- Información de tipo y categoría
    t.name AS type_name,
    cat.name AS category_name,
    -- Costos
    s.price AS unit_price,
    (wi.final_dose * s.price) AS cost_per_ha,
    ((wi.final_dose * s.price) * w.effective_area) AS total_cost
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  JOIN types t ON t.id = s.type_id AND t.deleted_at IS NULL
  JOIN categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
)

SELECT
  li.id,
  li.number,
  li.project_id,           
  li.field_id,
  li.lot_id,
  li.labor_id,
  li.effective_area,
  li.project_name,
  li.field_name,
  li.lot_name,
  li.date,
  li.crop_name,
  li.labor_name,
  -- COMPATIBILIDAD: Mantener nombres originales para el código Go
  COALESCE(sd.type_name, '') AS type_name,           -- ← NOMBRE ORIGINAL
  li.contractor,
  li.effective_area AS surface_ha,                    -- ← NOMBRE ORIGINAL
  COALESCE(sd.supply_name, '') AS supply_name,       -- ← NOMBRE ORIGINAL
  COALESCE(sd.consumption, 0) AS consumption,        -- ← NOMBRE ORIGINAL
  COALESCE(sd.category_name, '') AS category_name,   -- ← NOMBRE ORIGINAL
  COALESCE(sd.dose, 0) AS dose,                      -- ← NOMBRE ORIGINAL
  COALESCE(sd.cost_per_ha, 0) AS cost_per_ha,        -- ← NOMBRE ORIGINAL
  COALESCE(sd.unit_price, 0) AS unit_price,          -- ← NOMBRE ORIGINAL
  COALESCE(sd.total_cost, 0) AS total_cost           -- ← NOMBRE ORIGINAL
FROM location_info li
LEFT JOIN supply_details sd ON sd.workorder_id = li.id
WHERE li.effective_area > 0; -- Solo workorders con superficie válida

-- =======================
-- ÍNDICES OPTIMIZADOS PARA CLOUD SQL (GCP)
-- =======================

-- Índices parciales para soft-delete (estándar en GCP)
CREATE INDEX IF NOT EXISTS idx_workorder_list_workorders_notdel 
  ON workorders(id, number, project_id, field_id, lot_id, labor_id, crop_id, effective_area, date) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_projects_notdel 
  ON projects(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_fields_notdel 
  ON fields(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_lots_notdel 
  ON lots(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_crops_notdel 
  ON crops(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_labors_notdel 
  ON labors(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_workorder_items_notdel 
  ON workorder_items(workorder_id, supply_id, total_used, final_dose) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_supplies_notdel 
  ON supplies(id, name, price, type_id, category_id) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_types_notdel 
  ON types(id, name) 
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_list_categories_notdel 
  ON categories(id, name) 
  WHERE deleted_at IS NULL;

-- Índice compuesto para JOINs frecuentes
CREATE INDEX IF NOT EXISTS idx_workorder_list_composite 
  ON workorders(project_id, field_id, lot_id, labor_id, crop_id, effective_area) 
  WHERE deleted_at IS NULL AND effective_area > 0;

-- Índice para búsquedas por número
CREATE INDEX IF NOT EXISTS idx_workorder_list_number 
  ON workorders(number) 
  WHERE deleted_at IS NULL;

-- Índice para búsquedas por fecha
CREATE INDEX IF NOT EXISTS idx_workorder_list_date 
  ON workorders(date) 
  WHERE deleted_at IS NULL; */



-----------------------------------------------
-- reemplazar lo de abajo por lo de arriba
-----------------------------------------------

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
  t.name              AS type_name,
  w.contractor,
  w.effective_area    AS surface_ha,
  s.name              AS supply_name,
  wi.total_used       AS consumption,
  cat.name            AS category_name,
  wi.final_dose       AS dose,
  (wi.final_dose * s.price)                     AS cost_per_ha,
  s.price                                       AS unit_price,
  ((wi.final_dose * s.price) * w.effective_area) AS total_cost
FROM workorders w
JOIN projects p       ON p.id   = w.project_id
JOIN fields f         ON f.id   = w.field_id
JOIN lots l           ON l.id   = w.lot_id
JOIN crops c          ON c.id   = w.crop_id
JOIN labors lb        ON lb.id  = w.labor_id
JOIN workorder_items wi ON wi.workorder_id = w.id
JOIN supplies s        ON s.id        = wi.supply_id
JOIN types t           ON t.id        = s.type_id
JOIN categories cat    ON cat.id      = s.category_id
WHERE w.deleted_at IS NULL;