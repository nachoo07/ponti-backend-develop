-- ========================================
-- MIGRACIÓN 000167: SEPARATE FERTILIZERS IN INVESTOR REPORT (UP)
-- ========================================
-- 
-- Propósito: Separar "Fertilizantes" de "Agroquímicos" en el informe de Aportes por Inversor
-- Fecha: 2025-10-29
-- Autor: Sistema
-- 
-- Problema identificado:
--   En el informe de aportes, "Agroquímicos" actualmente incluye todos los agroquímicos
--   pero Fertilizantes (type_id = 3) debe ser una categoría independiente
--
-- Corrección:
--   - Agroquímicos: Solo Coadyuvantes, Curasemillas, Herbicidas, Insecticidas, Fungicidas, Otros Insumos (type_id = 2)
--   - Fertilizantes: Nueva categoría independiente (type_id = 3)
--
-- Impacto:
--   - Solo afecta informe de Aportes por Inversor
--   - Los totales se mantienen iguales (la suma de ambas categorías == total anterior de Agroquímicos)
--   - Los controles de integridad deben seguir pasando
--
-- Note: Código en inglés, comentarios en español.

BEGIN;

-- ============================================================================
-- RECREAR VISTA 2: v3_report_investor_contribution_categories
-- ============================================================================
-- Agregar campo fertilizers_total_usd

DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;

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
  
  -- =========================================================================
  -- CATEGORÍA 1: AGROQUÍMICOS (INVERTIDOS - desde supply_movements)
  -- CORREGIDO: Solo agroquímicos SIN fertilizantes
  -- =========================================================================
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0
      AND cat.type_id = 2
      AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 1B: FERTILIZANTES (INVERTIDOS - desde supply_movements)
  -- NUEVO: Fertilizantes como categoría independiente
  -- =========================================================================
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0
      AND cat.type_id = 3
  ), 0)::numeric AS fertilizers_total_usd,
  
  -- CATEGORÍA 2: SEMILLA (INVERTIDOS - desde supply_movements)
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM public.supply_movements sm
    JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
    JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
    WHERE sm.project_id = lb.project_id
      AND sm.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
      AND sm.is_entry = TRUE
      AND s.price IS NOT NULL
      AND sm.quantity > 0
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0)::numeric AS seeds_total_usd,
  
  -- CATEGORÍA 3: LABORES GENERALES (INVERTIDOS - desde workorders)
  -- Incluye Pulverización, Otras Labores Y COSECHA
  COALESCE((
    SELECT SUM(lab.price * w.effective_area)
    FROM public.workorders w
    JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
    JOIN public.categories cat ON lab.category_id = cat.id
    WHERE w.lot_id IN (SELECT lot_id FROM lot_base WHERE project_id = lb.project_id)
      AND w.deleted_at IS NULL
      AND cat.type_id = 4
      AND cat.name IN ('Pulverización', 'Otras Labores', 'Cosecha')
  ), 0)::numeric AS general_labors_total_usd,
  
  -- CATEGORÍA 4: SIEMBRA (INVERTIDOS - desde workorders)
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
  
  -- CATEGORÍA 5: RIEGO (INVERTIDOS - desde workorders)
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
  
  -- CATEGORÍA 6: ARRIENDO CAPITALIZABLE
  COALESCE(SUM(
    v3_calc.rent_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- CATEGORÍA 7: ADMINISTRACIÓN Y ESTRUCTURA
  COALESCE(SUM(
    v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS administration_total_usd,
  
  -- SUPERFICIE TOTAL (para cálculos de USD/HA)
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb

GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
  'Vista 2/4 para informe de Aportes por Inversor. ACTUALIZADO 167: Separa Fertilizantes de Agroquímicos.';

-- ============================================================================
-- RECREAR VISTA 3: v3_report_investor_distributions
-- ============================================================================
-- Agregar campo fertilizers_real_usd

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
    fertilizers_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd + general_labors_total_usd + 
     sowing_total_usd + irrigation_total_usd + rent_capitalizable_total_usd + 
     administration_total_usd) AS total_contributions_usd
  FROM v3_report_investor_contribution_categories
),

-- Agroquímicos (SIN fertilizantes)
investor_agrochemicals_invested AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS agrochemicals_invested_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND sm.is_entry = TRUE
    AND s.price IS NOT NULL
    AND sm.quantity > 0
    AND cat.type_id = 2
    AND cat.name IN ('Coadyuvantes', 'Curasemillas', 'Herbicidas', 'Insecticidas', 'Fungicidas', 'Otros Insumos')
  GROUP BY sm.project_id, sm.investor_id
),

-- NUEVO: Fertilizantes independientes
investor_fertilizers_invested AS (
  SELECT
    sm.project_id,
    sm.investor_id,
    COALESCE(SUM(sm.quantity * s.price), 0)::numeric AS fertilizers_invested_usd
  FROM public.supply_movements sm
  JOIN public.supplies s ON s.id = sm.supply_id AND s.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = s.category_id AND cat.deleted_at IS NULL
  WHERE sm.deleted_at IS NULL
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND sm.is_entry = TRUE
    AND s.price IS NOT NULL
    AND sm.quantity > 0
    AND cat.type_id = 3
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
    AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada')
    AND sm.is_entry = TRUE
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

investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    
    COALESCE(agro.agrochemicals_invested_usd, 0) AS agrochemicals_real_usd,
    COALESCE(fert.fertilizers_invested_usd, 0) AS fertilizers_real_usd,
    COALESCE(seed.seeds_invested_usd, 0) AS seeds_real_usd,
    COALESCE(glabor.general_labors_real_usd, 0) AS general_labors_real_usd,
    COALESCE(sow.sowing_real_usd, 0) AS sowing_real_usd,
    COALESCE(irrig.irrigation_real_usd, 0) AS irrigation_real_usd,
    
    -- Placeholders para arriendo y administración (distribuidos proporcionalmente)
    (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100) AS rent_real_usd,
    (ct.administration_total_usd * ib.share_pct_agreed / 100) AS administration_real_usd,
    
    (
      COALESCE(agro.agrochemicals_invested_usd, 0) +
      COALESCE(fert.fertilizers_invested_usd, 0) +
      COALESCE(seed.seeds_invested_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0) +
      (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100) +
      (ct.administration_total_usd * ib.share_pct_agreed / 100)
    )::numeric AS total_real_contribution_usd,
    
    ct.total_contributions_usd AS project_total_contributions_usd
    
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN investor_agrochemicals_invested agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_fertilizers_invested fert ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
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
    
    (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric AS agreed_contribution_usd,
    irc.total_real_contribution_usd AS real_contribution_usd,
    
    (
      irc.total_real_contribution_usd - 
      (irc.project_total_contributions_usd * irc.share_pct_agreed / 100)
    )::numeric AS adjustment_usd,
    
    irc.agrochemicals_real_usd,
    irc.fertilizers_real_usd,
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
  fertilizers_real_usd,
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
  'Vista 3/4 para informe de Aportes por Inversor. ACTUALIZADO 167: Agrega fertilizers_real_usd.';

-- ============================================================================
-- RECREAR VISTA 4: v3_investor_contribution_data_view
-- ============================================================================
-- Agregar categoría "Fertilizantes" con sort_index = 2

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
    cc.fertilizers_total_usd,
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
      -- Semillas (sort_index = 1)
      SELECT 
        'seeds'::text AS key, 1 AS sort_index, 'pre_harvest'::text AS type, 'Semilla'::text AS label,
        cc.seeds_total_usd AS total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.seeds_real_usd,
          'share_pct', CASE WHEN cc.seeds_total_usd > 0 THEN (id2.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Agroquímicos (sort_index = 2)
      SELECT 
        'agrochemicals'::text, 2, 'pre_harvest'::text, 'Agroquímicos'::text,
        cc.agrochemicals_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
          'amount_usd', id2.agrochemicals_real_usd, 
          'share_pct', CASE WHEN cc.agrochemicals_total_usd > 0 THEN (id2.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- NUEVO: Fertilizantes (sort_index = 3)
      SELECT 
        'fertilizers'::text, 3, 'pre_harvest'::text, 'Fertilizantes'::text,
        cc.fertilizers_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.fertilizers_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
          'amount_usd', id2.fertilizers_real_usd, 
          'share_pct', CASE WHEN cc.fertilizers_total_usd > 0 THEN (id2.fertilizers_real_usd / cc.fertilizers_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Siembra Directa (sort_index = 4)
      SELECT 'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text, cc.sowing_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.sowing_real_usd,
          'share_pct', CASE WHEN cc.sowing_total_usd > 0 THEN (id2.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Labores Generales (sort_index = 5)
      SELECT 'general_labors'::text, 5, 'pre_harvest'::text, 'Labores Generales'::text, cc.general_labors_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.general_labors_real_usd,
          'share_pct', CASE WHEN cc.general_labors_total_usd > 0 THEN (id2.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Riego (sort_index = 6)
      SELECT 'irrigation'::text, 6, 'pre_harvest'::text, 'Riego'::text, cc.irrigation_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.irrigation_real_usd,
          'share_pct', CASE WHEN cc.irrigation_total_usd > 0 THEN (id2.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, NULL AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Administración y Estructura (sort_index = 7)
      SELECT 'administration_structure'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text, cc.administration_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.administration_real_usd,
          'share_pct', CASE WHEN cc.administration_total_usd > 0 THEN (id2.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, 'Distribuido proporcionalmente según aportes reales' AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
      UNION ALL
      -- Arriendos Capitalizables (sort_index = 8)
      SELECT 'capitalizable_lease'::text, 8, 'pre_harvest'::text, 'Arriendo Capitalizable'::text, cc.rent_capitalizable_total_usd,
        CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
        (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name,
          'amount_usd', id2.rent_real_usd,
          'share_pct', CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id2.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
          FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
        false AS requires_manual_attribution, 'Distribuido proporcionalmente según aportes reales' AS attribution_note
      FROM contribution_categories cc WHERE cc.project_id = pb.project_id
    ) cat_data
  ) AS contribution_categories,
  
  jsonb_build_object(
    'total_usd', (SELECT SUM(
      cc.agrochemicals_total_usd + cc.fertilizers_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
      cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
      cc.administration_total_usd
    ) FROM contribution_categories cc WHERE cc.project_id = pb.project_id),
    'total_usd_ha', CASE 
      WHEN pb.surface_total_ha > 0 
      THEN (SELECT SUM(
        cc.agrochemicals_total_usd + cc.fertilizers_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
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
  'Vista 4/4 para informe de Aportes por Inversor. ACTUALIZADO 167: Separa Fertilizantes de Agroquímicos.';

COMMIT;

