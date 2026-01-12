-- ============================================================================
-- SCRIPT DE VALIDACIÓN: Paridad v3 vs v4 Report Field Crop Economicos
-- ============================================================================
--
-- Propósito: Verificar que v4_report_field_crop_economicos produce los mismos
--            resultados que v3_report_field_crop_economicos.
--
-- Uso: Ejecutar después de aplicar migraciones 000208 y 000209.
--      Si el resultado está vacío, la paridad es correcta.
--
-- Secciones:
--   1. Verificación de unicidad
--   2. Detección de filas faltantes/sobrantes
--   3. Comparación numérica con tolerancia
--
-- ============================================================================


-- ============================================================================
-- SECCIÓN 1: Verificación de Unicidad
-- ============================================================================
-- Verifica que (project_id, field_id, current_crop_id) sea único en ambas vistas

SELECT 'DUPLICATES_V3' AS issue_type, project_id, field_id, current_crop_id, COUNT(*) AS row_count
FROM public.v3_report_field_crop_economicos
GROUP BY project_id, field_id, current_crop_id
HAVING COUNT(*) > 1

UNION ALL

SELECT 'DUPLICATES_V4' AS issue_type, project_id, field_id, current_crop_id, COUNT(*) AS row_count
FROM public.v4_report_field_crop_economicos
GROUP BY project_id, field_id, current_crop_id
HAVING COUNT(*) > 1;


-- ============================================================================
-- SECCIÓN 2: Filas Faltantes o Sobrantes
-- ============================================================================
-- Detecta filas que existen en una vista pero no en la otra

SELECT 
  CASE 
    WHEN v3.project_id IS NULL THEN 'MISSING_IN_V3'
    WHEN v4.project_id IS NULL THEN 'MISSING_IN_V4'
  END AS issue_type,
  COALESCE(v3.project_id, v4.project_id) AS project_id,
  COALESCE(v3.field_id, v4.field_id) AS field_id,
  COALESCE(v3.current_crop_id, v4.current_crop_id) AS current_crop_id
FROM public.v3_report_field_crop_economicos v3
FULL OUTER JOIN public.v4_report_field_crop_economicos v4
  ON v3.project_id = v4.project_id
  AND v3.field_id = v4.field_id
  AND v3.current_crop_id = v4.current_crop_id
WHERE v3.project_id IS NULL OR v4.project_id IS NULL;


-- ============================================================================
-- SECCIÓN 3: Comparación Numérica con Tolerancia
-- ============================================================================
-- Compara cada columna numérica. Reporta diferencias > 0.01 USD
-- Mapeo de columnas: v3 (español) → v4 (inglés)

