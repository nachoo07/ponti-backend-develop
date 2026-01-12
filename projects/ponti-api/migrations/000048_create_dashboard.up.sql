DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
WITH
executed_labors_by_project AS (
  SELECT lb.project_id, SUM(lb.price) AS labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorders w
      WHERE w.labor_id = lb.id
        AND w.effective_area > 0
        AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),
used_supplies_by_project AS (
  SELECT sp.project_id, SUM(sp.price) AS supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorder_items wi
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

-- -----------------------------------------------------------------
-- Costos directos por proyecto (SOLO ejecutado / utilizado)
--   B = direct_costs_usd = labors_usd + supplies_usd
-- -----------------------------------------------------------------
v_direct_costs_by_project AS (
  SELECT
    p.id AS project_id,
    COALESCE(el.labors_usd,   0)::numeric(14,2) AS labors_usd,
    COALESCE(us.supplies_usd, 0)::numeric(14,2) AS supplies_usd,
    (COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0))::numeric(14,2) AS direct_costs_usd
  FROM projects p
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project  us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Ingresos por field (tons * 200)  → A = income_usd
-- -----------------------------------------------------------------
v_income_by_field AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(l.tons * 200), 0)::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- -----------------------------------------------------------------
-- Dimensión de niveles (todas las combinaciones)
-- -----------------------------------------------------------------
levels AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Siembra agregada  → card "Avance de siembra"
-- -----------------------------------------------------------------
sowing AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS sowed_area,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Cosecha agregada  → card "Avance de cosecha"
-- -----------------------------------------------------------------
harvest AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS harvested_area,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Costos agregados (admin + directos ejecutados)
--   C = budget_cost_usd (= admin_cost en projects)
--   B = executed_costs_usd (= direct_costs_usd de v_direct_costs_by_project)
-- -----------------------------------------------------------------
costs_agg AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(p.admin_cost,0)),0)::numeric(14,2)        AS budget_cost_usd,        -- C
    COALESCE(SUM(COALESCE(dc.labors_usd,0)),0)::numeric(14,2)       AS executed_labors_usd,
    COALESCE(SUM(COALESCE(dc.supplies_usd,0)),0)::numeric(14,2)     AS executed_supplies_usd,
    COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0)::numeric(14,2) AS executed_costs_usd      -- B
  FROM projects p
  LEFT JOIN v_direct_costs_by_project dc ON dc.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Resultado operativo (RENTABILIDAD) → % rojo = (A-B)/B*100
--   A = income_usd ; B = total_invested_usd (= direct labors + supplies)
-- -----------------------------------------------------------------
operating_result AS (
  WITH income_by_project AS (
    SELECT f.project_id, COALESCE(SUM(vf.income_usd),0)::numeric(14,2) AS income_usd
    FROM v_income_by_field vf
    JOIN fields f ON f.id = vf.field_id AND f.deleted_at IS NULL
    GROUP BY f.project_id
  )
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(ip.income_usd,0)),0)::numeric(14,2)  AS income_usd,          -- A
    COALESCE(SUM(COALESCE(el.labors_usd,0)),0)::numeric(14,2)  AS direct_labors_usd,
    COALESCE(SUM(COALESCE(us.supplies_usd,0)),0)::numeric(14,2) AS direct_supplies_usd,
    (COALESCE(SUM(COALESCE(el.labors_usd,0)),0)
     + COALESCE(SUM(COALESCE(us.supplies_usd,0)),0))::numeric(14,2) AS total_invested_usd -- B
  FROM projects p
  LEFT JOIN income_by_project         ip ON ip.project_id = p.id
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project   us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
)

-- -----------------------------------------------------------------
-- SALIDA ÚNICA (base = levels)
-- -----------------------------------------------------------------
SELECT
  lvl.customer_id,
  lvl.project_id,
  lvl.campaign_id,
  lvl.field_id,

  -- Siembra
  COALESCE(s.sowed_area,0)::numeric(14,2)     AS sowing_hectares,
  COALESCE(s.total_hectares,0)::numeric(14,2) AS sowing_total_hectares,

  -- Cosecha
  COALESCE(h.harvested_area,0)::numeric(14,2) AS harvest_hectares,
  COALESCE(h.total_hectares,0)::numeric(14,2) AS harvest_total_hectares,

  -- Costos (B y C)
  COALESCE(ca.executed_costs_usd,0)::numeric(14,2)     AS executed_costs_usd,     -- B
  COALESCE(ca.executed_labors_usd,0)::numeric(14,2)    AS executed_labors_usd,
  COALESCE(ca.executed_supplies_usd,0)::numeric(14,2)  AS executed_supplies_usd,
  COALESCE(ca.budget_cost_usd,0)::numeric(14,2)        AS budget_cost_usd,        -- C (admin)

  -- Ingresos (A) y Resultado operativo
  COALESCE(o.income_usd,0)::numeric(14,2)             AS income_usd,              -- A
  (COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))::numeric(14,2) AS operating_result_usd,
  CASE WHEN COALESCE(o.total_invested_usd,0) > 0
       THEN ROUND(((COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))
                  / NULLIF(o.total_invested_usd,0) * 100)::numeric, 2)
       ELSE 0 END AS operating_result_pct,                                      -- % rojo = (A-B)/B*100

  -- NÚMERO GRIS 2 listo para UI: (B + C)
  (COALESCE(ca.executed_costs_usd,0) + COALESCE(ca.budget_cost_usd,0))::numeric(14,2)
     AS operating_result_total_costs_usd,                                       -- gris 2 = B + C

  -- Aportes (placeholder hasta que se implemente la tabla de aportes)
  0::numeric(14,2) AS contributions_progress_pct,

  -- Identificador de fila
  'metric'::text AS row_kind

FROM levels lvl
LEFT JOIN sowing s
  ON s.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND s.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND s.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
 AND s.field_id    IS NOT DISTINCT FROM lvl.field_id
LEFT JOIN harvest h
  ON h.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND h.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND h.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
 AND h.field_id    IS NOT DISTINCT FROM lvl.field_id
LEFT JOIN costs_agg ca
  ON ca.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND ca.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND ca.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN operating_result o
  ON o.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND o.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND o.campaign_id IS NOT DISTINCT FROM lvl.campaign_id;