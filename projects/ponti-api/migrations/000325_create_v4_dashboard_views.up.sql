-- ========================================
-- MIGRACIÓN 000325: CREATE v4_report.dashboard_* views (UP)
-- ========================================
--
-- Propósito: Migrar vistas del dashboard a schema v4_report (patrón Strangler Fig)
-- Vistas: dashboard_metrics, dashboard_contributions_progress, dashboard_management_balance,
--         dashboard_crop_incidence, dashboard_operational_indicators
-- Nota: Código en inglés, comentarios en español
-- Fecha: 2025-01-XX

BEGIN;

-- ========================================
-- 1. dashboard_metrics
-- ========================================
CREATE VIEW v4_report.dashboard_metrics AS
WITH lot_data AS (
  SELECT
    lm.project_id,
    lm.lot_id,
    lm.hectares,
    lm.sowed_area_ha,
    lm.harvested_area_ha,
    lm.direct_cost_per_ha_usd
  FROM public.v3_lot_metrics lm
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) AS total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- CARD 1: AVANCE DE SIEMBRA
  COALESCE(SUM(ld.sowed_area_ha), 0)::double precision AS sowing_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS sowing_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS sowing_progress_pct,
  
  -- CARD 2: AVANCE DE COSECHA
  COALESCE(SUM(ld.harvested_area_ha), 0)::double precision AS harvest_hectares,
  COALESCE(SUM(ld.hectares), 0)::double precision AS harvest_total_hectares,
  v3_core_ssot.percentage(
    COALESCE(SUM(ld.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(ld.hectares), 0)::numeric
  ) AS harvest_progress_pct,
  
  -- CARD 3: AVANCE DE COSTOS
  COALESCE(
    SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
    0
  )::double precision AS executed_costs_usd,
  COALESCE(p.planned_cost, 0)::double precision AS budget_cost_usd,
  v3_core_ssot.percentage(
    COALESCE(
      SUM(ld.direct_cost_per_ha_usd * ld.sowed_area_ha) / NULLIF(SUM(ld.sowed_area_ha), 0),
      0
    )::numeric,
    COALESCE(p.planned_cost, 0)::numeric
  ) AS costs_progress_pct,
  
  -- CARD 4: RESULTADO OPERATIVO
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(ld.lot_id)), 0) AS operating_result_income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  (
    COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
    COALESCE(p.admin_cost * ph.total_hectares, 0) +
    COALESCE((
      SELECT f.lease_type_value * ph.total_hectares
      FROM public.fields f
      WHERE f.project_id = p.id AND f.deleted_at IS NULL
      LIMIT 1
    ), 0)
  )::double precision AS operating_result_total_costs_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (
      COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) +
      COALESCE(p.admin_cost * ph.total_hectares, 0) +
      COALESCE((
        SELECT f.lease_type_value * ph.total_hectares
        FROM public.fields f
        WHERE f.project_id = p.id AND f.deleted_at IS NULL
        LIMIT 1
      ), 0)
    )::double precision
  ) AS operating_result_pct,
  
  COALESCE(ph.total_hectares, 0)::numeric AS project_total_hectares
  
FROM public.projects p
LEFT JOIN lot_data ld ON ld.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares;

COMMENT ON VIEW v4_report.dashboard_metrics IS 'Dashboard metrics migrada a v4_report (000325)';

