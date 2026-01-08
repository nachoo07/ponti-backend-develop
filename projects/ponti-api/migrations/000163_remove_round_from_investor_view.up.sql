-- ========================================
-- MIGRATION 000163: REMOVE ROUND FROM INVESTOR VIEW (UP)
-- ========================================
-- 
-- Purpose: Eliminar ROUND() de v3_investor_contribution_data_view para mantener precisión completa
-- Date: 2025-10-21
-- Author: System
-- 
-- Problema identificado:
--   La vista v3_investor_contribution_data_view contiene múltiples ROUND() que reducen la precisión
--   de los datos monetarios y de superficie
--
-- Corrección:
--   Eliminar todos los ROUND() de la vista, dejando que los DTOs de salida manejen el formato
--
-- Impacto:
--   - Solo afecta informe de Aportes por Inversor
--   - Los datos mantendrán precisión completa desde la base de datos
--   - Los DTOs de salida aplicarán el formato visual correcto
--
-- REGLA CRÍTICA: NUNCA usar ROUND() en migraciones SQL, mantener precisión completa
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- RECREAR VISTA: v3_investor_contribution_data_view (SIN ROUND)
-- ============================================================================

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

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
  
  jsonb_build_object(
    'surface_total_ha', pb.surface_total_ha,
    'lease_fixed_total_usd', pb.lease_fixed_total_usd,
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', pb.lease_per_ha_usd,
    'admin_total_usd', pb.admin_total_usd,
    'admin_per_ha_usd', pb.admin_per_ha_usd
  ) AS general_project_data,
  
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
      SELECT 
        'agrochemicals'::text AS key, 1 AS sort_index, 'pre_harvest'::text AS type, 'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name, 
          'amount_usd', id.agrochemicals_real_usd, 
          'share_pct', CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'seeds'::text, 2, 'pre_harvest'::text, 'Semilla'::text, cc.seeds_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.seeds_real_usd,
          'share_pct', CASE WHEN cc.seeds_total_usd > 0 THEN (id.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'general_labors'::text, 3, 'pre_harvest'::text, 'Labores Generales'::text, cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.general_labors_real_usd,
          'share_pct', CASE WHEN cc.general_labors_total_usd > 0 THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text, cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.sowing_real_usd,
          'share_pct', CASE WHEN cc.sowing_total_usd > 0 THEN (id.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'irrigation'::text, 5, 'pre_harvest'::text, 'Riego'::text, cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.irrigation_real_usd,
          'share_pct', CASE WHEN cc.irrigation_total_usd > 0 THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'capitalizable_lease'::text, 6, 'pre_harvest'::text, 'Arriendo Capitalizable'::text, cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.rent_real_usd,
          'share_pct', CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, 'Distribuido proporcionalmente según aportes reales'
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'administration_structure'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text, cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', id.administration_real_usd,
          'share_pct', CASE WHEN cc.administration_total_usd > 0 THEN (id.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, 'Distribuido proporcionalmente según aportes reales'
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
    ) cat_data
  ) AS contribution_categories,
  
  jsonb_build_object(
    'total_usd', (SELECT SUM(
      cc.agrochemicals_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
      cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
      cc.administration_total_usd
    ) FROM contribution_categories cc WHERE cc.project_id = pb.project_id),
    'total_usd_ha', CASE 
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
          'amount_usd', id.real_contribution_usd,
          'share_pct', CASE 
              WHEN id.project_total_contributions_usd > 0 
              THEN (id.real_contribution_usd / id.project_total_contributions_usd * 100)
              ELSE 0 
            END
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    )
  ) AS pre_harvest,
  
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct_agreed', id.share_pct_agreed,
        'agreed_contribution_usd', id.agreed_contribution_usd,
        'real_contribution_usd', id.real_contribution_usd,
        'adjustment_usd', id.adjustment_usd
      )
      ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'label', 'Cosecha',
        'total_usd', hd.total_harvest_usd,
        'total_usd_ha', CASE WHEN pb.surface_total_ha > 0 THEN (hd.total_harvest_usd / pb.surface_total_ha) ELSE 0 END,
        'investors', (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', (COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100),
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
          'amount_usd', (COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100),
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
          'amount_usd', (0 - (COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100))
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    )
  ) AS harvest_settlement

FROM project_base pb
LEFT JOIN contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista 4/4 para informe de Aportes por Inversor. ACTUALIZADO 163: Elimina ROUND() para mantener precisión completa.';

