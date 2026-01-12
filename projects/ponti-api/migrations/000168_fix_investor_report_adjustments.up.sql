-- ========================================
-- MIGRACIÓN 000168: FIX INVESTOR REPORT ADJUSTMENTS (UP)
-- ========================================
-- 
-- Propósito: Ajustar el informe de Aportes por Inversor según requerimientos del usuario
-- Fecha: 2025-10-29
-- Autor: Sistema
-- 
-- Ajustes implementados:
--   1. Excluir "Cosecha" del cálculo de "Labores Generales" (solo pre-cosecha)
--   2. Arriendo: Solo considerarlo aporte si el lease_type_id es 3 (ARRIENDO FIJO) o 4 (ARRIENDO FIJO + % INGRESO NETO)
--
-- Note: Los ajustes de formateo (decimales, USD/Ha) se manejan en el frontend.
--       La distribución de Administración por inversor se maneja automáticamente si no hay datos en admin_cost_investors.
--

BEGIN;

-- ============================================================================
-- RECREAR VISTA 2: v3_report_investor_contribution_categories
-- ============================================================================
-- Ajustar Labores Generales para excluir Cosecha
-- Ajustar Arriendo para solo considerar arriendos fijos

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
      AND cat.name = 'Fertilizantes'
  ), 0)::numeric AS fertilizers_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 2: SEMILLAS (INVERTIDOS - desde supply_movements)
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
      AND cat.type_id = 1
      AND cat.name = 'Semilla'
  ), 0)::numeric AS seeds_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 3: LABORES GENERALES (INVERTIDOS - desde workorders)
  -- AJUSTE 000168: EXCLUIR "Cosecha" - Solo Pulverización y Otras Labores
  -- =========================================================================
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
  
  -- =========================================================================
  -- CATEGORÍA 6: ARRIENDO CAPITALIZABLE
  -- AJUSTE 000168: Solo si lease_type_id IN (3, 4) - ARRIENDO FIJO o ARRIENDO FIJO + %
  -- =========================================================================
  COALESCE(SUM(
    CASE 
      WHEN f.lease_type_id IN (3, 4) THEN v3_calc.rent_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
      ELSE 0
    END
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- CATEGORÍA 7: ADMINISTRACIÓN Y ESTRUCTURA
  COALESCE(SUM(
    v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS administration_total_usd,
  
  -- SUPERFICIE TOTAL (para cálculos de USD/HA)
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb
JOIN public.fields f ON f.id = (SELECT field_id FROM public.lots WHERE id = lb.lot_id)

GROUP BY lb.project_id;

COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
  'Vista 2/4 para informe de Aportes por Inversor. ACTUALIZADO 168: Excluye Cosecha de Labores Generales, Arriendo solo si es fijo.';

-- ============================================================================
-- RECREAR VISTA 3: v3_report_investor_distributions
-- ============================================================================
-- Distribuir las categorías entre inversores

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
  FROM public.v3_report_investor_contribution_categories
),
-- Distribución real de Agroquímicos (SIN fertilizantes) por inversor
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
-- Distribución real de Fertilizantes por inversor
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
-- Distribución real de Semillas por inversor
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
  GROUP BY sm.project_id, sm.investor_id
),
-- Distribución real de Labores (Siembra, Generales, Riego) por inversor
investor_labors_invested AS (
  SELECT
    w.project_id,
    w.investor_id,
    cat.name AS labor_category,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS labor_invested_usd
  FROM public.workorders w
  JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
  GROUP BY w.project_id, w.investor_id, cat.name
),
investor_sowing AS (
  SELECT project_id, investor_id, SUM(labor_invested_usd) AS sowing_invested_usd
  FROM investor_labors_invested
  WHERE labor_category = 'Siembra'
  GROUP BY project_id, investor_id
),
investor_general_labors AS (
  SELECT project_id, investor_id, SUM(labor_invested_usd) AS general_labors_invested_usd
  FROM investor_labors_invested
  WHERE labor_category IN ('Pulverización', 'Otras Labores')
  GROUP BY project_id, investor_id
),
investor_irrigation AS (
  SELECT project_id, investor_id, SUM(labor_invested_usd) AS irrigation_invested_usd
  FROM investor_labors_invested
  WHERE labor_category = 'Riego'
  GROUP BY project_id, investor_id
),
-- Aportes reales consolidados por inversor
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    COALESCE(agro.agrochemicals_invested_usd, 0) AS agrochemicals_real_usd,
    COALESCE(ferti.fertilizers_invested_usd, 0) AS fertilizers_real_usd,
    COALESCE(seed.seeds_invested_usd, 0) AS seeds_real_usd,
    COALESCE(gen_lab.general_labors_invested_usd, 0) AS general_labors_real_usd,
    COALESCE(sow.sowing_invested_usd, 0) AS sowing_real_usd,
    COALESCE(irrig.irrigation_invested_usd, 0) AS irrigation_real_usd,
    COALESCE(agro.agrochemicals_invested_usd, 0) + 
    COALESCE(ferti.fertilizers_invested_usd, 0) +
    COALESCE(seed.seeds_invested_usd, 0) + 
    COALESCE(gen_lab.general_labors_invested_usd, 0) + 
    COALESCE(sow.sowing_invested_usd, 0) + 
    COALESCE(irrig.irrigation_invested_usd, 0) AS real_contribution_usd
  FROM investor_base ib
  LEFT JOIN investor_agrochemicals_invested agro ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_fertilizers_invested ferti ON ferti.project_id = ib.project_id AND ferti.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_invested seed ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors gen_lab ON gen_lab.project_id = ib.project_id AND gen_lab.investor_id = ib.investor_id
  LEFT JOIN investor_sowing sow ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation irrig ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
),
-- Calcular totales reales por proyecto
project_real_totals AS (
  SELECT
    project_id,
    SUM(real_contribution_usd) AS total_real_contribution_usd
  FROM investor_real_contributions
  GROUP BY project_id
),
-- Distribución proporcional de Arriendo y Administración según aportes reales
investor_agreed_vs_real AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    (ib.share_pct_agreed::numeric / 100.0) * ct.total_contributions_usd AS agreed_contribution_usd,
    irc.agrochemicals_real_usd,
    irc.fertilizers_real_usd,
    irc.seeds_real_usd,
    irc.general_labors_real_usd,
    irc.sowing_real_usd,
    irc.irrigation_real_usd,
    irc.real_contribution_usd,
    -- Arriendo proporcional según aporte real
    CASE 
      WHEN prt.total_real_contribution_usd > 0 THEN
        (irc.real_contribution_usd / prt.total_real_contribution_usd) * ct.rent_capitalizable_total_usd
      ELSE 
        (ib.share_pct_agreed::numeric / 100.0) * ct.rent_capitalizable_total_usd
    END AS rent_real_usd,
    -- Administración: Verificar si existe distribución manual, sino usar proporcional
    COALESCE(
      (SELECT (aci.percentage::numeric / 100.0) * ct.administration_total_usd 
       FROM public.admin_cost_investors aci 
       WHERE aci.project_id = ib.project_id AND aci.investor_id = ib.investor_id AND aci.deleted_at IS NULL),
      -- Si no hay distribución manual, usar proporcional según aportes reales
      CASE 
        WHEN prt.total_real_contribution_usd > 0 THEN
          (irc.real_contribution_usd / prt.total_real_contribution_usd) * ct.administration_total_usd
        ELSE 
          (ib.share_pct_agreed::numeric / 100.0) * ct.administration_total_usd
      END
    ) AS administration_real_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN investor_real_contributions irc ON irc.project_id = ib.project_id AND irc.investor_id = ib.investor_id
  LEFT JOIN project_real_totals prt ON prt.project_id = ib.project_id
)
SELECT
  project_id,
  investor_id,
  investor_name,
  share_pct_agreed,
  agreed_contribution_usd,
  (agrochemicals_real_usd + fertilizers_real_usd + seeds_real_usd + general_labors_real_usd + 
   sowing_real_usd + irrigation_real_usd + rent_real_usd + administration_real_usd) AS real_contribution_usd,
  (agrochemicals_real_usd + fertilizers_real_usd + seeds_real_usd + general_labors_real_usd + 
   sowing_real_usd + irrigation_real_usd + rent_real_usd + administration_real_usd - agreed_contribution_usd) AS adjustment_usd,
  agrochemicals_real_usd,
  fertilizers_real_usd,
  seeds_real_usd,
  general_labors_real_usd,
  sowing_real_usd,
  irrigation_real_usd,
  rent_real_usd,
  administration_real_usd,
  project_total_contributions_usd
