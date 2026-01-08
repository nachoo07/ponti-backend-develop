-- Rollback: Restaurar la vista dashboard_view de la migración 000050
DROP VIEW IF EXISTS dashboard_view;

-- Restaurar la vista completa de la migración 000050
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
      JOIN workorders w ON w.id = wi.workorder_id
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
        AND w.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

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

v_income_by_field AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(l.tons), 0)::numeric(14,2) AS total_tons,
    0::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

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

costs_agg AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(p.admin_cost,0)),0)::numeric(14,2)        AS budget_cost_usd,
    COALESCE(SUM(COALESCE(dc.labors_usd,0)),0)::numeric(14,2)       AS executed_labors_usd,
    COALESCE(SUM(COALESCE(dc.supplies_usd,0)),0)::numeric(14,2)     AS executed_supplies_usd,
    COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0)::numeric(14,2) AS executed_costs_usd,
    COALESCE(SUM(20000),0)::numeric(14,2)                            AS budget_total_usd,
    CASE WHEN COALESCE(SUM(20000),0) > 0
         THEN ROUND(((COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0) / NULLIF(SUM(20000),0)) * 100)::numeric, 2)
         ELSE 0 END AS costs_progress_pct
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
    COALESCE(SUM(COALESCE(ip.income_usd,0)),0)::numeric(14,2)  AS income_usd,
    COALESCE(SUM(COALESCE(el.labors_usd,0)),0)::numeric(14,2)  AS direct_labors_usd,
    COALESCE(SUM(COALESCE(us.supplies_usd,0)),0)::numeric(14,2) AS direct_supplies_usd,
    (COALESCE(SUM(COALESCE(el.labors_usd,0)),0)
     + COALESCE(SUM(COALESCE(us.supplies_usd,0)),0))::numeric(14,2) AS total_invested_usd
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
),

contributions AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    100.00::numeric(6,2) AS contributions_progress_pct
  FROM projects p
  LEFT JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
  LEFT JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, pi.investor_id, i.name, pi.percentage),
    (p.customer_id, p.id, pi.investor_id, i.name, pi.percentage),
    (p.customer_id, pi.investor_id, i.name, pi.percentage),
    (pi.investor_id, i.name, pi.percentage)
  )
)

SELECT
  lvl.customer_id,
  lvl.project_id,
  lvl.campaign_id,
  lvl.field_id,

  COALESCE(s.sowed_area,0)::numeric(14,2)     AS sowing_hectares,
  COALESCE(s.total_hectares,0)::numeric(14,2) AS sowing_total_hectares,
  CASE WHEN COALESCE(s.total_hectares,0) > 0 
       THEN ROUND((COALESCE(s.sowed_area,0) / NULLIF(s.total_hectares,0) * 100)::numeric, 2)
       ELSE 0 END AS sowing_progress_percent,

  COALESCE(h.harvested_area,0)::numeric(14,2) AS harvest_hectares,
  COALESCE(h.total_hectares,0)::numeric(14,2) AS harvest_total_hectares,
  CASE WHEN COALESCE(h.total_hectares,0) > 0 
       THEN ROUND((COALESCE(h.harvested_area,0) / NULLIF(h.total_hectares,0) * 100)::numeric, 2)
       ELSE 0 END AS harvest_progress_percent,

  COALESCE(ca.executed_costs_usd,0)::numeric(14,2)     AS executed_costs_usd,
  COALESCE(ca.executed_labors_usd,0)::numeric(14,2)    AS executed_labors_usd,
  COALESCE(ca.executed_supplies_usd,0)::numeric(14,2)  AS executed_supplies_usd,
  COALESCE(ca.budget_cost_usd,0)::numeric(14,2)        AS budget_cost_usd,
  COALESCE(ca.budget_total_usd,0)::numeric(14,2)       AS budget_total_usd,
  COALESCE(ca.costs_progress_pct,0)::numeric(14,2)     AS costs_progress_pct,

  COALESCE(o.income_usd,0)::numeric(14,2)             AS income_usd,
  (COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))::numeric(14,2) AS operating_result_usd,
  CASE WHEN COALESCE(o.total_invested_usd,0) > 0
       THEN ROUND(((COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))
                  / NULLIF(o.total_invested_usd,0) * 100)::numeric, 2)
       ELSE 0 END AS operating_result_pct,

  (COALESCE(ca.executed_costs_usd,0) + COALESCE(ca.budget_cost_usd,0))::numeric(14,2)
     AS operating_result_total_costs_usd,

  COALESCE(c.investor_id,0)::bigint AS investor_id,
  COALESCE(c.investor_name,'') AS investor_name,
  COALESCE(c.investor_percentage_pct,0)::numeric(6,2) AS investor_percentage_pct,
  COALESCE(c.contributions_progress_pct,0)::numeric(6,2) AS contributions_progress_pct,

  0::bigint AS crop_id,
  '' AS crop_name,
  0::numeric(14,2) AS crop_hectares,
  0::numeric(14,2) AS project_total_hectares,
  0::numeric(6,2) AS incidence_pct,
  0::numeric(14,2) AS crop_direct_costs_usd,
  0::numeric(14,2) AS cost_per_ha_usd,

  0::numeric(14,2) AS semilla_ejecutados_usd,
  0::numeric(14,2) AS semilla_invertidos_usd,
  0::numeric(14,2) AS semilla_stock_usd,
  0::numeric(14,2) AS insumos_ejecutados_usd,
  0::numeric(14,2) AS insumos_invertidos_usd,
  0::numeric(14,2) AS insumos_stock_usd,
  0::numeric(14,2) AS labores_ejecutados_usd,
  0::numeric(14,2) AS labores_invertidos_usd,
  0::numeric(14,2) AS labores_stock_usd,
  0::numeric(14,2) AS arriendo_ejecutados_usd,
  0::numeric(14,2) AS arriendo_invertidos_usd,
  0::numeric(14,2) AS arriendo_stock_usd,
  0::numeric(14,2) AS estructura_ejecutados_usd,
  0::numeric(14,2) AS estructura_invertidos_usd,
  0::numeric(14,2) AS estructura_stock_usd,
  0::numeric(14,2) AS costos_directos_ejecutados_usd,
  0::numeric(14,2) AS costos_directos_invertidos_usd,
  0::numeric(14,2) AS costos_directos_stock_usd,

  NULL::timestamp AS primera_orden_fecha,
  0::bigint AS primera_orden_id,
  NULL::timestamp AS ultima_orden_fecha,
  0::bigint AS ultima_orden_id,
  NULL::timestamp AS arqueo_stock_fecha,
  NULL::timestamp AS cierre_campana_fecha,

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
  ON o.customer_id IS NOT DISTINCT FROM lvl.campaign_id
 AND o.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND o.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN contributions c
  ON c.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND c.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND c.campaign_id IS NOT DISTINCT FROM lvl.campaign_id;