-- ========================================
-- 2. dashboard_management_balance
-- ========================================
CREATE VIEW v4_report.dashboard_management_balance AS
SELECT
  p.id AS project_id,
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(l.id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    v3_dashboard_ssot.total_costs_for_project(p.id)
  ) AS operating_result_pct,
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  (v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
   COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) AS costos_directos_invertidos_usd,
  ((v3_dashboard_ssot.supply_movements_invested_total_for_project(p.id) +
    COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0)) -
   v3_dashboard_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  (v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) -
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0)) AS semillas_stock_usd,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  (v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) -
   COALESCE(SUM(
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
     v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
   ), 0)) AS agroquimicos_stock_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_ejecutados_usd,
  (SELECT COALESCE(SUM(sm.quantity * s.price), 0)
   FROM public.supply_movements sm
   JOIN public.supplies s ON s.id = sm.supply_id
   JOIN public.categories c ON s.category_id = c.id
   WHERE sm.project_id = p.id
     AND sm.deleted_at IS NULL
     AND s.deleted_at IS NULL
     AND sm.is_entry = TRUE
     AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
     AND c.type_id = 3) AS fertilizantes_invertidos_usd,
  ((SELECT COALESCE(SUM(sm.quantity * s.price), 0)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id
    JOIN public.categories c ON s.category_id = c.id
    WHERE sm.project_id = p.id
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.is_entry = TRUE
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND c.type_id = 3) -
   COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0)) AS fertilizantes_stock_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_lot_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0) AS labores_invertidos_usd,
  -- Arriendo: Ejecutados = total (fijo + %), Invertidos = fijo
  v3_dashboard_ssot.lease_invested_for_project(p.id) AS arriendo_ejecutados_usd,
  v3_dashboard_ssot.lease_executed_for_project(p.id) AS arriendo_invertidos_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
  v3_dashboard_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla')), 0) AS semilla_cost,
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos')
  ), 0) AS insumos_cost,
  COALESCE(SUM(v3_lot_ssot.labor_cost_for_lot(l.id)), 0) AS labores_cost,
  COALESCE(SUM(v3_lot_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes')), 0) AS fertilizantes_cost
FROM public.projects p
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id;

COMMENT ON VIEW v4_report.dashboard_management_balance IS 'Balance de gestión migrada a v4_report (000325)';

-- ========================================
-- 3. dashboard_crop_incidence
-- ========================================
CREATE VIEW v4_report.dashboard_crop_incidence AS
SELECT
  p.id AS project_id,
  ci.current_crop_id,
  ci.crop_name,
  ci.crop_hectares,
  ci.crop_incidence_pct,
  v3_dashboard_ssot.cost_per_ha_for_crop_ssot(p.id, ci.current_crop_id)::numeric AS cost_per_ha_usd
FROM public.projects p
CROSS JOIN LATERAL v3_dashboard_ssot.crop_incidence_for_project(p.id) ci
WHERE p.deleted_at IS NULL
ORDER BY p.id, ci.crop_name;

COMMENT ON VIEW v4_report.dashboard_crop_incidence IS 'Incidencia de costos por cultivo migrada a v4_report (000325)';

-- ========================================
-- 4. dashboard_operational_indicators
-- ========================================
CREATE VIEW v4_report.dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  v3_dashboard_ssot.first_workorder_number_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_number_for_project(p.id) AS last_workorder_id,
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW v4_report.dashboard_operational_indicators IS 'Indicadores operativos migrada a v4_report (000325)';

-- ========================================
-- 5. dashboard_contributions_progress
-- ========================================
CREATE VIEW v4_report.dashboard_contributions_progress AS
WITH investor_base AS (
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.deleted_at IS NULL
),
category_totals AS (
  SELECT
    project_id,
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd +
     general_labors_total_usd + sowing_total_usd + irrigation_total_usd +
     rent_capitalizable_total_usd + administration_total_usd) AS total_contributions_usd
  FROM public.v3_report_investor_contribution_categories
),
investor_agrochemicals_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS agrochemicals_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 2
    AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  GROUP BY sm.project_id, sm.investor_id
),
investor_fertilizers_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS fertilizers_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 3
  GROUP BY sm.project_id, sm.investor_id
),
investor_seeds_real AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS seeds_real_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id
  WHERE sm.deleted_at IS NULL
    AND sm.is_entry = TRUE
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND cat.type_id = 1
  GROUP BY sm.project_id, sm.investor_id
),
investor_general_labors_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS general_labors_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name NOT IN ('Siembra', 'Riego', 'Cosecha')
  GROUP BY w.project_id, w.investor_id
),
investor_sowing_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS sowing_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Siembra'
  GROUP BY w.project_id, w.investor_id
),
investor_irrigation_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS irrigation_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Riego'
  GROUP BY w.project_id, w.investor_id
),
investor_rent_real AS (
  SELECT
    f.project_id,
    fi.investor_id,
    SUM(
      v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares *
      COALESCE(fi.percentage, 0) / 100.0
    )::numeric AS rent_real_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN public.field_investors fi ON fi.field_id = f.id AND fi.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, fi.investor_id
),
investor_admin_real AS (
  SELECT
    aci.project_id,
    aci.investor_id,
    (ct.total_contributions_usd * COALESCE(aci.percentage, 0) / 100.0)::numeric AS admin_real_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN public.admin_cost_investors aci
    ON aci.project_id = ib.project_id
    AND aci.investor_id = ib.investor_id
    AND aci.deleted_at IS NULL
),
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    (
      COALESCE(agro.agrochemicals_real_usd, 0) +
      COALESCE(fert.fertilizers_real_usd, 0) +
      COALESCE(seed.seeds_real_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0) +
      COALESCE(ir.rent_real_usd, 0) +
      COALESCE(ia.admin_real_usd, 0)
    )::numeric AS total_real_contribution_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN investor_agrochemicals_real agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_fertilizers_real fert ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_real seed ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors_real glabor ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
  LEFT JOIN investor_sowing_real sow ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation_real irrig ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
  LEFT JOIN investor_rent_real ir ON ir.project_id = ib.project_id AND ir.investor_id = ib.investor_id
  LEFT JOIN investor_admin_real ia ON ia.project_id = ib.project_id AND ia.investor_id = ib.investor_id
)
SELECT
  project_id,
  investor_id,
  investor_name,
  share_pct_agreed AS investor_percentage_pct,
  CASE
    WHEN project_total_contributions_usd > 0
    THEN (total_real_contribution_usd / project_total_contributions_usd * 100)::numeric
    ELSE 0
  END AS contributions_progress_pct
FROM investor_real_contributions;

COMMENT ON VIEW v4_report.dashboard_contributions_progress IS 'Avance de aportes de inversores migrada a v4_report (000325)';

COMMIT;