FROM investor_agreed_vs_real;

COMMENT ON VIEW public.v3_report_investor_distributions IS 
  'Vista 3/4 para informe de Aportes por Inversor. ACTUALIZADO 168: Distribuye categorías ajustadas entre inversores.';

-- ============================================================================
-- RECREAR VISTA 4: v3_investor_contribution_data_view (Vista final consolidada)
-- ============================================================================

CREATE OR REPLACE VIEW public.v3_investor_contribution_data_view AS
WITH project_base AS (
  SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.customer_id,
    c.name AS customer_name,
    ca.id AS campaign_id,
    ca.name AS campaign_name
  FROM public.projects p
  JOIN public.customers c ON c.id = p.customer_id AND c.deleted_at IS NULL
  JOIN public.campaigns ca ON ca.id = p.campaign_id AND ca.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
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
    cc.total_seeded_area_ha,
    (cc.agrochemicals_total_usd + cc.fertilizers_total_usd + cc.seeds_total_usd + cc.general_labors_total_usd + 
     cc.sowing_total_usd + cc.irrigation_total_usd + cc.rent_capitalizable_total_usd + 
     cc.administration_total_usd) AS total_contributions_usd
  FROM public.v3_report_investor_contribution_categories cc
)
SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  
  -- =========================================================================
  -- SECCIÓN 1: HEADERS DE INVERSORES
  -- =========================================================================
  (SELECT jsonb_agg(jsonb_build_object(
      'investor_id', id.investor_id,
      'investor_name', id.investor_name,
      'share_pct_agreed', id.share_pct_agreed
    ) ORDER BY id.investor_id)
   FROM v3_report_investor_distributions id
   WHERE id.project_id = pb.project_id
  ) AS investor_headers,
  
  -- =========================================================================
  -- SECCIÓN 2: DATOS GENERALES DEL PROYECTO
  -- =========================================================================
  jsonb_build_object(
    'general', jsonb_build_object(
      'total_usd', cc.total_contributions_usd,
      'total_usd_ha', CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.total_contributions_usd / cc.total_seeded_area_ha ELSE 0 END
    ),
    'pre_harvest', jsonb_build_object(
      'total_usd', cc.total_contributions_usd,
      'total_usd_ha', CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.total_contributions_usd / cc.total_seeded_area_ha ELSE 0 END,
      'investors', (
        SELECT jsonb_agg(jsonb_build_object(
          'investor_id', id2.investor_id,
          'investor_name', id2.investor_name,
          'amount_usd', id2.real_contribution_usd,
          'share_pct', CASE WHEN cc.total_contributions_usd > 0 THEN (id2.real_contribution_usd / cc.total_contributions_usd * 100) ELSE 0 END
        ) ORDER BY id2.investor_id)
        FROM v3_report_investor_distributions id2
        WHERE id2.project_id = pb.project_id
      )
    ),
    'harvest', jsonb_build_object(
      'total_usd', 0,
      'total_usd_ha', 0,
      'investors', '[]'::jsonb
    )
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
      )
      ORDER BY cat_data.sort_index
    )
    FROM (
     -- Semilla (sort_index = 1)
     SELECT 
       'seeds'::text AS key, 1 AS sort_index, 'pre_harvest'::text AS type, 'Semilla'::text AS label,
       cc.seeds_total_usd AS total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.seeds_total_usd / cc.total_seeded_area_ha ELSE 0 END AS total_usd_ha,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.seeds_real_usd, 
         'share_pct', CASE WHEN cc.seeds_total_usd > 0 THEN (id2.seeds_real_usd / cc.seeds_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, NULL::text AS attribution_note
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
       false AS requires_manual_attribution, NULL::text AS attribution_note
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
       false AS requires_manual_attribution, NULL::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
     
     UNION ALL
     
     -- Siembra (sort_index = 4)
     SELECT 
       'sowing'::text, 4, 'pre_harvest'::text, 'Siembra'::text,
       cc.sowing_total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.sowing_total_usd / cc.total_seeded_area_ha ELSE 0 END,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.sowing_real_usd, 
         'share_pct', CASE WHEN cc.sowing_total_usd > 0 THEN (id2.sowing_real_usd / cc.sowing_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, NULL::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
     
     UNION ALL
     
     -- Labores Generales (sort_index = 5)
     SELECT 
       'general_labors'::text, 5, 'pre_harvest'::text, 'Labores Generales'::text,
       cc.general_labors_total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.general_labors_total_usd / cc.total_seeded_area_ha ELSE 0 END,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.general_labors_real_usd, 
         'share_pct', CASE WHEN cc.general_labors_total_usd > 0 THEN (id2.general_labors_real_usd / cc.general_labors_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, NULL::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
     
     UNION ALL
     
     -- Riego (sort_index = 6)
     SELECT 
       'irrigation'::text, 6, 'pre_harvest'::text, 'Riego'::text,
       cc.irrigation_total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.irrigation_total_usd / cc.total_seeded_area_ha ELSE 0 END,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.irrigation_real_usd, 
         'share_pct', CASE WHEN cc.irrigation_total_usd > 0 THEN (id2.irrigation_real_usd / cc.irrigation_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, NULL::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
     
     UNION ALL
     
     -- Administración y Estructura (sort_index = 7)
     SELECT 
       'administration_structure'::text, 7, 'pre_harvest'::text, 'Administración y Estructura'::text,
       cc.administration_total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.administration_total_usd / cc.total_seeded_area_ha ELSE 0 END,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.administration_real_usd, 
         'share_pct', CASE WHEN cc.administration_total_usd > 0 THEN (id2.administration_real_usd / cc.administration_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, 
       'Distribuido proporcionalmente según aportes reales'::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
     
     UNION ALL
     
     -- Arriendo Capitalizable (sort_index = 8)
     SELECT 
       'capitalizable_lease'::text, 8, 'pre_harvest'::text, 'Arriendo Capitalizable'::text,
       cc.rent_capitalizable_total_usd,
       CASE WHEN cc.total_seeded_area_ha > 0 THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha ELSE 0 END,
       (SELECT jsonb_agg(jsonb_build_object('investor_id', id2.investor_id, 'investor_name', id2.investor_name, 
         'amount_usd', id2.rent_real_usd, 
         'share_pct', CASE WHEN cc.rent_capitalizable_total_usd > 0 THEN (id2.rent_real_usd / cc.rent_capitalizable_total_usd * 100) ELSE 0 END) ORDER BY id2.investor_id)
         FROM v3_report_investor_distributions id2 WHERE id2.project_id = pb.project_id) AS investors,
       false AS requires_manual_attribution, 
       'Distribuido proporcionalmente según aportes reales'::text AS attribution_note
     FROM contribution_categories cc WHERE cc.project_id = pb.project_id
   ) AS cat_data
  ) AS contribution_categories,
  
  -- =========================================================================
  -- SECCIÓN 4: COMPARACIÓN APORTE ACORDADO VS REAL
  -- =========================================================================
  (SELECT jsonb_agg(jsonb_build_object(
      'investor_id', id.investor_id,
      'investor_name', id.investor_name,
      'agreed_usd', id.agreed_contribution_usd,
      'actual_usd', id.real_contribution_usd,
      'adjustment_usd', id.adjustment_usd
    ) ORDER BY id.investor_id)
   FROM v3_report_investor_distributions id
   WHERE id.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  
  -- =========================================================================
  -- SECCIÓN 5: LIQUIDACIÓN DE COSECHA
  -- Incluye el rubro "Cosecha" distribuido proporcionalmente según % de participación
  -- =========================================================================
  (
    WITH harvest_total AS (
      SELECT
        COALESCE((
          SELECT SUM(lab.price * w.effective_area)
          FROM public.workorders w
          JOIN public.labors lab ON lab.id = w.labor_id AND lab.deleted_at IS NULL
          JOIN public.categories cat ON lab.category_id = cat.id
          WHERE w.lot_id IN (
            SELECT l.id 
            FROM public.lots l 
            JOIN public.fields f ON l.field_id = f.id 
            WHERE f.project_id = pb.project_id
          )
          AND w.deleted_at IS NULL
          AND cat.name = 'Cosecha'
          AND cat.type_id = 4
        ), 0)::numeric AS total_harvest_usd
    ),
    harvest_distribution AS (
      SELECT
        pi.project_id,
        pi.investor_id,
        i.name AS investor_name,
        pi.percentage AS share_pct_agreed,
        -- Distribuir cosecha proporcionalmente según % de participación
        (pi.percentage::numeric / 100.0) * ht.total_harvest_usd AS harvest_amount_usd,
        ht.total_harvest_usd
      FROM public.project_investors pi
      JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
      CROSS JOIN harvest_total ht
      WHERE pi.project_id = pb.project_id
    )
    SELECT jsonb_build_object(
      'rows', jsonb_build_array(
        jsonb_build_object(
          'key', 'harvest',
          'type', 'harvest',
          'label', 'Cosecha',
          'total_usd', (SELECT total_harvest_usd FROM harvest_total),
          'total_usd_ha', CASE 
            WHEN cc.total_seeded_area_ha > 0 
            THEN (SELECT total_harvest_usd FROM harvest_total) / cc.total_seeded_area_ha 
            ELSE 0 
          END,
          'investors', (
            SELECT jsonb_agg(
              jsonb_build_object(
                'investor_id', hd.investor_id,
                'investor_name', hd.investor_name,
                'amount_usd', hd.harvest_amount_usd,
                'share_pct', hd.share_pct_agreed
              ) ORDER BY hd.investor_id
            )
            FROM harvest_distribution hd
            WHERE hd.project_id = pb.project_id
          )
        )
      ),
      'footer_payment_agreed', (
        SELECT jsonb_agg(
          jsonb_build_object(
            'investor_id', hd.investor_id,
            'investor_name', hd.investor_name,
            'amount_usd', hd.harvest_amount_usd,
            'share_pct', hd.share_pct_agreed
          ) ORDER BY hd.investor_id
        )
        FROM harvest_distribution hd
        WHERE hd.project_id = pb.project_id
      ),
      'footer_payment_adjustment', '[]'::jsonb
    )
  ) AS harvest_settlement

FROM project_base pb
LEFT JOIN contribution_categories cc ON cc.project_id = pb.project_id;

COMMENT ON VIEW public.v3_investor_contribution_data_view IS 
  'Vista 4/4 consolidada para informe de Aportes por Inversor. ACTUALIZADO 168: Ajustes en Labores Generales y Arriendo.';

COMMIT;

