-- ========================================
-- MIGRATION 000193: Fix harvest section structure in investor report (UP)
-- ========================================
--
-- Objetivo: Exponer la liquidación de cosecha con la estructura esperada
--           ({rows, footer_payment_agreed, footer_payment_adjustment}) para
--           que el frontend muestre los totales y porcentajes por inversor.
--
-- Nota: Código en inglés, comentarios en español.

BEGIN;

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH harvest_data AS (
  -- =========================================================================
  -- SECCIÓN 4: LIQUIDACIÓN DE COSECHA
  -- =========================================================================
  SELECT
    f.project_id,
    
    -- COSECHA: Montos pagados por cada inversor en Cosecha
    -- Los Montos pagados por cada inversor en Cosecha no se consideran aportes
    -- Invertidos, sino que eso se calcula en función de aporte acordado
    
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
    
    -- Total de Cosecha por hectárea (usando superficie sembrada)
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
    
  FROM public.projects p
  JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN v3_report_investor_contribution_categories cc ON cc.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY f.project_id, cc.total_seeded_area_ha
)
SELECT
  -- Información del Proyecto (de Vista 1)
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  -- =========================================================================
  -- INVESTOR HEADERS: Lista de inversores con sus porcentajes acordados
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
    'lease_fixed_total_usd', ROUND(pb.lease_fixed_total_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2),
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2)
  ) AS general_project_data,
  
  -- =========================================================================
  -- SECCIÓN 2: APORTES POR INVERSOR (Categorías de Aportes)
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
      
      -- Fertilizantes
      SELECT 
        'fertilizers'::text,
        2,
        'pre_harvest'::text,
        'Fertilizantes'::text,
        cc.fertilizers_total_usd,
        CASE 
          WHEN cc.total_seeded_area_ha > 0 
          THEN cc.fertilizers_total_usd / cc.total_seeded_area_ha 
          ELSE 0 
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND(id.fertilizers_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN cc.fertilizers_total_usd > 0 
                  THEN (id.fertilizers_real_usd / cc.fertilizers_total_usd * 100)
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
      
      -- Semilla
      SELECT 
        'seeds'::text,
        3,
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
        4,
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
        5,
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
        6,
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
        'rent_capitalizable'::text,
        7,
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
        'administration'::text,
        8,
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
  -- SECCIÓN 3: COMPARACIÓN APORTE TEÓRICO VS. REAL
  -- =========================================================================
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'agreed_share_pct', id.share_pct_agreed,
        'agreed_usd', ROUND(id.agreed_contribution_usd::numeric, 2),
        'actual_usd', ROUND(id.real_contribution_usd::numeric, 2),
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
    'rows', jsonb_build_array(
      jsonb_build_object(
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', ROUND(COALESCE(hd.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(hd.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND((COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(hd.total_harvest_usd, 0)::numeric > 0 
                  THEN ((COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100) / COALESCE(hd.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ), '[]'::jsonb)
      ),
      jsonb_build_object(
        'key', 'totals',
        'type', 'totals',
        'total_usd', ROUND(COALESCE(hd.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(hd.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND((COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(hd.total_harvest_usd, 0)::numeric > 0 
                  THEN ((COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100) / COALESCE(hd.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY id.investor_id
          )
          FROM v3_report_investor_distributions id
          WHERE id.project_id = pb.project_id
        ), '[]'::jsonb)
      )
    ),
    'footer_payment_agreed', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND((COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100)::numeric, 2),
          'share_pct', ROUND(id.share_pct_agreed::numeric, 2)
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    ), '[]'::jsonb),
    'footer_payment_adjustment', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND((
            (COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100) - 
            (COALESCE(hd.total_harvest_usd, 0)::numeric * id.share_pct_agreed::numeric / 100)
          )::numeric, 2)
        )
        ORDER BY id.investor_id
      )
      FROM v3_report_investor_distributions id
      WHERE id.project_id = pb.project_id
    ), '[]'::jsonb)
  ) AS harvest_settlement

FROM v3_report_investor_project_base pb
JOIN v3_report_investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_data hd ON hd.project_id = pb.project_id

ORDER BY pb.project_id;

-- Comentario de la vista
COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista 4/4 (FINAL) para informe de Aportes por Inversor. Consolida las 3 vistas anteriores y genera la estructura JSON completa para la API. Incluye: Datos Generales, Categorías de Aportes, Comparación Acordado vs. Real, y Liquidación de Cosecha. Recreada en migración 146 para asegurar consistencia entre local y GCP.';

COMMIT;

