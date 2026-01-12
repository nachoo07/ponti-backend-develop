-- ========================================
-- MIGRATION 000136: FIX INVESTOR REPORT RENT CALCULATION (UP)
-- ========================================
-- 
-- Purpose: Corregir informe de Aportes por Inversor (4 vistas completas)
-- Date: 2025-10-12
-- Author: System
-- 
-- Cambios incluidos:
--   1. Nueva función SSOT: rent_fixed_only_for_lot() - solo considera arriendo fijo
--   2. Actualiza Vistas 1 y 2 para usar SSOT correcto
--   3. Recrea Vistas 3 y 4 con dependencias actualizadas
--   4. Agrega investor_headers y pre_harvest a Vista 4
--   5. harvest.rows con solo 1 fila (harvest) según especificación
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- FUNCIÓN SSOT: rent_fixed_only_for_lot
-- ============================================================================
-- Purpose: Devuelve SOLO el valor fijo del arriendo, si aplica
--
-- Lógica:
--   - lease_type_id = 1 (% sobre ingreso) → 0 (no es aporte)
--   - lease_type_id = 2 (% sobre resultado) → 0 (no es aporte)
--   - lease_type_id = 3 (Fijo) → lease_type_value
--   - lease_type_id = 4 (Fijo + %) → lease_type_value (solo la parte fija)
--
-- SSOT: Esta función está en v3_lot_ssot porque es específica para lotes
-- ============================================================================

CREATE OR REPLACE FUNCTION v3_lot_ssot.rent_fixed_only_for_lot(p_lot_id bigint) 
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT
    CASE
      -- Tipo 1: % sobre ingreso → NO se considera aporte
      WHEN f.lease_type_id = 1 THEN 0
      
      -- Tipo 2: % sobre resultado → NO se considera aporte
      WHEN f.lease_type_id = 2 THEN 0
      
      -- Tipo 3: Fijo → Se considera el valor fijo completo
      WHEN f.lease_type_id = 3 THEN COALESCE(f.lease_type_value, 0)
      
      -- Tipo 4: Fijo + % → Solo se considera la parte fija (NO el porcentaje)
      WHEN f.lease_type_id = 4 THEN COALESCE(f.lease_type_value, 0)
      
      -- Default: 0
      ELSE 0
    END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

COMMENT ON FUNCTION v3_lot_ssot.rent_fixed_only_for_lot(bigint) IS 
  'SSOT: Devuelve SOLO el valor fijo del arriendo por ha para un lote. Usado en informe de Aportes por Inversor. Para tipo Fijo+%, solo retorna la parte fija (NO el porcentaje).';

-- ============================================================================
-- RECREAR VISTA 1: v3_report_investor_project_base
-- ============================================================================

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_project_base CASCADE;

