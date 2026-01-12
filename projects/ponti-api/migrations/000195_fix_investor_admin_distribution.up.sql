-- ========================================
-- MIGRATION 000195: Fix investor admin distribution (UP)
-- ========================================
--
-- Objetivo: eliminar redondeos y usar porcentajes específicos de
--           admin_cost_investors cuando existan para distribuir
--           Administración y Estructura. Además, mantener el cálculo
--           consistente en dashboard y en el informe de aportes.
--
-- Nota: Código en inglés, comentarios en español.

BEGIN;

-- ============================================================================
-- Actualizar vista v3_report_investor_distributions con lógica corregida
-- ============================================================================
CREATE OR REPLACE VIEW public.v3_report_investor_distributions AS
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
admin_config AS (
  -- Detectar si un proyecto tiene porcentajes específicos de administración
  SELECT
    project_id,
    COUNT(*) FILTER (WHERE deleted_at IS NULL) > 0 AS has_custom_admin
  FROM public.admin_cost_investors
  GROUP BY project_id
),
category_totals AS (
  -- Totales por categoría (vista 2)
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
  FROM public.v3_report_investor_contribution_categories
),
rent_real_by_investor AS (
  -- Calcular arriendo real usando porcentajes de field_investors
  SELECT
    f.project_id,
    LOWER(inv.name) AS investor_key,
    SUM(
      v3_lot_ssot.rent_fixed_only_for_lot(l.id) * l.hectares *
      COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.field_investors fi
  JOIN public.fields f ON f.id = fi.field_id AND f.deleted_at IS NULL
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  JOIN public.investors inv ON inv.id = fi.investor_id AND inv.deleted_at IS NULL
  WHERE fi.deleted_at IS NULL
  GROUP BY f.project_id, LOWER(inv.name)
),
investor_real_contributions AS (
  -- Aportes reales por inversor
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100)::numeric AS agrochemicals_real_usd,
    (ct.fertilizers_total_usd * ib.share_pct_agreed / 100)::numeric AS fertilizers_real_usd,
    (ct.seeds_total_usd * ib.share_pct_agreed / 100)::numeric AS seeds_real_usd,
    (ct.general_labors_total_usd * ib.share_pct_agreed / 100)::numeric AS general_labors_real_usd,
    (ct.sowing_total_usd * ib.share_pct_agreed / 100)::numeric AS sowing_real_usd,
    (ct.irrigation_total_usd * ib.share_pct_agreed / 100)::numeric AS irrigation_real_usd,
    COALESCE(rri.rent_real_usd, 0)::numeric AS rent_real_usd,
    (
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0)::numeric / 100)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100)
      END
    )::numeric AS administration_real_usd,
    (
      (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100) +
      (ct.fertilizers_total_usd * ib.share_pct_agreed / 100) +
      (ct.seeds_total_usd * ib.share_pct_agreed / 100) +
      (ct.general_labors_total_usd * ib.share_pct_agreed / 100) +
      (ct.sowing_total_usd * ib.share_pct_agreed / 100) +
      (ct.irrigation_total_usd * ib.share_pct_agreed / 100) +
      COALESCE(rri.rent_real_usd, 0) +
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0)::numeric / 100)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100)
      END
    )::numeric AS total_real_contribution_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
  LEFT JOIN admin_config ac ON ac.project_id = ib.project_id
  LEFT JOIN rent_real_by_investor rri
    ON rri.project_id = ib.project_id
   AND rri.investor_key = LOWER(ib.investor_name)
  LEFT JOIN public.admin_cost_investors aci
    ON aci.project_id = ib.project_id
   AND aci.investor_id = ib.investor_id
   AND aci.deleted_at IS NULL
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

COMMENT ON VIEW public.v3_report_investor_distributions IS
  'Vista 3/4 para informe de Aportes por Inversor. FIX 000195: Administración usa admin_cost_investors y se eliminan redondeos.';

-- ============================================================================
-- Actualizar vista v3_dashboard_contributions_progress con misma lógica
-- ============================================================================
CREATE OR REPLACE VIEW public.v3_dashboard_contributions_progress AS
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
admin_config AS (
  SELECT
    project_id,
    COUNT(*) FILTER (WHERE deleted_at IS NULL) > 0 AS has_custom_admin
  FROM public.admin_cost_investors
  GROUP BY project_id
),
category_totals AS (
  SELECT
    project_id,
    administration_total_usd,
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
      COALESCE(fi.percentage, 0)::numeric / 100
    )::numeric AS rent_real_usd
  FROM public.fields f
  JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN public.field_investors fi ON fi.field_id = f.id AND fi.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, fi.investor_id
),
investor_admin_real AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    (
      CASE
        WHEN COALESCE(ac.has_custom_admin, FALSE)
          THEN (ct.administration_total_usd * COALESCE(aci.percentage, 0) / 100.0)
        ELSE (ct.administration_total_usd * ib.share_pct_agreed / 100.0)
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
investor_real_contributions AS (
  SELECT
    ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    COALESCE(agro.agrochemicals_real_usd, 0) AS agrochemicals_real_usd,
    COALESCE(fert.fertilizers_real_usd, 0) AS fertilizers_real_usd,
    COALESCE(seed.seeds_real_usd, 0) AS seeds_real_usd,
    COALESCE(glabor.general_labors_real_usd, 0) AS general_labors_real_usd,
    COALESCE(sow.sowing_real_usd, 0) AS sowing_real_usd,
    COALESCE(irrig.irrigation_real_usd, 0) AS irrigation_real_usd,
    COALESCE(ir.rent_real_usd, 0) AS rent_real_usd,
    COALESCE(ia.admin_real_usd, 0) AS admin_real_usd,
    ct.total_contributions_usd AS project_total_contributions_usd
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
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
  LEFT JOIN investor_rent_real ir
    ON ir.project_id = ib.project_id AND ir.investor_id = ib.investor_id
  LEFT JOIN investor_admin_real ia
    ON ia.project_id = ib.project_id AND ia.investor_id = ib.investor_id
)
SELECT
  project_id,
  investor_id,
  investor_name,
  share_pct_agreed AS investor_percentage_pct,
  CASE
    WHEN project_total_contributions_usd > 0
      THEN (COALESCE(agrochemicals_real_usd, 0) +
            COALESCE(fertilizers_real_usd, 0) +
            COALESCE(seeds_real_usd, 0) +
            COALESCE(general_labors_real_usd, 0) +
            COALESCE(sowing_real_usd, 0) +
            COALESCE(irrigation_real_usd, 0) +
            COALESCE(rent_real_usd, 0) +
            COALESCE(admin_real_usd, 0)
           ) / project_total_contributions_usd * 100
    ELSE 0
  END::numeric AS contributions_progress_pct
FROM investor_real_contributions;

COMMENT ON VIEW public.v3_dashboard_contributions_progress IS
  'Avance de aportes de inversores. FIX 000195: Administración usa admin_cost_investors y se eliminan redondeos.';

COMMIT;


