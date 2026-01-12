-- ========================================
-- MIGRATION 000326: Create v4_report investor views with bug fixes (UP)
-- ========================================
--
-- Objetivo: Migrar vistas de aportes de inversores a v4_report con corrección de bugs.
--
-- BUG CORREGIDO: La vista v3_report_investor_distributions calculaba aportes de
--                insumos y labores usando "total × % acordado" cuando debería
--                sumar los aportes REALES desde supply_movements.investor_id y
--                workorders.investor_id.
--
-- Vistas incluidas:
--   1. v4_report.investor_project_base (datos generales del proyecto)
--   2. v4_report.investor_contribution_categories (totales por categoría)
--   3. v4_report.investor_distributions (FIX: aportes reales por inversor)
--   4. v4_report.investor_contribution_data (vista final para el reporte)
--
-- Nota: Código en inglés, comentarios en español.

BEGIN;

-- =========================================================================
-- Vista 1/4: investor_project_base
-- Datos generales del proyecto: superficie, arriendo y administración
-- =========================================================================
CREATE VIEW v4_report.investor_project_base AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  COALESCE(SUM(l.hectares), 0::double precision)::numeric AS surface_total_ha,
  COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS lease_fixed_total_usd,
  CASE
    WHEN COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) > 0::double precision THEN TRUE
    ELSE FALSE
  END AS lease_is_fixed,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS lease_per_ha_usd,
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision)::numeric AS admin_total_usd,
  CASE
    WHEN COALESCE(SUM(l.hectares), 0::double precision) > 0::double precision THEN
      COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0::double precision) / SUM(l.hectares)
    ELSE 0::double precision
  END::numeric AS admin_per_ha_usd
FROM public.projects p
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL
LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY
  p.id,
  p.name,
  p.customer_id,
  c.name,
  p.campaign_id,
  cam.name;

COMMENT ON VIEW v4_report.investor_project_base IS
'Vista 1/4 para informe de Aportes por Inversor. Contiene datos generales del proyecto (000326).';

-- =========================================================================
-- Vista 2/4: investor_contribution_categories
-- Totales por categoría de aporte (sin desglose por inversor)
-- =========================================================================
CREATE VIEW v4_report.investor_contribution_categories AS
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

COMMENT ON VIEW v4_report.investor_contribution_categories IS 
'Vista 2/4 para informe de aportes. Totales por categoría sin desglose por inversor (000326).';