CREATE OR REPLACE VIEW public.v3_report_investor_project_base AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  
  COALESCE(SUM(l.hectares), 0)::numeric AS surface_total_ha,
  
  -- Usa rent_fixed_only_for_lot() - solo arriendo fijo
  COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0)::numeric AS lease_fixed_total_usd,
  
  CASE 
    WHEN COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0) > 0 
    THEN true 
    ELSE false 
  END AS lease_is_fixed,
  
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS lease_per_ha_usd,
  
  -- Usa SSOT v3_lot_ssot.admin_cost_per_ha_for_lot
  COALESCE(SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_total_usd,
  
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_lot_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS admin_per_ha_usd

FROM public.projects p
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.customer_id, c.name, p.campaign_id, cam.name;

COMMENT ON VIEW public.v3_report_investor_project_base IS 
  'Vista 1/4 para informe de Aportes por Inversor. Usa SSOT v3_lot_ssot.rent_fixed_only_for_lot() y v3_lot_ssot.admin_cost_per_ha_for_lot().';

-- ============================================================================
-- RECREAR VISTA 2: v3_report_investor_contribution_categories  
-- ============================================================================

CREATE OR REPLACE VIEW public.v3_report_investor_contribution_categories AS
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
)
SELECT
  lb.project_id,
  
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla')
  ), 0)::numeric AS seeds_total_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name IN ('Pulverización', 'Otras Labores')
  ), 0)::numeric AS general_labors_total_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.name = 'Siembra'
      AND cat.type_id = 4
  ), 0)::numeric AS sowing_total_usd,
  
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.name = 'Riego'
      AND cat.type_id = 4
  ), 0)::numeric AS irrigation_total_usd,
  
  -- Usa rent_fixed_only_for_lot() - solo arriendo fijo
  COALESCE(SUM(
    v3_lot_ssot.rent_fixed_only_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- Usa SSOT v3_lot_ssot.admin_cost_per_ha_for_lot
  COALESCE(SUM(
    v3_lot_ssot.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS administration_total_usd,
  
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb
GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
  'Vista 2/4 para informe de Aportes por Inversor. Usa SSOT v3_lot_ssot.rent_fixed_only_for_lot() y v3_lot_ssot.admin_cost_per_ha_for_lot().';

-- ============================================================================
-- RECREAR VISTA 3: v3_report_investor_distributions
-- ============================================================================

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
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    
    ROUND((ct.agrochemicals_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS agrochemicals_real_usd,
    ROUND((ct.seeds_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS seeds_real_usd,
    ROUND((ct.general_labors_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS general_labors_real_usd,
    ROUND((ct.sowing_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS sowing_real_usd,
    ROUND((ct.irrigation_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS irrigation_real_usd,
    ROUND((ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS rent_real_usd,
    ROUND((ct.administration_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS administration_real_usd,
    
    ROUND((
      (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100) +
      (ct.seeds_total_usd * ib.share_pct_agreed / 100) +
      (ct.general_labors_total_usd * ib.share_pct_agreed / 100) +
      (ct.sowing_total_usd * ib.share_pct_agreed / 100) +
      (ct.irrigation_total_usd * ib.share_pct_agreed / 100) +
      (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100) +
      (ct.administration_total_usd * ib.share_pct_agreed / 100)
    )::numeric, 2) AS total_real_contribution_usd,
    
    ct.total_contributions_usd AS project_total_contributions_usd
    
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
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
  'Vista 3/4 para informe de Aportes por Inversor. Calcula distribución y comparación acordado vs. real.';

-- ============================================================================
-- RECREAR VISTA 4: v3_investor_contribution_data_view (FINAL)
-- ============================================================================

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH harvest_data AS (
  SELECT
    f.project_id,
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
),
pre_harvest_totals AS (
  SELECT
    cc.project_id,
    (cc.agrochemicals_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
     cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
     cc.administration_total_usd) AS total_pre_harvest_usd,
    CASE 
      WHEN cc.total_seeded_area_ha > 0 
      THEN (cc.agrochemicals_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
            cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
            cc.administration_total_usd) / cc.total_seeded_area_ha
      ELSE 0
    END AS total_pre_harvest_usd_ha
  FROM v3_report_investor_contribution_categories cc
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
      SELECT 
        'agrochemicals'::text AS key, 1 AS sort_index, 'pre_harvest'::text AS type, 'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name, 
          'amount_usd', ROUND(id.agrochemicals_real_usd::numeric, 2), 
          'share_pct', ROUND(CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'seeds'::text, 2, 'pre_harvest'::text, 'Semilla'::text, cc.seeds_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.seeds_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.seeds_total_usd > 0 THEN (id.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'general_labors'::text, 3, 'pre_harvest'::text, 'Labores Generales'::text, cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.general_labors_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.general_labors_total_usd > 0 THEN (id.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text, cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.sowing_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.sowing_total_usd > 0 THEN (id.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'irrigation'::text, 5, 'pre_harvest'::text, 'Riego'::text, cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.irrigation_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.irrigation_total_usd > 0 THEN (id.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        false, NULL
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'rent_capitalizable'::text, 6, 'pre_harvest'::text, 'Arriendo Capitalizable'::text, cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.rent_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        true, 'Requiere atribución manual por inversor'
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      SELECT 'administration'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text, cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id.investor_id, 'investor_name', id.investor_name,
          'amount_usd', ROUND(id.administration_real_usd::numeric, 2),
          'share_pct', ROUND(CASE WHEN cc.administration_total_usd > 0 THEN (id.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END::numeric, 2)) ORDER BY id.investor_id)
          FROM v3_report_investor_distributions id WHERE id.project_id = pb.project_id),
        true, 'Requiere atribución manual por inversor'
      FROM v3_report_investor_contribution_categories cc WHERE cc.project_id = pb.project_id
    ) AS cat
  ) AS contribution_categories,
  
  jsonb_build_object(
    'total_usd', ROUND(pht.total_pre_harvest_usd::numeric, 2),
    'total_us_ha', ROUND(pht.total_pre_harvest_usd_ha::numeric, 2),
    'investors', (
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', id.investor_id,
          'investor_name', id.investor_name,
          'amount_usd', ROUND(id.real_contribution_usd::numeric, 2),
          'share_pct', ROUND(
            CASE 
              WHEN pht.total_pre_harvest_usd > 0 
              THEN (id.real_contribution_usd / pht.total_pre_harvest_usd * 100)
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
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', ROUND(COALESCE(hd.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(hd.total_harvest_usd_ha, 0)::numeric, 2),
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
          'amount_usd', ROUND(0::numeric, 2),
          'share_pct', ROUND(0::numeric, 2)
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
LEFT JOIN pre_harvest_totals pht ON pht.project_id = pb.project_id

ORDER BY pb.project_id;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista 4/4 (FINAL) para informe de Aportes por Inversor. Incluye investor_headers, pre_harvest y harvest.rows con 1 sola fila.';