WITH comparison AS (
  SELECT
    v3.project_id,
    v3.field_id,
    v3.current_crop_id,
    
    -- gastos_directos_usd ↔ direct_costs_usd
    v3.gastos_directos_usd AS v3_direct_costs_usd,
    v4.direct_costs_usd AS v4_direct_costs_usd,
    ABS(COALESCE(v3.gastos_directos_usd, 0) - COALESCE(v4.direct_costs_usd, 0)) AS diff_direct_costs_usd,
    
    -- gastos_directos_usd_ha ↔ direct_costs_usd_ha
    v3.gastos_directos_usd_ha AS v3_direct_costs_usd_ha,
    v4.direct_costs_usd_ha AS v4_direct_costs_usd_ha,
    ABS(COALESCE(v3.gastos_directos_usd_ha, 0) - COALESCE(v4.direct_costs_usd_ha, 0)) AS diff_direct_costs_usd_ha,
    
    -- margen_bruto_usd ↔ gross_margin_usd
    v3.margen_bruto_usd AS v3_gross_margin_usd,
    v4.gross_margin_usd AS v4_gross_margin_usd,
    ABS(COALESCE(v3.margen_bruto_usd, 0) - COALESCE(v4.gross_margin_usd, 0)) AS diff_gross_margin_usd,
    
    -- margen_bruto_usd_ha ↔ gross_margin_usd_ha
    v3.margen_bruto_usd_ha AS v3_gross_margin_usd_ha,
    v4.gross_margin_usd_ha AS v4_gross_margin_usd_ha,
    ABS(COALESCE(v3.margen_bruto_usd_ha, 0) - COALESCE(v4.gross_margin_usd_ha, 0)) AS diff_gross_margin_usd_ha,
    
    -- arriendo_usd ↔ rent_usd
    v3.arriendo_usd AS v3_rent_usd,
    v4.rent_usd AS v4_rent_usd,
    ABS(COALESCE(v3.arriendo_usd, 0) - COALESCE(v4.rent_usd, 0)) AS diff_rent_usd,
    
    -- arriendo_usd_ha ↔ rent_usd_ha
    v3.arriendo_usd_ha AS v3_rent_usd_ha,
    v4.rent_usd_ha AS v4_rent_usd_ha,
    ABS(COALESCE(v3.arriendo_usd_ha, 0) - COALESCE(v4.rent_usd_ha, 0)) AS diff_rent_usd_ha,
    
    -- adm_estructura_usd ↔ admin_usd
    v3.adm_estructura_usd AS v3_admin_usd,
    v4.admin_usd AS v4_admin_usd,
    ABS(COALESCE(v3.adm_estructura_usd, 0) - COALESCE(v4.admin_usd, 0)) AS diff_admin_usd,
    
    -- adm_estructura_usd_ha ↔ admin_usd_ha
    v3.adm_estructura_usd_ha AS v3_admin_usd_ha,
    v4.admin_usd_ha AS v4_admin_usd_ha,
    ABS(COALESCE(v3.adm_estructura_usd_ha, 0) - COALESCE(v4.admin_usd_ha, 0)) AS diff_admin_usd_ha,
    
    -- resultado_operativo_usd ↔ operating_result_usd
    v3.resultado_operativo_usd AS v3_operating_result_usd,
    v4.operating_result_usd AS v4_operating_result_usd,
    ABS(COALESCE(v3.resultado_operativo_usd, 0) - COALESCE(v4.operating_result_usd, 0)) AS diff_operating_result_usd,
    
    -- resultado_operativo_usd_ha ↔ operating_result_usd_ha
    v3.resultado_operativo_usd_ha AS v3_operating_result_usd_ha,
    v4.operating_result_usd_ha AS v4_operating_result_usd_ha,
    ABS(COALESCE(v3.resultado_operativo_usd_ha, 0) - COALESCE(v4.operating_result_usd_ha, 0)) AS diff_operating_result_usd_ha

  FROM public.v3_report_field_crop_economicos v3
  JOIN public.v4_report_field_crop_economicos v4
    ON v3.project_id = v4.project_id
    AND v3.field_id = v4.field_id
    AND v3.current_crop_id = v4.current_crop_id
)
SELECT
  'NUMERIC_DIFF' AS issue_type,
  project_id,
  field_id,
  current_crop_id,
  -- Mostrar solo las columnas con diferencias significativas
  CASE WHEN diff_direct_costs_usd > 0.01 
       THEN 'direct_costs_usd: v3=' || v3_direct_costs_usd || ' v4=' || v4_direct_costs_usd 
       ELSE NULL END AS diff_1,
  CASE WHEN diff_direct_costs_usd_ha > 0.01 
       THEN 'direct_costs_usd_ha: v3=' || v3_direct_costs_usd_ha || ' v4=' || v4_direct_costs_usd_ha 
       ELSE NULL END AS diff_2,
  CASE WHEN diff_gross_margin_usd > 0.01 
       THEN 'gross_margin_usd: v3=' || v3_gross_margin_usd || ' v4=' || v4_gross_margin_usd 
       ELSE NULL END AS diff_3,
  CASE WHEN diff_gross_margin_usd_ha > 0.01 
       THEN 'gross_margin_usd_ha: v3=' || v3_gross_margin_usd_ha || ' v4=' || v4_gross_margin_usd_ha 
       ELSE NULL END AS diff_4,
  CASE WHEN diff_rent_usd > 0.01 
       THEN 'rent_usd: v3=' || v3_rent_usd || ' v4=' || v4_rent_usd 
       ELSE NULL END AS diff_5,
  CASE WHEN diff_rent_usd_ha > 0.01 
       THEN 'rent_usd_ha: v3=' || v3_rent_usd_ha || ' v4=' || v4_rent_usd_ha 
       ELSE NULL END AS diff_6,
  CASE WHEN diff_admin_usd > 0.01 
       THEN 'admin_usd: v3=' || v3_admin_usd || ' v4=' || v4_admin_usd 
       ELSE NULL END AS diff_7,
  CASE WHEN diff_admin_usd_ha > 0.01 
       THEN 'admin_usd_ha: v3=' || v3_admin_usd_ha || ' v4=' || v4_admin_usd_ha 
       ELSE NULL END AS diff_8,
  CASE WHEN diff_operating_result_usd > 0.01 
       THEN 'operating_result_usd: v3=' || v3_operating_result_usd || ' v4=' || v4_operating_result_usd 
       ELSE NULL END AS diff_9,
  CASE WHEN diff_operating_result_usd_ha > 0.01 
       THEN 'operating_result_usd_ha: v3=' || v3_operating_result_usd_ha || ' v4=' || v4_operating_result_usd_ha 
       ELSE NULL END AS diff_10
FROM comparison
WHERE 
  diff_direct_costs_usd > 0.01 OR
  diff_direct_costs_usd_ha > 0.01 OR
  diff_gross_margin_usd > 0.01 OR
  diff_gross_margin_usd_ha > 0.01 OR
  diff_rent_usd > 0.01 OR
  diff_rent_usd_ha > 0.01 OR
  diff_admin_usd > 0.01 OR
  diff_admin_usd_ha > 0.01 OR
  diff_operating_result_usd > 0.01 OR
  diff_operating_result_usd_ha > 0.01;


-- ============================================================================
-- RESUMEN FINAL
-- ============================================================================
-- Cuenta total de diferencias encontradas

SELECT 
  'SUMMARY' AS report_type,
  (SELECT COUNT(*) FROM public.v3_report_field_crop_economicos) AS v3_row_count,
  (SELECT COUNT(*) FROM public.v4_report_field_crop_economicos) AS v4_row_count,
  CASE 
    WHEN (SELECT COUNT(*) FROM public.v3_report_field_crop_economicos) = 
         (SELECT COUNT(*) FROM public.v4_report_field_crop_economicos)
    THEN 'ROW_COUNT_MATCH'
    ELSE 'ROW_COUNT_MISMATCH'
  END AS row_count_status;
