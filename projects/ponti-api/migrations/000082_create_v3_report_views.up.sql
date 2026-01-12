-- ========================================
-- MIGRATION 000082: CREATE v3_report_views (UP)
-- ========================================
-- 
-- Purpose: Create report field crop metrics view
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_report_field_crop_metrics_view: métricas por field y cultivo
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_field_crop_metrics_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)::double precision AS seeded_area_ha,
    v3_calc.harvested_area(l.tons::numeric, l.hectares::numeric)::double precision AS harvested_area_ha
  FROM public.lots l
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
)
SELECT
  lb.project_id,
  lb.field_id,
  lb.field_name::text AS field_name,
  lb.current_crop_id,
  lb.crop_name::text  AS crop_name,

  COALESCE(SUM(v3_calc.income_net_total_for_lot(lb.lot_id)), 0)             AS income_usd,

  -- Costos directos ejecutados (labores + insumos)
  COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)                  AS direct_costs_executed_usd,

  -- Costos directos invertidos (labores+insumos+arriendo+estructura)
  (
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
    + COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
    + COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision
  )                                                                        AS direct_costs_invested_usd,

  -- Componentes de invertidos
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)     AS rent_invested_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS structure_invested_usd,

  -- Resultado operativo y ratio
  COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0) AS operating_result_usd,
  v3_calc.renta_pct(
    COALESCE(SUM(v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::double precision,
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::double precision
  )                                                                          AS operating_result_pct
FROM lot_base lb
LEFT JOIN public.crop_commercializations cc
  ON cc.project_id = lb.project_id
 AND cc.crop_id   = lb.current_crop_id
 AND cc.deleted_at IS NULL
WHERE lb.current_crop_id IS NOT NULL
GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name;

-- -------------------------------------------------------------------
-- v3_report_summary_results_view: resumen de resultados generales por cultivo
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_report_summary_results_view AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares,
    COALESCE(l.tons, 0)::numeric AS tons,
    -- Superficie sembrada: suma de effective_area de workorders de tipo siembra
    COALESCE((
      SELECT SUM(w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id
      JOIN public.categories cat ON lab.category_id = cat.id
      WHERE w.lot_id = l.id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS seeded_area_ha
  FROM public.lots l
  JOIN public.fields   f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares > 0
),
by_crop AS (
  SELECT
    lb.project_id,
    lb.current_crop_id,
    lb.crop_name::text AS crop_name,
    
    -- Superficie total por cultivo
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS surface_ha,
    
    -- Ingreso neto total por cultivo: superficie * ingreso neto por ha (de comercializaciones)
    COALESCE(SUM(lb.seeded_area_ha * (
      SELECT COALESCE(cc.net_price, 0)
      FROM public.crop_commercializations cc
      WHERE cc.project_id = lb.project_id 
        AND cc.crop_id = lb.current_crop_id
        AND cc.deleted_at IS NULL
      LIMIT 1
    )), 0)::numeric AS net_income_usd,
    
    -- Costos directos totales por cultivo (labores + insumos)
    COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric AS direct_costs_usd,
    
    -- Arriendo total por cultivo: superficie * costo de arriendo por ha (usando función SSOT)
    COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric AS rent_usd,
    
    -- Estructura total por cultivo: costo fijo * superficie (usando función SSOT)
    COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric AS structure_usd,
    
    -- Total invertido por cultivo: Costos directos + Arriendo + Estructura
    (
      COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric
      + COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric
      + COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric
    ) AS total_invested_usd,
    
    -- Resultado operativo total por cultivo: Ingreso Neto - Total Invertido
    (
      COALESCE(SUM(lb.seeded_area_ha * (
        SELECT COALESCE(cc.net_price, 0)
        FROM public.crop_commercializations cc
        WHERE cc.project_id = lb.project_id 
          AND cc.crop_id = lb.current_crop_id
          AND cc.deleted_at IS NULL
        LIMIT 1
      )), 0)::numeric
      - (
        COALESCE(SUM(v3_calc.direct_cost_for_lot(lb.lot_id)), 0)::numeric
        + COALESCE(SUM(lb.seeded_area_ha * v3_calc.rent_per_ha_for_lot(lb.lot_id)), 0)::numeric
        + COALESCE(SUM(lb.seeded_area_ha * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)), 0)::numeric
      )
    ) AS operating_result_usd
    
  FROM lot_base lb
  WHERE lb.current_crop_id IS NOT NULL
  GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
),
project_totals AS (
  SELECT
    project_id,
    SUM(surface_ha)::numeric AS total_surface_ha,
    SUM(net_income_usd)::numeric AS total_net_income_usd,
    SUM(direct_costs_usd)::numeric AS total_direct_costs_usd,
    SUM(rent_usd)::numeric AS total_rent_usd,
    SUM(structure_usd)::numeric AS total_structure_usd,
    SUM(total_invested_usd)::numeric AS total_invested_usd,
    SUM(operating_result_usd)::numeric AS total_operating_result_usd
  FROM by_crop
  GROUP BY project_id
)
SELECT
  bc.project_id,
  bc.current_crop_id,
  bc.crop_name,
  bc.surface_ha,
  bc.net_income_usd,
  bc.direct_costs_usd,
  bc.rent_usd,
  bc.structure_usd,
  bc.total_invested_usd,
  bc.operating_result_usd,
  
  -- Renta del cultivo (%)
  v3_calc.renta_pct(
    bc.operating_result_usd::double precision,
    bc.total_invested_usd::double precision
  )::numeric AS crop_return_pct,
  
  -- Totales del proyecto (para comparación)
  pt.total_surface_ha,
  pt.total_net_income_usd,
  pt.total_direct_costs_usd,
  pt.total_rent_usd,
  pt.total_structure_usd,
  pt.total_invested_usd AS total_invested_project_usd,
  pt.total_operating_result_usd,
  
  -- Renta total del proyecto (%)
  v3_calc.renta_pct(
    pt.total_operating_result_usd::double precision,
    pt.total_invested_usd::double precision
  )::numeric AS project_return_pct

FROM by_crop bc
JOIN project_totals pt ON pt.project_id = bc.project_id
ORDER BY bc.project_id, bc.current_crop_id;

-- -------------------------------------------------------------------
-- v3_investor_contribution_data_view: datos de contribución de inversores v3
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH project_base AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.customer_id,
    c.name AS customer_name,
    p.campaign_id,
    cam.name AS campaign_name,
    -- Calcular superficie total usando funciones SSOT
    COALESCE(SUM(l.hectares), 0)::numeric AS surface_total_ha,
    -- Datos de arriendo usando funciones SSOT - solo si es fijo
    COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS lease_fixed_usd,
    CASE 
      WHEN COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0) > 0 
      THEN true 
      ELSE false 
    END AS lease_is_fixed,
    -- Datos de administración usando funciones SSOT
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_per_ha_usd,
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_total_usd
  FROM public.projects p
  JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
  JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
  LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, p.name, p.customer_id, c.name, p.campaign_id, cam.name
),
-- Categorías de aportes según documentación
contribution_categories AS (
  SELECT
    pb.project_id,
    -- Agroquímicos: supplies con categorías de agroquímicos
    COALESCE((
      SELECT SUM(wi.total_used * s.price)
      FROM workorders w
      JOIN workorder_items wi ON w.id = wi.workorder_id
      JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
      JOIN categories cat ON s.category_id = cat.id
      WHERE w.project_id = pb.project_id 
        AND w.deleted_at IS NULL
        AND cat.type_id = 2
        AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
    ), 0)::numeric AS agrochemicals_total,
    
    -- Semillas: supplies con categoría de semillas
    COALESCE((
      SELECT SUM(wi.total_used * s.price)
      FROM workorders w
      JOIN workorder_items wi ON w.id = wi.workorder_id
      JOIN supplies s ON wi.supply_id = s.id AND s.deleted_at IS NULL
      JOIN categories cat ON s.category_id = cat.id
      WHERE w.project_id = pb.project_id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Semilla'
        AND cat.type_id = 1
    ), 0)::numeric AS seeds_total,
    
    -- Labores Generales: labores que NO son siembra, riego ni cosecha
    COALESCE((
      SELECT SUM(l.price * w.effective_area)
      FROM workorders w
      JOIN labors l ON w.labor_id = l.id AND l.deleted_at IS NULL
      JOIN categories cat ON l.category_id = cat.id
      WHERE w.project_id = pb.project_id 
        AND w.deleted_at IS NULL
        AND cat.type_id = 4
        AND cat.name IN ('Pulverización', 'Otras Labores')
    ), 0)::numeric AS general_labors_total,
    
    -- Siembra: labores de siembra
    COALESCE((
      SELECT SUM(l.price * w.effective_area)
      FROM workorders w
      JOIN labors l ON w.labor_id = l.id AND l.deleted_at IS NULL
      JOIN categories cat ON l.category_id = cat.id
      WHERE w.project_id = pb.project_id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Siembra'
        AND cat.type_id = 4
    ), 0)::numeric AS sowing_total,
    
    -- Riego: labores de riego
    COALESCE((
      SELECT SUM(l.price * w.effective_area)
      FROM workorders w
      JOIN labors l ON w.labor_id = l.id AND l.deleted_at IS NULL
      JOIN categories cat ON l.category_id = cat.id
      WHERE w.project_id = pb.project_id 
        AND w.deleted_at IS NULL
        AND cat.name = 'Riego'
        AND cat.type_id = 4
    ), 0)::numeric AS irrigation_total,
    
    -- Arriendo: solo si es fijo (ya calculado en project_base)
    pb.lease_fixed_usd AS rent_total,
    
    -- Administración: ya calculado en project_base
    pb.admin_total_usd AS administration_total
    
  FROM project_base pb
),
-- Aportes reales por inversor (simplificado - asumiendo distribución igual)
investor_contributions AS (
  SELECT
    pb.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage,
    -- Calcular aportes reales por categoría (simplificado)
    (cc.agrochemicals_total * pi.percentage / 100) AS agrochemicals_contribution,
    (cc.seeds_total * pi.percentage / 100) AS seeds_contribution,
    (cc.general_labors_total * pi.percentage / 100) AS general_labors_contribution,
    (cc.sowing_total * pi.percentage / 100) AS sowing_contribution,
    (cc.irrigation_total * pi.percentage / 100) AS irrigation_contribution,
    -- Arriendo y administración requieren carga manual
    0 AS rent_contribution,
    0 AS administration_contribution
  FROM project_base pb
  JOIN contribution_categories cc ON cc.project_id = pb.project_id
  JOIN project_investors pi ON pi.project_id = pb.project_id
  JOIN investors i ON pi.investor_id = i.id
),
contributions_data AS (
  SELECT
    pb.project_id,
    -- Contributions data as JSON - construido según documentación
    (
      SELECT COALESCE(jsonb_agg(
        jsonb_build_object(
          'type', cat_data.type,
          'label', cat_data.label,
          'total_usd', cat_data.total_usd,
          'total_usd_ha', CASE 
            WHEN pb.surface_total_ha > 0 
            THEN cat_data.total_usd / pb.surface_total_ha
            ELSE 0 
          END,
          'investors', cat_data.investors,
          'requires_manual_attribution', cat_data.requires_manual_attribution
        )
      ), '[]'::jsonb)
      FROM (
        -- Agroquímicos
        SELECT 
          'agrochemicals'::text AS type,
          'Agroquímicos'::text AS label,
          cc.agrochemicals_total AS total_usd,
          COALESCE((
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', ic.investor_id,
                'investor_name', ic.investor_name,
                'amount_usd', ic.agrochemicals_contribution,
                'share_pct', CASE 
                  WHEN cc.agrochemicals_total > 0 
                  THEN (ic.agrochemicals_contribution / cc.agrochemicals_total * 100)
                  ELSE 0 
                END
              )
            )
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id
          ), '[]'::jsonb) AS investors,
          false AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id
        
        UNION ALL
        
        -- Semillas
        SELECT 
          'seeds'::text AS type,
          'Semilla'::text AS label,
          cc.seeds_total AS total_usd,
          COALESCE((
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', ic.investor_id,
                'investor_name', ic.investor_name,
                'amount_usd', ic.seeds_contribution,
                'share_pct', CASE 
                  WHEN cc.seeds_total > 0 
                  THEN (ic.seeds_contribution / cc.seeds_total * 100)
                  ELSE 0 
                END
              )
            )
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id
          ), '[]'::jsonb) AS investors,
          false AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id
        
        UNION ALL
        
        -- Labores Generales
        SELECT 
          'general_labors'::text AS type,
          'Labores Generales'::text AS label,
          cc.general_labors_total AS total_usd,
          COALESCE((
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', ic.investor_id,
                'investor_name', ic.investor_name,
                'amount_usd', ic.general_labors_contribution,
                'share_pct', CASE 
                  WHEN cc.general_labors_total > 0 
                  THEN (ic.general_labors_contribution / cc.general_labors_total * 100)
                  ELSE 0 
                END
              )
            )
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id
          ), '[]'::jsonb) AS investors,
          false AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id
        
        UNION ALL
        
        -- Siembra
        SELECT 
          'sowing'::text AS type,
          'Siembra'::text AS label,
          cc.sowing_total AS total_usd,
          COALESCE((
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', ic.investor_id,
                'investor_name', ic.investor_name,
                'amount_usd', ic.sowing_contribution,
                'share_pct', CASE 
                  WHEN cc.sowing_total > 0 
                  THEN (ic.sowing_contribution / cc.sowing_total * 100)
                  ELSE 0 
                END
              )
            )
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id
          ), '[]'::jsonb) AS investors,
          false AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id
        
        UNION ALL
        
        -- Riego
        SELECT 
          'irrigation'::text AS type,
          'Riego'::text AS label,
          cc.irrigation_total AS total_usd,
          COALESCE((
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', ic.investor_id,
                'investor_name', ic.investor_name,
                'amount_usd', ic.irrigation_contribution,
                'share_pct', CASE 
                  WHEN cc.irrigation_total > 0 
                  THEN (ic.irrigation_contribution / cc.irrigation_total * 100)
                  ELSE 0 
                END
              )
            )
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id
          ), '[]'::jsonb) AS investors,
          false AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id
        
        UNION ALL
        
        -- Arriendo Capitalizable
        SELECT 
          'capitalizable_lease'::text AS type,
          'Arriendo Capitalizable'::text AS label,
          cc.rent_total AS total_usd,
          '[]'::jsonb AS investors,
          true AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id AND cc.rent_total > 0
        
        UNION ALL
        
        -- Administración y Estructura
        SELECT 
          'administration_structure'::text AS type,
          'Administración y Estructura'::text AS label,
          cc.administration_total AS total_usd,
          '[]'::jsonb AS investors,
          true AS requires_manual_attribution
        FROM contribution_categories cc
        WHERE cc.project_id = pb.project_id AND cc.administration_total > 0
        
      ) cat_data
    ) as contributions_data
  FROM project_base pb
),
-- Cálculo de totales para comparación
project_totals AS (
  SELECT
    pb.project_id,
    (cc.agrochemicals_total + cc.seeds_total + cc.general_labors_total + 
     cc.sowing_total + cc.irrigation_total + cc.rent_total + cc.administration_total) AS total_contributions
  FROM project_base pb
  JOIN contribution_categories cc ON cc.project_id = pb.project_id
),
comparison_data AS (
  SELECT
    pb.project_id,
    -- Comparison data as JSON - calculado según documentación
    (
      SELECT COALESCE(jsonb_agg(
        jsonb_build_object(
          'investor_id', pi.investor_id,
          'investor_name', i.name,
          'agreed_share_pct', pi.percentage,
          'agreed_usd', pt.total_contributions * (pi.percentage / 100),
          'actual_usd', COALESCE((
            SELECT SUM(ic.agrochemicals_contribution + ic.seeds_contribution + 
                      ic.general_labors_contribution + ic.sowing_contribution + 
                      ic.irrigation_contribution + ic.rent_contribution + 
                      ic.administration_contribution)
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id AND ic.investor_id = pi.investor_id
          ), 0),
          'adjustment_usd', COALESCE((
            SELECT SUM(ic.agrochemicals_contribution + ic.seeds_contribution + 
                      ic.general_labors_contribution + ic.sowing_contribution + 
                      ic.irrigation_contribution + ic.rent_contribution + 
                      ic.administration_contribution)
            FROM investor_contributions ic
            WHERE ic.project_id = pb.project_id AND ic.investor_id = pi.investor_id
          ), 0) - (pt.total_contributions * (pi.percentage / 100))
        )
      ), '[]'::jsonb)
      FROM project_investors pi
      JOIN investors i ON pi.investor_id = i.id
      JOIN project_totals pt ON pt.project_id = pb.project_id
      WHERE pi.project_id = pb.project_id
    ) as comparison_data
  FROM project_base pb
),
harvest_data AS (
  SELECT
    pb.project_id,
    -- Harvest data as JSON - calculado según documentación
    jsonb_build_object(
      'total_harvest_usd', COALESCE((
        SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
        FROM lots l
        JOIN fields f ON l.field_id = f.id
        WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
      ), 0),
      'total_harvest_usd_ha', CASE 
        WHEN pb.surface_total_ha > 0 
        THEN COALESCE((
          SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
          FROM lots l
          JOIN fields f ON l.field_id = f.id
          WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
        ), 0) / pb.surface_total_ha
        ELSE 0 
      END,
      'investors', COALESCE((
        SELECT jsonb_agg(
          jsonb_build_object(
            'investor_id', pi.investor_id,
            'investor_name', i.name,
            'paid_usd', 0, -- Los pagos reales de cosecha no están en la BD actualmente
            'agreed_usd', COALESCE((
              SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
              FROM lots l
              JOIN fields f ON l.field_id = f.id
              WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
            ), 0) * (pi.percentage / 100),
            'adjustment_usd', 0 - (COALESCE((
              SELECT SUM(v3_calc.income_net_total_for_lot(l.id))
              FROM lots l
              JOIN fields f ON l.field_id = f.id
              WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
            ), 0) * (pi.percentage / 100)) -- Ajuste = Pagado - Acordado
          )
        )
        FROM project_investors pi
        JOIN investors i ON pi.investor_id = i.id
        WHERE pi.project_id = pb.project_id
      ), '[]'::jsonb)
    ) as harvest_data
  FROM project_base pb
)
SELECT 
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  pb.surface_total_ha,
  pb.lease_fixed_usd,
  pb.lease_is_fixed,
  pb.admin_per_ha_usd,
  pb.admin_total_usd,
  cd.contributions_data,
  compd.comparison_data,
  hd.harvest_data
FROM project_base pb
LEFT JOIN contributions_data cd ON cd.project_id = pb.project_id
LEFT JOIN comparison_data compd ON compd.project_id = pb.project_id
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id;
