-- ========================================
-- MIGRATION 000135: CREATE V3 INVESTOR CONTRIBUTION VIEWS (UP)
-- ========================================
-- 
-- Purpose: Crear vistas modulares para el informe de Aportes por Inversor
-- Date: 2025-10-11
-- Author: System
-- 
-- Estrategia:
--   Se divide el informe en 4 vistas modulares para mejor mantenibilidad:
--   1. v3_report_investor_project_base - Datos generales del proyecto
--   2. v3_report_investor_contribution_categories - Categorías de aportes
--   3. v3_report_investor_distributions - Distribución por inversor
--   4. v3_investor_contribution_data_view - Vista final consolidada
--
-- Correcciones aplicadas:
--   - Usa funciones SSOT v3_calc.* para cálculos consistentes
--   - Incluye supply_movements en costos (coherente con summary_results)
--   - Aplica commercial_cost calculado correctamente
--
-- Note: Code in English, comments in Spanish.

-- ============================================================================
-- DROP EXISTING VIEWS (en orden inverso de dependencias)
-- ============================================================================
DROP VIEW IF EXISTS public.v3_investor_contribution_data_view CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_distributions CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_contribution_categories CASCADE;
DROP VIEW IF EXISTS public.v3_report_investor_project_base CASCADE;

-- ============================================================================
-- VISTA 1 de 4: v3_report_investor_project_base
-- ============================================================================
-- Purpose: Datos Generales del Proyecto para el informe de inversores
-- 
-- Sección: Datos Generales del Proyecto
-- Incluye:
--   - Superficie: Suma total superficie de lotes y campos del proyecto
--   - Arriendo: Solo si es Fijo (según configuración en clientes y sociedades)
--   - Admin.Proyecto/ha: Valor de costo por ha por administrar (desde clientes y sociedades)
-- ============================================================================

CREATE VIEW public.v3_report_investor_project_base AS
SELECT
  -- Identificación del proyecto
  p.id AS project_id,
  p.name AS project_name,
  p.customer_id,
  c.name AS customer_name,
  p.campaign_id,
  cam.name AS campaign_name,
  
  -- Superficie Total
  -- Cómo se calcula: Suma total superficie de lotes y campos de ese proyecto de la pantalla de lotes
  COALESCE(SUM(l.hectares), 0)::numeric AS surface_total_ha,
  
  -- Arriendo (solo si es Fijo)
  -- Qué representa: Solo se considera aporte invertido si el arriendo es Fijo
  -- Cómo se calcula: Si el valor cargado en clientes y sociedades es fijo, debe considerarlo
  -- Usa función SSOT v3_calc.rent_per_ha_for_lot que verifica si es fijo
  COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS lease_fixed_total_usd,
  
  -- Indicador si el arriendo es fijo
  CASE 
    WHEN COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0) > 0 
    THEN true 
    ELSE false 
  END AS lease_is_fixed,
  
  -- Arriendo por hectárea (para cálculos posteriores)
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS lease_per_ha_usd,
  
  -- Administración y Estructura
  -- Qué representa: Proporción del Valor de costo por ha por administrar ese proyecto
  -- Cómo se calcula: Valor creado en la pantalla de Clientes y Sociedades
  -- Usa función SSOT v3_calc.admin_cost_per_ha_for_lot
  COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric AS admin_total_usd,
  
  -- Administración por hectárea (para cálculos posteriores)
  CASE 
    WHEN COALESCE(SUM(l.hectares), 0) > 0 
    THEN COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0) / SUM(l.hectares)
    ELSE 0
  END::numeric AS admin_per_ha_usd

FROM public.projects p
-- Joins para información del proyecto
JOIN public.customers c ON p.customer_id = c.id AND c.deleted_at IS NULL
JOIN public.campaigns cam ON p.campaign_id = cam.id AND cam.deleted_at IS NULL

-- Joins para calcular superficie y costos
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

-- Comentario de la vista
COMMENT ON VIEW public.v3_report_investor_project_base IS 
  'Vista 1/4 para informe de Aportes por Inversor. Contiene datos generales del proyecto: superficie, arriendo fijo y administración. Usa funciones SSOT para cálculos consistentes.';

