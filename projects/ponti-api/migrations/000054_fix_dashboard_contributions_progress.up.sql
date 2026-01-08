-- Migración 000053: Corregir SOLO el avance de aportes
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de aportes
-- Partir de la migración 000050 (no tocar la 000051, 000052 ni 000053)

DROP VIEW IF EXISTS dashboard_contributions_progress_view;

CREATE OR REPLACE VIEW dashboard_contributions_progress_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    
    -- Campos de siembra (placeholder)
    0::numeric(14,2) AS sowing_hectares,
    0::numeric(14,2) AS sowing_total_hectares,
    0::numeric(6,2) AS sowing_progress_pct,
    
    -- Campos de cosecha (placeholder)
    0::numeric(14,2) AS harvest_hectares,
    0::numeric(14,2) AS harvest_total_hectares,
    0::numeric(6,2) AS harvest_progress_pct,
    
    -- Campos de costos (placeholder)
    0::numeric(14,2) AS executed_costs_usd,
    0::numeric(14,2) AS executed_labors_usd,
    0::numeric(14,2) AS executed_supplies_usd,
    0::numeric(14,2) AS budget_cost_usd,
    0::numeric(14,2) AS budget_total_usd,
    0::numeric(6,2) AS costs_progress_pct,
    
    -- Campos de resultado operativo (placeholder)
    0::numeric(14,2) AS income_usd,
    0::numeric(14,2) AS operating_result_usd,
    0::numeric(6,2) AS operating_result_pct,
    0::numeric(14,2) AS operating_result_total_costs_usd,
    
    -- Campos de aportes (REALES)
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    100.00::numeric(6,2) AS contributions_progress_pct,
    
    -- Campos de cultivos (placeholder)
    0::bigint AS crop_id,
    '' AS crop_name,
    0::numeric(14,2) AS crop_hectares,
    0::numeric(14,2) AS project_total_hectares,
    0::numeric(6,2) AS incidence_pct,
    0::numeric(14,2) AS crop_direct_costs_usd,
    0::numeric(14,2) AS cost_per_ha_usd,
    
    -- Campos de balance de gestión (placeholder)
    0::numeric(14,2) AS balance_executed_costs_usd,
    0::numeric(14,2) AS balance_budget_cost_usd,
    0::numeric(14,2) AS balance_operating_result_total_costs_usd,
    0::numeric(14,2) AS balance_operating_result_usd,
    0::numeric(6,2) AS balance_operating_result_pct,
    
    -- Campos de fechas (placeholder)
    NULL::timestamp AS primera_orden_fecha,
    0::bigint AS primera_orden_id,
    NULL::timestamp AS ultima_orden_fecha,
    0::bigint AS ultima_orden_id,
    NULL::timestamp AS arqueo_stock_fecha,
    NULL::timestamp AS cierre_campana_fecha,
    
    -- Campo de tipo de fila
    'metric'::text AS row_kind
    
FROM projects p
JOIN fields f ON f.project_id=p.id
JOIN project_investors pi ON pi.project_id=p.id
JOIN investors i ON i.id=pi.investor_id;
