-- ========================================
-- MIGRACIÓN 000147: FIX v3_investor_contribution_data_view - Add investor_headers (UP)
-- ========================================
-- 
-- Propósito: Agregar columna investor_headers a v3_investor_contribution_data_view
-- Problema: La vista actual no incluye investor_headers, causando error en repository.go:551
-- Solución: Recrear la vista incluyendo la columna investor_headers con los datos de inversores
-- Fecha: 2025-10-14
-- Autor: Sistema
-- 
-- Nota: Código en inglés, comentarios en español
-- Basado en migración 000146 con agregado de investor_headers

BEGIN;

-- ========================================
-- RECREAR v3_investor_contribution_data_view CON INVESTOR_HEADERS
-- ========================================
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;

CREATE VIEW public.v3_investor_contribution_data_view AS
WITH harvest_data AS (
  -- =========================================================================
  -- SECCIÓN 4: LIQUIDACIÓN DE COSECHA
  -- =========================================================================
  SELECT
    f.project_id,
    
    -- Total de Cosecha (todos los inversores)
    COALESCE((
      SELECT SUM(lab.price * w.effective_area)
      FROM public.workorders w
      JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
      JOIN public.categories cat ON lab.category_id = cat.id
      JOIN public.lots l ON w.lot_id = l.id AND l.deleted_at IS NULL
      JOIN public.fields fld ON l.field_id = fld.id AND fld.deleted_at IS NULL
      WHERE fld.project_id = f.project_id
        AND w.deleted_at IS NULL
        AND cat.name = 'Cosecha'
        AND cat.type_id = 4
    ), 0)::numeric AS total_harvest_usd,
    
    -- Total de Cosecha por hectárea
    CASE 
      WHEN cc.total_seeded_area_ha > 0 
      THEN COALESCE((
        SELECT SUM(lab.price * w.effective_area)
        FROM public.workorders w
        JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
        JOIN public.categories cat ON lab.category_id = cat.id
        JOIN public.lots l ON w.lot_id = l.id AND l.deleted_at IS NULL
        JOIN public.fields fld ON l.field_id = fld.id AND fld.deleted_at IS NULL
        WHERE fld.project_id = f.project_id
          AND w.deleted_at IS NULL
          AND cat.name = 'Cosecha'
          AND cat.type_id = 4
      ), 0) / cc.total_seeded_area_ha
      ELSE 0
    END::numeric AS total_harvest_usd_ha
    
  FROM v3_report_investor_contribution_categories cc
  JOIN public.projects p ON p.id = cc.project_id AND p.deleted_at IS NULL
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
)

SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  -- =========================================================================
  -- INVESTOR HEADERS: Array de inversores con % acordado (★ AGREGADO EN 000147)
  -- =========================================================================
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct', ROUND(id.share_pct_agreed::numeric, 2)
      )
      ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_headers,
  
  -- =========================================================================
  -- SECCIÓN 1: DATOS GENERALES DEL PROYECTO
  -- =========================================================================
  jsonb_build_object(
    'surface_total_ha', ROUND(pb.surface_total_ha::numeric, 2),
    'lease_fixed_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2)
  ) AS general_project_data,
  
  -- =========================================================================
  -- SECCIÓN 2: PRE-COSECHA - CATEGORÍAS DE APORTES
  -- =========================================================================
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'key', cat.key,
        'sort_index', cat.sort_index,
        'type', cat.type,
        'label', cat.label,
        'total_usd', ROUND(cat.total_usd::numeric, 2),
        'total_usd_ha', ROUND(cat.total_usd_ha::numeric, 2),
        'investors', cat.investors,
        'requires_manual_attribution', cat.requires_manual_attribution,
        'attribution_note', cat.attribution_note
      )
      ORDER BY cat.sort_index
    )
    FROM (
      -- Agroquímicos
      SELECT 
        'agrochemicals'::text AS key,
        1 AS sort_index,
        'pre_harvest'::text AS type,
        'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END AS total_usd_ha,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.agrochemicals_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.agrochemicals_total_usd > 0 
                  THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution,
        NULL AS attribution_note
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Semilla
      SELECT 
        'seeds'::text,
        2,
        'pre_harvest'::text,
        'Semilla'::text,
        cc.seeds_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.seeds_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.seeds_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.seeds_total_usd > 0 
                  THEN (id.seeds_real_usd / cc.seeds_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Labores Generales
      SELECT 
        'general_labors'::text,
        3,
        'pre_harvest'::text,
        'Labores Generales'::text,
        cc.general_labors_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.general_labors_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.general_labors_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.general_labors_total_usd > 0 
                  THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Siembra
      SELECT 
        'sowing'::text,
        4,
        'pre_harvest'::text,
        'Siembra'::text,
        cc.sowing_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.sowing_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.sowing_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.sowing_total_usd > 0 
                  THEN (id.sowing_real_usd / cc.sowing_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Riego
      SELECT 
        'irrigation'::text,
        5,
        'pre_harvest'::text,
        'Riego'::text,
        cc.irrigation_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.irrigation_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.irrigation_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.irrigation_total_usd > 0 
                  THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Arriendo Capitalizable
      SELECT 
        'capitalizable_lease'::text,
        6,
        'pre_harvest'::text,
        'Arriendo Capitalizable'::text,
        cc.rent_capitalizable_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.rent_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.rent_capitalizable_total_usd > 0 
                  THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
      
      UNION ALL
      
      -- Administración y Estructura
      SELECT 
        'administration_structure'::text,
        7,
        'pre_harvest'::text,
        'Administración y Estructura'::text,
        cc.administration_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.administration_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.administration_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.administration_total_usd > 0 
                  THEN (id.administration_real_usd / cc.administration_total_usd * 100)
                  ELSE 0 
                END::numeric, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v3_report_investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
    ) AS cat
  ) AS contribution_categories,
  
  -- =========================================================================
  -- SECCIÓN 3: APORTE ACORDADO vs REAL (desde vista 3)
  -- =========================================================================
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct_agreed', id.share_pct_agreed,
        'agreed_contribution_usd', ROUND(id.agreed_contribution_usd::numeric, 2),
        'real_contribution_usd', ROUND(id.real_contribution_usd::numeric, 2),
        'adjustment_usd', ROUND(id.adjustment_usd::numeric, 2)
      )
      ORDER BY id.investor_id
    )
    FROM v3_report_investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  -- =========================================================================
  -- SECCIÓN 4: LIQUIDACIÓN DE COSECHA
  -- =========================================================================
  jsonb_build_object(
    'total_harvest_usd', ROUND(hd.total_harvest_usd::numeric, 2),
    'total_harvest_usd_ha', ROUND(hd.total_harvest_usd_ha::numeric, 2),
    'investors', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'harvest_paid_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100)::numeric, 2),
          'harvest_agreed_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100)::numeric, 2),
          'harvest_adjustment_usd', ROUND((
            (hd.total_harvest_usd * id.share_pct_agreed / 100) - 
            (hd.total_harvest_usd * id.share_pct_agreed / 100)
          )::numeric, 2)
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    )
  ) AS harvest_settlement

FROM v3_report_investor_project_base pb
JOIN v3_report_investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id

ORDER BY pb.project_id;

COMMIT;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista FINAL para informe de Aportes por Inversor. Incluye investor_headers, datos generales, categorías de aportes, comparación acordado vs real, y liquidación de cosecha. Corregida en migración 147 para incluir columna investor_headers faltante.';

