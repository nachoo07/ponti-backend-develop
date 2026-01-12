-- ========================================
-- MIGRACIÓN 000171: FIX General Project Data Structure (UP)
-- ========================================
-- 
-- Propósito: Restaurar los campos de general_project_data que se perdieron en migración 000168
-- Problema: La migración 000168 cambió la estructura de general_project_data, eliminando surface_total_ha, admin_total_usd, etc.
-- Solución: Restaurar la estructura original con los campos necesarios
-- Fecha: 2025-10-31
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español

BEGIN;

-- ========================================
-- RECREAR VISTA: v3_investor_contribution_data_view
-- ========================================
-- Propósito: Restaurar estructura correcta de general_project_data

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS

-- Cálculo de los totales de contribución por proyecto
WITH contribution_totals AS (
  SELECT
    cc.project_id,
    -- Sumar todas las categorías de aportes
    COALESCE(cc.agrochemicals_total_usd, 0) +
    COALESCE(cc.fertilizers_total_usd, 0) +
    COALESCE(cc.seeds_total_usd, 0) +
    COALESCE(cc.general_labors_total_usd, 0) +
    COALESCE(cc.sowing_total_usd, 0) +
    COALESCE(cc.irrigation_total_usd, 0) +
    COALESCE(cc.rent_capitalizable_total_usd, 0) +
    COALESCE(cc.administration_total_usd, 0) AS total_contributions_usd,
    cc.total_seeded_area_ha
  FROM v3_report_investor_contribution_categories cc
)

SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  -- =========================================================================
  -- SECCIÓN 1: CABECERAS DE INVERSORES (headers con % acordado)
  -- =========================================================================
  (
   SELECT jsonb_agg(jsonb_build_object(
      'investor_id', id.investor_id,
      'investor_name', id.investor_name,
      'share_pct', id.share_pct_agreed
    ) ORDER BY id.investor_id)
   FROM v3_report_investor_distributions id
   WHERE id.project_id = pb.project_id
  ) AS investor_headers,
  
  -- =========================================================================
  -- SECCIÓN 2: DATOS GENERALES DEL PROYECTO (Card Superficie)
  -- =========================================================================
  jsonb_build_object(
    'surface_total_ha', COALESCE(cc.total_seeded_area_ha, 0),
    'lease_fixed_total_usd', COALESCE(pb.lease_fixed_total_usd, 0),
    'lease_is_fixed', COALESCE(pb.lease_is_fixed, false),
    'admin_per_ha_usd', CASE 
      WHEN COALESCE(cc.total_seeded_area_ha, 0) > 0 
      THEN COALESCE(cc.administration_total_usd, 0) / cc.total_seeded_area_ha
      ELSE 0 
    END,
    'admin_total_usd', COALESCE(cc.administration_total_usd, 0)
  ) AS general_project_data,
  
  -- =========================================================================
  -- SECCIÓN 3: CATEGORÍAS DE APORTE (rows de la tabla principal)
  -- =========================================================================
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
      ) ORDER BY cat_data.sort_index
    )
    FROM (
      -- Seeds
      SELECT 
        'seeds' AS key,
        1 AS sort_index,
        'pre_harvest' AS type,
        'Semilla' AS label,
        cc.seeds_total_usd AS total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.seeds_real_usd,
            'share_pct', CASE WHEN cc.seeds_total_usd > 0 THEN (id.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution,
        NULL AS attribution_note

      UNION ALL

      -- Agrochemicals
      SELECT 
        'agrochemicals',
        2,
        'pre_harvest',
        'Agroquímicos',
        cc.agrochemicals_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.agrochemicals_real_usd,
            'share_pct', CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- Fertilizers
      SELECT 
        'fertilizers',
        3,
        'pre_harvest',
        'Fertilizantes',
        cc.fertilizers_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.fertilizers_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.fertilizers_real_usd,
            'share_pct', CASE WHEN cc.fertilizers_total_usd > 0 THEN (id.fertilizers_real_usd / cc.fertilizers_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- Sowing
      SELECT 
        'sowing',
        4,
        'pre_harvest',
        'Siembra',
        cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.sowing_real_usd,
            'share_pct', CASE WHEN cc.sowing_total_usd > 0 THEN (id.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- General labors
      SELECT 
        'general_labors',
        5,
        'pre_harvest',
        'Labores Generales',
        cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.general_labors_real_usd,
            'share_pct', CASE WHEN cc.general_labors_total_usd > 0 THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- Irrigation
      SELECT 
        'irrigation',
        6,
        'pre_harvest',
        'Riego',
        cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.irrigation_real_usd,
            'share_pct', CASE WHEN cc.irrigation_total_usd > 0 THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- Administration and structure
      SELECT 
        'administration_structure',
        7,
        'pre_harvest',
        'Administración y Estructura',
        cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.administration_real_usd,
            'share_pct', CASE WHEN cc.administration_total_usd > 0 THEN (id.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL

      UNION ALL

      -- Capitalizable lease
      SELECT 
        'capitalizable_lease',
        8,
        'pre_harvest',
        'Arriendo Capitalizable',
        cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (
          SELECT jsonb_agg(jsonb_build_object(
            'investor_id', id.investor_id,
            'investor_name', id.investor_name,
            'amount_usd', id.rent_real_usd,
            'share_pct', CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END
          ) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
    ) cat_data
  ) AS contribution_categories,
  
  -- =========================================================================
  -- SECCIÓN 4: COMPARACIÓN ACORDADO vs REAL
  -- =========================================================================
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'agreed_share_pct', id.share_pct_agreed,
        'agreed_usd', CASE WHEN ct.total_contributions_usd > 0 THEN (ct.total_contributions_usd * id.share_pct_agreed / 100) ELSE 0 END,
        'actual_usd', id.real_contribution_usd,
        'adjustment_usd', id.real_contribution_usd - (CASE WHEN ct.total_contributions_usd > 0 THEN (ct.total_contributions_usd * id.share_pct_agreed / 100) ELSE 0 END)
      ) ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    CROSS JOIN contribution_totals ct
    WHERE id.project_id = pb.project_id AND ct.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  -- =========================================================================
  -- SECCIÓN 5: LIQUIDACIÓN DE COSECHA
  -- =========================================================================
  jsonb_build_object(
    'total_harvest_usd', 0,
    'total_harvest_usd_ha', 0,
    'investors', '[]'::jsonb
  ) AS harvest_settlement

FROM v3_report_investor_project_base pb
JOIN v3_report_investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN contribution_totals ct ON ct.project_id = pb.project_id

ORDER BY pb.project_id;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista consolidada para informe de Aportes por Inversor. VERSIÓN 171: Restaura estructura correcta de general_project_data con surface_total_ha, admin_total_usd, etc.';

COMMIT;