-- =========================================================================
-- Vista 3/4: investor_distributions
-- FIX: Distribución REAL de aportes por inversor (no acordado × total)
-- =========================================================================
CREATE VIEW v4_report.investor_distributions AS
WITH investor_base AS (
  -- Base de inversores por proyecto con sus % acordados
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.deleted_at IS NULL
),
admin_config AS (
  -- Detectar si un proyecto tiene porcentajes específicos de administración
  SELECT
    project_id,
    COUNT(*) FILTER (WHERE deleted_at IS NULL) > 0 AS has_custom_admin
  FROM public.admin_cost_investors
  GROUP BY project_id
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
    (agrochemicals_total_usd + fertilizers_total_usd + seeds_total_usd +
     general_labors_total_usd + sowing_total_usd + irrigation_total_usd +
     rent_capitalizable_total_usd + administration_total_usd) AS total_contributions_usd
  FROM v4_report.investor_contribution_categories
),
-- =======================================================================
-- FIX: Aportes REALES de insumos desde supply_movements.investor_id
-- =======================================================================
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
-- =======================================================================
-- FIX: Aportes REALES de labores desde workorders.investor_id
-- =======================================================================
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
    AND cat.name IN ('Pulverización', 'Otras Labores')
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
-- =======================================================================
-- Arriendo: usa field_investors (ya estaba bien)
-- =======================================================================
investor_rent_real AS (
  SELECT
    f.project_id,
    fi.investor_id,
    SUM(
      v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares *
      COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN public.field_investors fi ON fi.field_id = f.id AND fi.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, fi.investor_id
),
-- =======================================================================
-- Administración: usa admin_cost_investors (ya estaba bien)
-- =======================================================================
investor_admin_real AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    (
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0)::numeric / 100)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100)
      END
    )::numeric AS admin_real_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN admin_config ac ON ac.project_id = ib.project_id
  LEFT JOIN public.admin_cost_investors aci
    ON aci.project_id = ib.project_id
   AND aci.investor_id = ib.investor_id
   AND aci.deleted_at IS NULL
),
-- =======================================================================
-- Combinar todos los aportes reales
-- =======================================================================
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    -- FIX: Usar aportes REALES desde transacciones, no total × % acordado
    COALESCE(agro.agrochemicals_real_usd, 0)::numeric AS agrochemicals_real_usd,
    COALESCE(fert.fertilizers_real_usd, 0)::numeric AS fertilizers_real_usd,
    COALESCE(seed.seeds_real_usd, 0)::numeric AS seeds_real_usd,
    COALESCE(glabor.general_labors_real_usd, 0)::numeric AS general_labors_real_usd,
    COALESCE(sow.sowing_real_usd, 0)::numeric AS sowing_real_usd,
    COALESCE(irrig.irrigation_real_usd, 0)::numeric AS irrigation_real_usd,
    COALESCE(rri.rent_real_usd, 0)::numeric AS rent_real_usd,
    COALESCE(ia.admin_real_usd, 0)::numeric AS administration_real_usd,
    -- Total real aportado por el inversor
    (
      COALESCE(agro.agrochemicals_real_usd, 0) +
      COALESCE(fert.fertilizers_real_usd, 0) +
      COALESCE(seed.seeds_real_usd, 0) +
      COALESCE(glabor.general_labors_real_usd, 0) +
      COALESCE(sow.sowing_real_usd, 0) +
      COALESCE(irrig.irrigation_real_usd, 0) +
      COALESCE(rri.rent_real_usd, 0) +
      COALESCE(ia.admin_real_usd, 0)
    )::numeric AS total_real_contribution_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  -- FIX: LEFT JOIN con aportes reales por investor_id
  LEFT JOIN investor_agrochemicals_real agro
    ON agro.project_id = ib.project_id AND agro.investor_id = ib.investor_id
  LEFT JOIN investor_fertilizers_real fert
    ON fert.project_id = ib.project_id AND fert.investor_id = ib.investor_id
  LEFT JOIN investor_seeds_real seed
    ON seed.project_id = ib.project_id AND seed.investor_id = ib.investor_id
  LEFT JOIN investor_general_labors_real glabor
    ON glabor.project_id = ib.project_id AND glabor.investor_id = ib.investor_id
  LEFT JOIN investor_sowing_real sow
    ON sow.project_id = ib.project_id AND sow.investor_id = ib.investor_id
  LEFT JOIN investor_irrigation_real irrig
    ON irrig.project_id = ib.project_id AND irrig.investor_id = ib.investor_id
  LEFT JOIN investor_rent_real rri
    ON rri.project_id = ib.project_id AND rri.investor_id = ib.investor_id
  LEFT JOIN investor_admin_real ia
    ON ia.project_id = ib.project_id AND ia.investor_id = ib.investor_id
),
investor_agreed_vs_real AS (
  -- Comparar aporte acordado vs aporte real
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

COMMENT ON VIEW v4_report.investor_distributions IS
'Vista 3/4 para informe de Aportes por Inversor. FIX 000326: Usa aportes REALES desde supply_movements y workorders por investor_id.';

-- =========================================================================
-- Vista 4/4: investor_contribution_data
-- Vista final que combina todo para el reporte
-- =========================================================================
CREATE VIEW v4_report.investor_contribution_data AS
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
project_seed_data AS (
  SELECT
    project_id,
    MAX(total_seeded_area_ha)::numeric AS total_seeded_area_ha
  FROM v4_report.investor_contribution_categories
  GROUP BY project_id
),
investor_harvest_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS harvest_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Cosecha'
  GROUP BY w.project_id, w.investor_id
),
harvest_totals AS (
  SELECT
    psd.project_id,
    COALESCE(SUM(hr.harvest_real_usd), 0)::numeric AS total_harvest_usd,
    CASE
      WHEN COALESCE(psd.total_seeded_area_ha, 0) > 0
      THEN COALESCE(SUM(hr.harvest_real_usd), 0) / psd.total_seeded_area_ha
      ELSE 0
    END::numeric AS total_harvest_usd_ha
  FROM project_seed_data psd
  LEFT JOIN investor_harvest_real hr ON hr.project_id = psd.project_id
  GROUP BY psd.project_id, psd.total_seeded_area_ha
)
SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  -- Headers de inversores
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'share_pct', ROUND(id.share_pct_agreed::numeric, 2)
      )
      ORDER BY id.investor_id
    )
    FROM v4_report.investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_headers,
  -- Datos generales del proyecto
  jsonb_build_object(
    'surface_total_ha', ROUND(pb.surface_total_ha::numeric, 2),
    'lease_fixed_total_usd', ROUND(pb.lease_fixed_total_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2),
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2)
  ) AS general_project_data,
  -- Categorías de aportes (FIX: usa aportes reales)
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution,
        NULL AS attribution_note
      FROM v4_report.investor_contribution_categories cc
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Semillas
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Arriendo Capitalizable
      SELECT
        'capitalizable_lease'::text,
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Administración y Estructura
      SELECT
        'administration_structure'::text,
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
          FROM v4_report.investor_distributions id
          WHERE id.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
    ) AS cat
  ) AS contribution_categories,
  -- Comparación de aportes acordado vs real
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
    FROM v4_report.investor_distributions id
    WHERE id.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  -- Liquidación de cosecha
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', ROUND(COALESCE(ht.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(ht.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', ROUND(COALESCE(hr.harvest_real_usd, 0)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      ),
      jsonb_build_object(
        'key', 'totals',
        'type', 'totals',
        'total_usd', ROUND(COALESCE(ht.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(ht.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', ROUND(COALESCE(hr.harvest_real_usd, 0)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      )
    ),
    'footer_payment_agreed', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', ROUND((COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)::numeric, 2),
          'share_pct', ROUND(ib.share_pct_agreed::numeric, 2)
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb),
    'footer_payment_adjustment', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', ROUND((
            COALESCE(hr.harvest_real_usd, 0)::numeric -
            (COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)
          )::numeric, 2),
          'share_pct', ROUND(ib.share_pct_agreed::numeric, 2)
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      LEFT JOIN investor_harvest_real hr
        ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb)
  ) AS harvest_settlement
FROM v4_report.investor_project_base pb
JOIN v4_report.investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_totals ht ON ht.project_id = pb.project_id
ORDER BY pb.project_id;

COMMENT ON VIEW v4_report.investor_contribution_data IS
'Vista 4/4 (FINAL) para informe de Aportes por Inversor. FIX 000326: Usa aportes REALES por inversor desde transacciones.';

COMMIT;
