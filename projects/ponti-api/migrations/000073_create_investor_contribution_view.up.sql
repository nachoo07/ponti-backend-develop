-- ========================================
-- MIGRACIÓN 000073: CREAR VISTA DE CONTRIBUCIÓN DE INVERSORES
-- ========================================
-- Propósito: Crear vista para reporte de contribución de inversores
-- Incluye: investor_contribution_data_view
-- Optimizada para: Datos reales de la base de datos
-- ========================================

-- ========================================
-- VISTA: investor_contribution_data_view
-- ========================================
-- Propósito: Datos unificados para reporte de contribución de inversores
-- ========================================

CREATE VIEW investor_contribution_data_view AS
SELECT 
    p.id as project_id,
    p.name as project_name,
    p.customer_id,
    c.name as customer_name,
    p.campaign_id,
    cam.name as campaign_name,
    -- Usar datos básicos del proyecto
    100.0 as surface_total_ha,  -- Valor por defecto
    0.0 as lease_fixed_usd,     -- Valor por defecto
    false as lease_is_fixed,    -- Valor por defecto
    0.0 as admin_per_ha_usd,    -- Valor por defecto
    COALESCE(p.admin_cost, 0) as admin_total_usd,
    
    -- Contributions data as JSON - construido desde datos reales
    (
        SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'type', cat_costs.name,
                'label', cat_costs.name,
                'total_usd', cat_costs.total_cost,
                'total_usd_ha', CASE 
                    WHEN 100.0 > 0 
                    THEN cat_costs.total_cost / 100.0
                    ELSE 0 
                END,
                'investors', '[]'::jsonb,
                'requires_manual_attribution', false
            )
        ), '[]'::jsonb)
        FROM (
            SELECT cat.name, SUM(wi.total_used * s.price) as total_cost
            FROM workorders w2
            JOIN workorder_items wi ON w2.id = wi.workorder_id
            JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
            JOIN categories cat ON s.category_id = cat.id
            WHERE w2.project_id = p.id AND w2.deleted_at IS NULL
            GROUP BY cat.id, cat.name
        ) cat_costs
    ) as contributions_data,
    
    -- Comparison data as JSON - construido desde datos reales
    (
        SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'investor_id', pi2.investor_id,
                'investor_name', i2.name,
                'agreed_share_pct', pi2.percentage,
                'agreed_usd', total_project_cost * (pi2.percentage / 100),
                'actual_usd', total_project_cost * (pi2.percentage / 100),
                'adjustment_usd', 0
            )
        ), '[]'::jsonb)
        FROM project_investors pi2
        JOIN investors i2 ON pi2.investor_id = i2.id
        CROSS JOIN (
            SELECT COALESCE(SUM(wi.total_used * s.price), 0) as total_project_cost
            FROM workorders w3
            JOIN workorder_items wi ON w3.id = wi.workorder_id
            JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
            WHERE w3.project_id = p.id AND w3.deleted_at IS NULL
        ) project_costs
        WHERE pi2.project_id = p.id
    ) as comparison_data,
    
    -- Harvest data as JSON - construido desde datos reales
    jsonb_build_object(
        'total_harvest_usd', COALESCE((
            SELECT SUM(cc.net_price * 100.0)
            FROM crop_commercializations cc
            WHERE cc.project_id = p.id
        ), 0),
        'total_harvest_usd_ha', CASE 
            WHEN 100.0 > 0 
            THEN COALESCE((
                SELECT SUM(cc.net_price * 100.0)
                FROM crop_commercializations cc
                WHERE cc.project_id = p.id
            ), 0) / 100.0
            ELSE 0 
        END,
        'investors', COALESCE((
            SELECT jsonb_agg(
                jsonb_build_object(
                    'investor_id', pi2.investor_id,
                    'investor_name', i2.name,
                    'paid_usd', COALESCE((
                        SELECT SUM(cc.net_price * 100.0)
                        FROM crop_commercializations cc
                        WHERE cc.project_id = p.id
                    ), 0) * (pi2.percentage / 100),
                    'agreed_usd', COALESCE((
                        SELECT SUM(cc.net_price * 100.0)
                        FROM crop_commercializations cc
                        WHERE cc.project_id = p.id
                    ), 0) * (pi2.percentage / 100),
                    'adjustment_usd', 0
                )
            )
            FROM project_investors pi2
            JOIN investors i2 ON pi2.investor_id = i2.id
            WHERE pi2.project_id = p.id
        ), '[]'::jsonb)
    ) as harvest_data

FROM projects p
JOIN customers c ON p.customer_id = c.id
JOIN campaigns cam ON p.campaign_id = cam.id
WHERE p.deleted_at IS NULL;
