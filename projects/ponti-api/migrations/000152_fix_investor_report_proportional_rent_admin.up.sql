-- ========================================
-- MIGRATION 000152: FIX INVESTOR REPORT PROPORTIONAL RENT ADMIN (UP)
-- ========================================
-- 
-- Purpose: Distribuir Arriendos y Administración proporcionalmente según aportes reales
-- Date: 2025-10-14
-- Author: System
-- 
-- Problema identificado:
--   Arriendos y Administración usan valores fake de 100
--   Deben distribuirse proporcionalmente a los aportes reales de cada inversor en insumos y labores
--
-- Corrección:
--   - Calcular % real de cada inversor = suma de aportes reales / total aportes reales
--   - Distribuir Arriendos y Admin usando ese % real
--
-- Impacto:
--   - Solo afecta informe de Aportes por Inversor
--   - No afecta Dashboard ni otros reportes
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- RECREAR VISTA 3: v3_report_investor_distributions (CON DISTRIBUCIÓN PROPORCIONAL)
-- ============================================================================

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;

CREATE OR REPLACE VIEW public.v3_report_investor_distributions AS
WITH investor_base AS (
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
  SELECT
    project_id,
    agrochemicals_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    (agrochemicals_total_usd + seeds_total_usd + general_labors_total_usd + 
     sowing_total_usd + irrigation_total_usd + rent_capitalizable_total_usd + 
     administration_total_usd) AS total_contributions_usd
  FROM v3_report_investor_contribution_categories
),

investor_agrochemicals_invested AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS agrochemicals_invested_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.movement_type IN ('Stock', 'Remito oficial')
    AND s.price IS NOT NULL
    AND sm.quantity > 0
    AND cat.type_id = 2
    AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  GROUP BY sm.project_id, sm.investor_id
),

investor_seeds_invested AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS seeds_invested_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.movement_type IN ('Stock', 'Remito oficial')
    AND s.price IS NOT NULL
    AND sm.quantity > 0
    AND cat.type_id = 1
    AND cat.name = 'Semilla'
  GROUP BY sm.project_id, sm.investor_id
),

investor_general_labors_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lb.price * w.effective_area), 0)::numeric AS general_labors_real_usd
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND cat.type_id = 4
    AND cat.name IN ('Pulverización', 'Otras Labores', 'Cosecha')
  GROUP BY w.project_id, w.investor_id
),

investor_sowing_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lb.price * w.effective_area), 0)::numeric AS sowing_real_usd
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND cat.name = 'Siembra'
    AND cat.type_id = 4
  GROUP BY w.project_id, w.investor_id
),

investor_irrigation_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lb.price * w.effective_area), 0)::numeric AS irrigation_real_usd
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND cat.name = 'Riego'
    AND cat.type_id = 4
  GROUP BY w.project_id, w.investor_id
),

-- ============================================================================
-- NUEVO: Calcular aportes directos (sin Arriendos ni Admin) para calcular %
-- ============================================================================
investor_direct_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    (
      COALESCE(agro.agrochemicals_invested_usd, 0) +
      COALESCE(seed.seeds_invested_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0)
    )::numeric AS direct_contribution_usd
  FROM investor_base ib
  LEFT JOIN investor_agrochemicals_invested agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_invested seed ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors_real glabor ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
  LEFT JOIN investor_sowing_real sow ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation_real irrig ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
),

-- ============================================================================
-- NUEVO: Calcular total de aportes directos del proyecto (para calcular %)
-- ============================================================================
project_direct_totals AS (
  SELECT
    project_id,
    SUM(direct_contribution_usd) AS total_direct_contributions_usd
  FROM investor_direct_contributions
  GROUP BY project_id
),

-- ============================================================================
-- NUEVO: Calcular % real de cada inversor según sus aportes directos
-- ============================================================================
investor_real_pct AS (
  SELECT
    idc.project_id,
    idc.investor_id,
    idc.direct_contribution_usd,
    pdt.total_direct_contributions_usd,
    -- % real = aporte del inversor / total aportes directos
    CASE 
      WHEN pdt.total_direct_contributions_usd > 0 
      THEN (idc.direct_contribution_usd / pdt.total_direct_contributions_usd * 100)
      ELSE 0
    END::numeric AS real_contribution_pct
  FROM investor_direct_contributions idc
  JOIN project_direct_totals pdt ON pdt.project_id = idc.project_id
),

investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    
    ROUND(COALESCE(agro.agrochemicals_invested_usd, 0), 2) AS agrochemicals_real_usd,
    ROUND(COALESCE(seed.seeds_invested_usd, 0), 2) AS seeds_real_usd,
    ROUND(COALESCE(glabor.general_labors_real_usd, 0), 2) AS general_labors_real_usd,
    ROUND(COALESCE(sow.sowing_real_usd, 0), 2) AS sowing_real_usd,
    ROUND(COALESCE(irrig.irrigation_real_usd, 0), 2) AS irrigation_real_usd,
    
    -- ============================================================================
    -- CORREGIDO: Arriendo y Admin distribuidos proporcionalmente según aportes reales
    -- ============================================================================
    ROUND((ct.rent_capitalizable_total_usd * irp.real_contribution_pct / 100)::numeric, 2) AS rent_real_usd,
    ROUND((ct.administration_total_usd * irp.real_contribution_pct / 100)::numeric, 2) AS administration_real_usd,
    
    -- Total real = suma de todos los aportes (incluyendo Arriendos y Admin proporcionales)
    ROUND((
      COALESCE(agro.agrochemicals_invested_usd, 0) +
      COALESCE(seed.seeds_invested_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0) +
      (ct.rent_capitalizable_total_usd * irp.real_contribution_pct / 100) +
      (ct.administration_total_usd * irp.real_contribution_pct / 100)
    )::numeric, 2) AS total_real_contribution_usd,
    
    ct.total_contributions_usd AS project_total_contributions_usd
    
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  JOIN investor_real_pct irp ON irp.project_id = ib.project_id AND irp.investor_id = ib.investor_id
  LEFT JOIN investor_agrochemicals_invested agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_invested seed ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors_real glabor ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
  LEFT JOIN investor_sowing_real sow ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation_real irrig ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
),

investor_agreed_vs_real AS (
  SELECT
    irc.project_id,
    irc.investor_id,
    irc.investor_name,
    irc.share_pct_agreed,
    
    ROUND((irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric, 2) AS agreed_contribution_usd,
    irc.total_real_contribution_usd AS real_contribution_usd,
    
    ROUND((
      irc.total_real_contribution_usd - 
      (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)
    )::numeric, 2) AS adjustment_usd,
    
    irc.agrochemicals_real_usd,
    irc.seeds_real_usd,
    irc.general_labors_real_usd,
    irc.sowing_real_usd,
    irc.irrigation_real_usd,
    irc.rent_real_usd,
    irc.administration_real_usd,
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
  agrochemicals_real_usd,
  seeds_real_usd,
  general_labors_real_usd,
  sowing_real_usd,
  irrigation_real_usd,
  rent_real_usd,
  administration_real_usd,
  project_total_contributions_usd
FROM investor_agreed_vs_real
ORDER BY project_id, investor_id;

COMMENT ON VIEW public.v3_report_investor_distributions IS 
  'Vista 3/4 para informe de Aportes por Inversor. CORREGIDO 152: Arriendos y Admin distribuidos proporcionalmente según aportes reales de cada inversor en insumos y labores.';

-- ============================================================================
-- RECREAR VISTA 4: v3_investor_contribution_data_view
-- ============================================================================

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
    'surface_total_ha', ROUND(pb.surface_total_ha::numeric, 2),
    'lease_fixed_total_usd', ROUND(pb.lease_fixed_total_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2),
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2)
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
          'amount_usd', ROUND(id.agrochemicals_real_usd::numeric, 2), 
          'share_pct', ROUND(CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'seeds'::text, 2, 'pre_harvest'::text, 'Semilla'::text, cc.seeds_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.seeds_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.seeds_total_usd > 0 THEN (id.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'general_labors'::text, 3, 'pre_harvest'::text, 'Labores Generales'::text, cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.general_labors_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.general_labors_total_usd > 0 THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text, cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.sowing_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.sowing_total_usd > 0 THEN (id.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'irrigation'::text, 5, 'pre_harvest'::text, 'Riego'::text, cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.irrigation_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.irrigation_total_usd > 0 THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'rent_capitalizable'::text, 6, 'pre_harvest'::text, 'Arriendo Capitalizable'::text, cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.rent_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, 'Distribuido proporcionalmente según aportes reales'
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'administration'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text, cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.administration_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.administration_total_usd > 0 THEN (id.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
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
          'amount_usd', ROUND(id.real_contribution_usd::numeric, 2),
          'share_pct', ROUND(
            CASE 
              WHEN id.project_total_contributions_usd > 0 
              THEN (id.real_contribution_usd / id.project_total_contributions_usd * 100)
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
  
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'label', 'Cosecha',
        'total_usd', ROUND(hd.total_harvest_usd::numeric, 2),
        'total_usd_ha', CASE WHEN pb.surface_total_ha > 0 THEN ROUND((hd.total_harvest_usd / pb.surface_total_ha)::numeric, 2) ELSE 0 END,
        'investors', (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', id.investor_id,
              'investor_name', id.investor_name,
              'amount_usd', ROUND((COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100)::numeric, 2),
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
          'amount_usd', ROUND((COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100)::numeric, 2),
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
          'amount_usd', ROUND((
            0 - (COALESCE(hd.total_harvest_usd, 0) * id.share_pct_agreed / 100)
          )::numeric, 2)
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
  'Vista 4/4 para informe de Aportes por Inversor. CORREGIDO 152: Arriendos y Admin distribuidos proporcionalmente según aportes reales.';

