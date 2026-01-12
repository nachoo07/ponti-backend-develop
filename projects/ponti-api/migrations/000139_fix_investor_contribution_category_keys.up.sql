-- +goose Up
-- Corrige los keys de las categorías de Arriendo y Administración 
-- para que coincidan con lo que espera el frontend
-- Cambios: 'rent_capitalizable' → 'capitalizable_lease', 'administration' → 'administration_structure'

DROP VIEW IF EXISTS v3_investor_contribution_data_view CASCADE;

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH project_base AS (
  SELECT
    pb.project_id,
    pb.project_name,
    pb.customer_id,
    pb.customer_name,
    pb.campaign_id,
    pb.campaign_name,
    pb.surface_total_ha,
    pb.lease_fixed_total_usd,
    pb.lease_is_fixed,
    pb.lease_per_ha_usd,
    pb.admin_total_usd,
    pb.admin_per_ha_usd
  FROM v3_report_investor_project_base pb
),
contribution_categories AS (
  SELECT
    cc.project_id,
    cc.agrochemicals_total_usd,
    cc.seeds_total_usd,
    cc.general_labors_total_usd,
    cc.sowing_total_usd,
    cc.irrigation_total_usd,
    cc.rent_capitalizable_total_usd,
    cc.administration_total_usd,
    cc.total_seeded_area_ha
  FROM v3_report_investor_contribution_categories cc
),
harvest_data AS (
  SELECT
    pb.project_id,
    COALESCE((
      SELECT SUM(v3_lot_ssot.income_net_total_for_lot(l.id))
      FROM public.lots l
      JOIN public.fields f ON l.field_id = f.id
      WHERE f.project_id = pb.project_id AND l.deleted_at IS NULL
    ), 0)::numeric AS total_harvest_usd
  FROM project_base pb
)
SELECT 
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  -- investor_headers
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct', id.share_pct_agreed
      )
      ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_headers,
  
  -- general_project_data
  jsonb_build_object(
    'surface_total_ha', ROUND(pb.surface_total_ha::numeric, 2),
    'lease_fixed_total_usd', ROUND(pb.lease_fixed_total_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2),
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2)
  ) AS general_project_data,
  
  -- contribution_categories (con investors actualizados)
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'key', cat_data.key,
        'sort_index', cat_data.sort_index,
        'type', cat_data.type,
        'label', cat_data.label,
        'total_usd', cat_data.total_usd,
        'total_usd_ha', cat_data.total_usd_ha,
        'investors', cat_data.investors,
        'requires_manual_attribution', cat_data.requires_manual_attribution,
        'attribution_note', cat_data.attribution_note
      )
      ORDER BY cat_data.sort_index
    )
    FROM (
      -- Agroquímicos (con aportes REALES)
      SELECT 
        'agrochemicals'::text AS key, 1 AS sort_index, 'pre_harvest'::text AS type, 'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name, 
          'amount_usd', ROUND(id.agrochemicals_real_usd::numeric, 2), 
          'share_pct', ROUND(CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Semilla (con aportes REALES)
      SELECT 'seeds'::text, 2, 'pre_harvest'::text, 'Semilla'::text, cc.seeds_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.seeds_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.seeds_total_usd > 0 THEN (id.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Labores Generales (con aportes REALES)
      SELECT 'general_labors'::text, 3, 'pre_harvest'::text, 'Labores Generales'::text, cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.general_labors_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.general_labors_total_usd > 0 THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Siembra (con aportes REALES)
      SELECT 'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text, cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.sowing_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.sowing_total_usd > 0 THEN (id.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Riego (con aportes REALES)
      SELECT 'irrigation'::text, 5, 'pre_harvest'::text, 'Riego'::text, cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.irrigation_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.irrigation_total_usd > 0 THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Arriendo Capitalizable (distribución proporcional)
      -- CAMBIO: 'rent_capitalizable' → 'capitalizable_lease'
      SELECT 'capitalizable_lease'::text, 6, 'pre_harvest'::text, 'Arriendo Capitalizable'::text, cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.rent_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        true, 'Requiere atribución manual por inversor'
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Administración y Estructura (distribución proporcional)
      -- CAMBIO: 'administration' → 'administration_structure'
      SELECT 'administration_structure'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text, cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.administration_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.administration_total_usd > 0 THEN (id.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        true, 'Requiere atribución manual por inversor'
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
    ) cat_data
  ) AS contribution_categories,
  
  -- pre_harvest (totales)
  jsonb_build_object(
    'total_usd', (SELECT SUM(
      cc.agrochemicals_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
      cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
      cc.administration_total_usd
    ) FROM contribution_categories cc WHERE cc.project_id = pb.project_id),
    'total_us_ha', CASE 
      WHEN pb.surface_total_ha > 0 
      THEN (SELECT SUM(
        cc.agrochemicals_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
        cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
        cc.administration_total_usd
      ) / pb.surface_total_ha FROM contribution_categories cc WHERE cc.project_id = pb.project_id)
      ELSE 0 
    END,
    'investors', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND(id.real_contribution_usd::numeric, 2),
          'share_pct', ROUND(
            CASE 
              WHEN (SELECT SUM(id2.real_contribution_usd) 
                    FROM v3_report_investor_distributions id2 
                    WHERE id2.project_id = pb.project_id) > 0
              THEN (id.real_contribution_usd / 
                    (SELECT SUM(id2.real_contribution_usd) 
                     FROM v3_report_investor_distributions id2 
                     WHERE id2.project_id = pb.project_id) * 100)
              ELSE 0
            END::numeric, 2
          )
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    )
  ) AS pre_harvest,
  
  -- investor_contribution_comparison
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'agreed_share_pct', id.share_pct_agreed,
        'agreed_usd', id.agreed_contribution_usd,
        'actual_usd', id.real_contribution_usd,
        'adjustment_usd', id.adjustment_usd
      )
      ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  -- harvest_settlement
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'key', '',
        'type', '',
        'total_usd', ROUND(hd.total_harvest_usd, 2),
        'total_us_ha', CASE 
          WHEN pb.surface_total_ha > 0 
          THEN ROUND(hd.total_harvest_usd / pb.surface_total_ha, 2) 
          ELSE 0 
        END,
        'investors', (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100), 2),
              'share_pct', id.share_pct_agreed
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        )
      )
    ),
    'footer_payment_agreed', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100), 2),
          'share_pct', id.share_pct_agreed
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    ),
    'footer_payment_adjustment', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100) - id.adjustment_usd, 2),
          'share_pct', 0
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    )
  ) AS harvest_settlement

FROM project_base pb
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id;

COMMENT ON VIEW v3_investor_contribution_data_view IS 
  'Vista consolidada para informe de aportes de inversores v3. Keys corregidos para compatibilidad con frontend: capitalizable_lease, administration_structure.';
