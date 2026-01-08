-- Migración 000054: Revertir inclusión de costo administrativo en resultado operativo
-- Comentarios en español según reglas del proyecto

-- Recrear la vista anterior sin costo administrativo
DROP VIEW IF EXISTS dashboard_operating_result_view;

CREATE OR REPLACE VIEW dashboard_operating_result_view AS
WITH labors_costs AS (
  SELECT
    w.project_id,
    SUM(lb.price * w.effective_area) AS executed_labors_usd
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.effective_area > 0
  GROUP BY w.project_id
),
supplies_costs AS (
  SELECT
    w.project_id,
    SUM(wi.total_used * s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id
  JOIN supplies s ON s.id = wi.supply_id
  WHERE wi.final_dose > 0
  GROUP BY w.project_id
),
harvest AS (
  SELECT f.project_id, SUM(l.tons) AS total_tons
  FROM fields f
  JOIN lots l ON l.field_id=f.id
  GROUP BY f.project_id
),
-- Calcular ingresos por campo según tipo de arriendo
field_income AS (
  SELECT 
    f.id AS field_id,
    f.project_id,
    f.lease_type_id,
    f.lease_type_percent,
    f.lease_type_value,
    -- Hectáreas del campo
    COALESCE(SUM(l.hectares), 0) AS field_hectares,
    -- Comercializaciones del campo (toneladas cosechadas × precio placeholder)
    COALESCE(SUM(l.tons), 0) AS field_tons,
    -- Precio placeholder por tonelada ($400/ton)
    400.00 AS price_per_ton
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id
  GROUP BY f.id, f.project_id, f.lease_type_id, f.lease_type_percent, f.lease_type_value
),
-- Calcular ingresos por campo individual
field_income_calculation AS (
  SELECT 
    field_id,
    project_id,
    CASE lease_type_id
      WHEN 1 THEN -- % INGRESO NETO
        (lease_type_percent / 100.0) * (field_tons * price_per_ton)
      WHEN 2 THEN -- % DE UTILIDAD (HARDCODEAR)
        5000.00  -- Valor fijo hardcodeado
      WHEN 3 THEN -- ARRIENDO FIJO
        lease_type_value * field_hectares
      WHEN 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
        (lease_type_value * field_hectares) + 
        ((lease_type_percent / 100.0) * (field_tons * price_per_ton))
      ELSE 0
    END AS field_income_usd
  FROM field_income
),
-- Sumar ingresos de todos los campos por proyecto
project_income AS (
  SELECT 
    project_id,
    SUM(field_income_usd) AS total_income_usd
  FROM field_income_calculation
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(h.total_tons,0) AS total_tons,
  COALESCE(pi.total_income_usd, 0) AS income_usd,  -- Ingresos calculados por tipo de arriendo
  COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0) AS operating_result_total_costs_usd,
  -- RESULTADO OPERATIVO SIN COSTO ADMINISTRATIVO: Ingresos - Costos Directos
  (COALESCE(pi.total_income_usd, 0)) - (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) AS operating_result_usd,
  -- PORCENTAJE SIN COSTO ADMINISTRATIVO: ((Ingresos - Costos Directos) / Costos Directos) × 100
  CASE WHEN (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) > 0
       THEN ((COALESCE(pi.total_income_usd, 0)) - (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0))) / (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) * 100
       ELSE 0 END AS operating_result_pct,
  COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0) AS executed_costs_usd
FROM projects p
LEFT JOIN labors_costs lc ON lc.project_id = p.id
LEFT JOIN supplies_costs sc ON sc.project_id = p.id
LEFT JOIN harvest h ON h.project_id = p.id
LEFT JOIN project_income pi ON pi.project_id = p.id;
