-- ========================================
-- MIGRATION 000191: ALIGN Investor Contribution Totals With Dashboard (UP)
-- ========================================
--
-- Purpose: Usar funciones SSOT de dashboard para Semillas y Agroquímicos
--          asegurando que el informe de aportes coincida con el dashboard (control 7).
--
-- Date: 2025-11-08
-- Author: System
--
-- Note: Code in English, comentarios en español.

BEGIN;

-- =========================================================================
-- Paso 1: Eliminar vistas dependientes en orden seguro
-- =========================================================================
DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress;
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view;
DROP VIEW IF EXISTS public.v3_report_investor_distributions;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories;

-- =========================================================================
-- Paso 2: Recrear v3_report_investor_contribution_categories usando SSOT dashboard
-- =========================================================================
CREATE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
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
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
seed_area AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha,
    COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS rent_capitalizable_total_usd,
    COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares), 0)::numeric AS administration_total_usd
  FROM lot_base lb
  GROUP BY lb.project_id
),
labor_totals AS (
  SELECT
    lb.project_id,
    COALESCE(SUM(CASE WHEN cat.name IN ('Pulverización', 'Otras Labores') THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS general_labors_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Siembra' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS sowing_total_usd,
    COALESCE(SUM(CASE WHEN cat.name = 'Riego' THEN lab.price * w.effective_area ELSE 0 END), 0)::numeric AS irrigation_total_usd
  FROM lot_base lb
  JOIN public.workorders w ON w.lot_id = lb.lot_id AND w.deleted_at IS NULL
  JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE cat.type_id = 4
  GROUP BY lb.project_id
),
invested_totals AS (
  SELECT
    p.project_id,
    v3_dashboard_ssot.seeds_invested_for_project_mb(p.project_id)::numeric AS seeds_total_usd,
    v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.project_id)::numeric AS agrochemicals_total_usd,
    (
      SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
      FROM public.supply_movements sm
      JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
      JOIN public.categories cat ON cat.id = s.category_id
      WHERE sm.project_id = p.project_id
        AND sm.deleted_at IS NULL
        AND sm.is_entry = TRUE
        AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
        AND cat.type_id = 3
    ) AS fertilizers_total_usd
  FROM (SELECT DISTINCT project_id FROM lot_base) p
)
SELECT
  it.project_id,
  it.agrochemicals_total_usd,
  COALESCE(it.fertilizers_total_usd, 0)::numeric AS fertilizers_total_usd,
  it.seeds_total_usd,
  COALESCE(lt.general_labors_total_usd, 0)::numeric AS general_labors_total_usd,
  COALESCE(lt.sowing_total_usd, 0)::numeric AS sowing_total_usd,
  COALESCE(lt.irrigation_total_usd, 0)::numeric AS irrigation_total_usd,
  COALESCE(sa.rent_capitalizable_total_usd, 0)::numeric AS rent_capitalizable_total_usd,
  COALESCE(sa.administration_total_usd, 0)::numeric AS administration_total_usd,
  COALESCE(sa.total_seeded_area_ha, 0)::numeric AS total_seeded_area_ha
FROM invested_totals it
LEFT JOIN labor_totals lt ON lt.project_id = it.project_id
LEFT JOIN seed_area sa ON sa.project_id = it.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
'Vista 2/4 para informe de aportes. FIX 000191: Semillas y Agroquímicos usan funciones SSOT de dashboard.';


CREATE VIEW public.v3_report_investor_distributions AS
WITH investor_base AS (
  -- Base de inversores por proyecto con sus % acordados
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.project_id IS NOT NULL
),
category_totals AS (
  -- Totales por categoría (de Vista 2)
  SELECT
    project_id,
    agrochemicals_total_usd,
    fertilizers_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    -- Total de todos los aportes
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd + general_labors_total_usd + 
     sowing_total_usd + irrigation_total_usd + rent_capitalizable_total_usd + 
     administration_total_usd) AS total_contributions_usd
  FROM v3_report_investor_contribution_categories
),
rent_real_by_investor AS (
  SELECT
    f.project_id,
    LOWER(inv.name) AS investor_key,
    SUM(
      v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares * COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.field_investors fi
  JOIN public.fields f ON f.id = fi.field_id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  JOIN public.investors inv ON inv.id = fi.investor_id AND inv.deleted_at IS NULL
  WHERE fi.deleted_at IS NULL
  GROUP BY f.project_id, LOWER(inv.name)
),
investor_real_contributions AS (
  -- Aportes REALES por inversor y categoría
  -- Basado en los montos realmente invertidos por cada inversor
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    
    -- =========================================================================
    -- APORTES REALES POR CATEGORÍA
    -- Para las categorías automáticas, se distribuyen según % acordado
    -- (Simplificación: en el futuro se podría usar tabla de aportes reales)
    -- =========================================================================
    
    -- Agroquímicos: Distribuido según % acordado
    ROUND((ct.agrochemicals_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS agrochemicals_real_usd,
    
    -- Fertilizantes: Distribuido según % acordado
    ROUND((ct.fertilizers_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS fertilizers_real_usd,
    
    -- Semilla: Distribuido según % acordado
    ROUND((ct.seeds_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS seeds_real_usd,
    
    -- Labores Generales: Distribuido según % acordado
    ROUND((ct.general_labors_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS general_labors_real_usd,
    
    -- Siembra: Distribuido según % acordado
    ROUND((ct.sowing_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS sowing_real_usd,
    
    -- Riego: Distribuido según % acordado
    ROUND((ct.irrigation_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS irrigation_real_usd,
    
    -- Arriendo Capitalizable: usar porcentajes reales de field_investors cuando existan
    ROUND(COALESCE(rri.rent_real_usd, 0)::numeric, 2) AS rent_real_usd,
    
    -- Administración: Requiere atribución manual (por ahora distribuido según %)
    ROUND((ct.administration_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS administration_real_usd,
    
    -- Total Real por Inversor (suma de todas las categorías)
    ROUND((
      (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100) +
      (ct.fertilizers_total_usd * ib.share_pct_agreed / 100) +
      (ct.seeds_total_usd * ib.share_pct_agreed / 100) +
      (ct.general_labors_total_usd * ib.share_pct_agreed / 100) +
      (ct.sowing_total_usd * ib.share_pct_agreed / 100) +
      (ct.irrigation_total_usd * ib.share_pct_agreed / 100) +
      COALESCE(rri.rent_real_usd, 0) +
      (ct.administration_total_usd * ib.share_pct_agreed / 100)
    )::numeric, 2) AS total_real_contribution_usd,
    
    -- Total de aportes del proyecto
    ct.total_contributions_usd AS project_total_contributions_usd
    
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN rent_real_by_investor rri
    ON rri.project_id = ib.project_id
   AND rri.investor_key = LOWER(ib.investor_name)
),
investor_agreed_vs_real AS (
  -- Comparación entre Aporte Acordado vs. Real
  SELECT
    irc.project_id,
    irc.investor_id,
    irc.investor_name,
    irc.share_pct_agreed,
    
    -- =========================================================================
    -- APORTE ACORDADO (Teórico)
    -- Qué representa: % de participación pactado entre socios
    -- Cómo se calcula: Total Aportes * % Acordado
    -- Ejemplo: 500 usd total, Inversor 1 (50%) = 250 usd
    -- =========================================================================
    ROUND((irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric, 2) AS agreed_contribution_usd,
    
    -- =========================================================================
    -- APORTE REAL
    -- Suma de todos los aportes reales del inversor en todas las categorías
    -- =========================================================================
    irc.total_real_contribution_usd AS real_contribution_usd,
    
    -- =========================================================================
    -- AJUSTE DE APORTE
    -- Qué representa: Diferencia entre aporte total realizado y aporte acordado
    -- Cómo se calcula: =APORTE REAL USD - APORTE ACORDADO
    -- Si es positivo: el inversor aportó más de lo acordado
    -- Si es negativo: el inversor aportó menos de lo acordado
    -- =========================================================================
    ROUND((
      irc.total_real_contribution_usd - 
      (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)
    )::numeric, 2) AS adjustment_usd,
    
    -- Aportes por categoría (para detalle)
    irc.agrochemicals_real_usd,
    irc.fertilizers_real_usd,
    irc.seeds_real_usd,
    irc.general_labors_real_usd,
    irc.sowing_real_usd,
    irc.irrigation_real_usd,
    irc.rent_real_usd,
    irc.administration_real_usd,
    
    -- Total del proyecto
    irc.project_total_contributions_usd
    
  FROM investor_real_contributions irc
)
SELECT
  project_id,
  investor_id,
  investor_name,
  share_pct_agreed,
  agreed_contribution_usd,
  real_contribution_usd,
  adjustment_usd,
  -- Detalle por categoría
  agrochemicals_real_usd,
  fertilizers_real_usd,
  seeds_real_usd,
  general_labors_real_usd,
  sowing_real_usd,
  irrigation_real_usd,
  rent_real_usd,
  administration_real_usd,
  -- Total del proyecto
  project_total_contributions_usd
FROM investor_agreed_vs_real
ORDER BY project_id, investor_id;

-- Comentario de la vista
COMMENT ON VIEW public.v3_report_investor_distributions IS 
  'Vista 3/4 para informe de Aportes por Inversor. FIX 000189: Incluye fertilizers_real_usd.';

-- ============================================================================
-- VISTA 4 de 4: v3_investor_contribution_data_view (FINAL)
-- ============================================================================
-- Purpose: Vista Final Consolidada con Estructura JSON para el Informe de Aportes
--
-- Integra las 3 vistas anteriores:
--   - Vista 1: Datos Generales del Proyecto
--   - Vista 2: Categorías de Aportes
--   - Vista 3: Distribución por Inversor
--
-- Incluye:
--   - Sección 4: Liquidación de Cosecha
--
-- Esta vista genera la estructura JSON completa esperada por el frontend
-- ============================================================================

CREATE VIEW public.v3_investor_contribution_data_view AS
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
    -- Total de Cosecha
    'total_harvest_usd', ROUND(hd.total_harvest_usd::numeric, 2),
    'total_harvest_usd_ha', ROUND(hd.total_harvest_usd_ha::numeric, 2),
    
    -- Distribución por Inversor
    'investors', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          
          -- COSECHA PAGADO (distribuido según % acordado - simplificación actual)
          'harvest_paid_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100)::numeric, 2),
          
          -- PAGO ACORDADO (según % en Clientes y Sociedades)
          'harvest_agreed_usd', ROUND((hd.total_harvest_usd * id.share_pct_agreed / 100)::numeric, 2),
          
          -- USD AJUSTE DE PAGO = (COSECHA * % PAGO) - COSECHA PAGADO
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

-- Comentario de la vista
COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista 4/4 (FINAL) para informe de Aportes por Inversor. Consolida las 3 vistas anteriores y genera la estructura JSON completa para la API. Incluye: Datos Generales, Categorías de Aportes, Comparación Acordado vs. Real, y Liquidación de Cosecha. Recreada en migración 146 para asegurar consistencia entre local y GCP.';

-- ============================================================================
-- RECREAR VISTA DE DASHBOARD QUE SE BORRÓ CON CASCADE
-- ============================================================================
-- v3_dashboard_contributions_progress depende de v3_report_investor_contribution_categories
-- Por eso el CASCADE la borró - necesitamos recrearla

CREATE VIEW public.v3_dashboard_contributions_progress AS
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

COMMENT ON VIEW public.v3_dashboard_contributions_progress IS 
'Avance de aportes de inversores. FIX 000187: Muestra % real (arriba) vs % teórico acordado (abajo). Recreada en FIX 000188.';

COMMIT;
