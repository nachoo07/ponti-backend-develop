-- ========================================
-- MIGRATION 000205: Fix investor harvest payments distribution (DOWN)
-- ========================================
--
-- Objetivo: Revertir la corrección de pagos reales de cosecha y volver al
--           comportamiento anterior donde los pagos se prorratean usando los
--           porcentajes acordados.
--

BEGIN;

DROP VIEW IF EXISTS public.v3_dashboard_contributions_progress;
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view;

-- Restaurar vista con claves anteriores (rent_capitalizable y administration).
CREATE VIEW public.v3_investor_contribution_data_view AS
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
        'share_pct', ROUND(id.share_pct_agreed::numeric, 2)
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
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', id.investor_id,
        'investor_name', id.investor_name,
        'agreed_share_pct', id.share_pct_agreed,
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
          )::numeric, 2),
          'share_pct', ROUND(id.share_pct_agreed::numeric, 2)
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

COMMENT ON VIEW public.v3_investor_contribution_data_view IS
  'Vista 4/4 (FINAL) para informe de Aportes por Inversor. Versión anterior a FIX 000194.';

-- Restaurar vista auxiliar del dashboard (idéntica a la previa).
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
'Avance de aportes de inversores. Estado previo a FIX 000194.';

COMMIT;