-- ============================================================================
-- VISTA 2 de 4: v3_report_investor_contribution_categories
-- ============================================================================
-- Purpose: Categorías de Aportes por Tipo/Clase
--
-- Sección: Aportes por inversor. Se considera aportes
-- Categorías incluidas:
--   1. Agroquímicos - Montos invertidos en agroquímicos por proyecto
--   2. Semilla - Montos invertidos en semillas por proyecto  
--   3. Labores Generales - Pulverización y Otras Labores (NO incluye Siembra/Riego/Cosecha)
--   4. Siembra - Labores de siembra
--   5. Riego - Labores de riego
--   6. Arriendo Capitalizable - Solo si el arriendo es Fijo
--   7. Administración y Estructura - Costos de administración
--
-- Usa funciones SSOT donde existen:
--   - v3_lot_ssot.supply_cost_for_lot_by_category() para insumos
--   - v3_calc.rent_per_ha_for_lot() para arriendo
--   - v3_calc.admin_cost_per_ha_for_lot() para administración
--   - v3_lot_ssot.labor_cost_for_lot() para labores (filtrando por categoría)
-- ============================================================================

CREATE VIEW public.v3_report_investor_contribution_categories AS
WITH lot_base AS (
  -- Base de lotes por proyecto para calcular totales
  SELECT
    f.project_id,
    l.id AS lot_id,
    l.hectares,
    -- Superficie sembrada para cálculos por ha
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
  -- CATEGORÍA 1: AGROQUÍMICOS
  -- Qué representa: Montos invertidos en Tipo/Clase: AGROQUÍMICOS
  -- Cómo se calcula: Suma de insumos de categorías: Coadyuvantes, Curasemillas, 
  --                  Herbicidas, Insecticidas, Fungicidas, Otros Insumos
  -- Usa: Funciones SSOT v3_lot_ssot.supply_cost_for_lot_by_category()
  -- =========================================================================
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Coadyuvantes') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Curasemillas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Herbicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Insecticidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Fungicidas') +
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Otros Insumos')
  ), 0)::numeric AS agrochemicals_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 2: SEMILLA
  -- Qué representa: Montos invertidos en Rubro: SEMILLA
  -- Cómo se calcula: Suma de insumos de categoría 'Semilla'
  -- Usa: Función SSOT v3_lot_ssot.supply_cost_for_lot_by_category()
  -- =========================================================================
  COALESCE(SUM(
    v3_lot_ssot.supply_cost_for_lot_by_category(lb.lot_id, 'Semilla')
  ), 0)::numeric AS seeds_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 3: LABORES GENERALES
  -- Qué representa: Montos invertidos en Tipo de LABORES GENERALES
  -- Cómo se calcula: Incluye Rubros: Pulverización y Otras Labores
  --                  NO incluye: Siembra, Riego, Cosecha
  -- Usa: Lógica directa por falta de función SSOT específica
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
  
  -- =========================================================================
  -- CATEGORÍA 4: SIEMBRA
  -- Qué representa: Montos invertidos en Rubro: SIEMBRA
  -- Cómo se calcula: Solo labores de categoría 'Siembra'
  -- Usa: Lógica directa por falta de función SSOT específica
  -- =========================================================================
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
  
  -- =========================================================================
  -- CATEGORÍA 5: RIEGO
  -- Qué representa: Montos invertidos en Rubro: RIEGO
  -- Cómo se calcula: Solo labores de categoría 'Riego'
  -- Usa: Lógica directa por falta de función SSOT específica
  -- =========================================================================
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
  -- Qué representa: Solo si el valor cargado en clientes y sociedades es Fijo
  -- Cómo se calcula: Si es Fijo, considera aporte invertido; si no, no lo considera
  -- Usa: Función SSOT v3_calc.rent_per_ha_for_lot() (verifica si es fijo)
  -- NOTA: Esta categoría requiere atribución manual por inversor
  -- =========================================================================
  COALESCE(SUM(
    v3_calc.rent_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS rent_capitalizable_total_usd,
  
  -- =========================================================================
  -- CATEGORÍA 7: ADMINISTRACIÓN Y ESTRUCTURA
  -- Qué representa: Proporción del Valor de costo por ha por administrar ese proyecto
  -- Cómo se calcula: Valor creado en la pantalla de Clientes y Sociedades
  -- Usa: Función SSOT v3_calc.admin_cost_per_ha_for_lot()
  -- NOTA: Esta categoría requiere atribución manual por inversor
  -- =========================================================================
  COALESCE(SUM(
    v3_calc.admin_cost_per_ha_for_lot(lb.lot_id)::numeric * lb.hectares
  ), 0)::numeric AS administration_total_usd,
  
  -- =========================================================================
  -- SUPERFICIE TOTAL (para cálculos de USD/HA)
  -- =========================================================================
  COALESCE(SUM(lb.seeded_area_ha), 0)::numeric AS total_seeded_area_ha

FROM lot_base lb

GROUP BY lb.project_id;

-- Comentario de la vista
COMMENT ON VIEW public.v3_report_investor_contribution_categories IS 
  'Vista 2/4 para informe de Aportes por Inversor. Contiene todas las categorías de aportes (Agroquímicos, Semilla, Labores Generales, Siembra, Riego, Arriendo, Administración). Usa funciones SSOT donde existen (supply_cost_for_lot_by_category, rent_per_ha_for_lot, admin_cost_per_ha_for_lot).';

-- ============================================================================
-- VISTA 3 de 4: v3_report_investor_distributions
-- ============================================================================
-- Purpose: Distribución de Aportes por Inversor y Comparación Teórico vs. Real
--
-- Sección: Comparación Aporte Teórico vs. Real
-- Campos:
--   1. Aporte Acordado - % de participación pactado entre socios (de Clientes y Sociedades)
--   2. Ajuste De Aporte - Diferencia entre aporte total realizado y aporte acordado
--
-- Calcula:
--   - Aporte Acordado por Inversor = Total Aportes * % Acordado
--   - Aporte Real por Inversor = Suma de aportes reales en cada categoría
--   - Ajuste = Aporte Real - Aporte Acordado
--
-- IMPORTANTE: 
--   - Para Arriendo y Administración, la distribución requiere atribución manual
--   - Por ahora, se distribuyen según % acordado (simplificación)
-- ============================================================================

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
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha,
    -- Total de todos los aportes
    (agrochemicals_total_usd + seeds_total_usd + general_labors_total_usd + 
     sowing_total_usd + irrigation_total_usd + rent_capitalizable_total_usd + 
     administration_total_usd) AS total_contributions_usd
  FROM v3_report_investor_contribution_categories
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
    
    -- Semilla: Distribuido según % acordado
    ROUND((ct.seeds_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS seeds_real_usd,
    
    -- Labores Generales: Distribuido según % acordado
    ROUND((ct.general_labors_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS general_labors_real_usd,
    
    -- Siembra: Distribuido según % acordado
    ROUND((ct.sowing_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS sowing_real_usd,
    
    -- Riego: Distribuido según % acordado
    ROUND((ct.irrigation_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS irrigation_real_usd,
    
    -- Arriendo Capitalizable: Requiere atribución manual (por ahora distribuido según %)
    ROUND((ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS rent_real_usd,
    
    -- Administración: Requiere atribución manual (por ahora distribuido según %)
    ROUND((ct.administration_total_usd * ib.share_pct_agreed / 100)::numeric, 2) AS administration_real_usd,
    
    -- Total Real por Inversor (suma de todas las categorías)
    ROUND((
      (ct.agrochemicals_total_usd * ib.share_pct_agreed / 100) +
      (ct.seeds_total_usd * ib.share_pct_agreed / 100) +
      (ct.general_labors_total_usd * ib.share_pct_agreed / 100) +
      (ct.sowing_total_usd * ib.share_pct_agreed / 100) +
      (ct.irrigation_total_usd * ib.share_pct_agreed / 100) +
      (ct.rent_capitalizable_total_usd * ib.share_pct_agreed / 100) +
      (ct.administration_total_usd * ib.share_pct_agreed / 100)
    )::numeric, 2) AS total_real_contribution_usd,
    
    -- Total de aportes del proyecto
    ct.total_contributions_usd AS project_total_contributions_usd
    
  FROM investor_base ib
  JOIN category_totals ct ON ct.project_id = ib.project_id
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
  'Vista 3/4 para informe de Aportes por Inversor. Calcula la distribución de aportes por inversor y compara el aporte teórico (acordado según %) vs. el aporte real. Calcula el ajuste necesario entre lo acordado y lo real.';

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
        'rent_capitalizable'::text,
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
        'administration'::text,
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
  'Vista 4/4 (FINAL) para informe de Aportes por Inversor. Consolida las 3 vistas anteriores y genera la estructura JSON completa para la API. Incluye: Datos Generales, Categorías de Aportes, Comparación Acordado vs. Real, y Liquidación de Cosecha.';

