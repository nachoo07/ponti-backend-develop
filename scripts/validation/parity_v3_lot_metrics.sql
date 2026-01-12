-- =============================================================================
-- PARIDAD: v3_lot_metrics vs v4_report.lot_metrics
-- Ejecutar: psql -v ON_ERROR_STOP=1 -d ponti -f parity_v3_lot_metrics.sql
-- =============================================================================

DO $$
DECLARE
  v_errors int;
  r record;
BEGIN
  CREATE TEMP TABLE parity_results (
    check_type text,
    lot_id bigint,
    column_name text,
    message text
  ) ON COMMIT DROP;

  -- CHECK 1: Unicidad v4
  INSERT INTO parity_results
  SELECT 'UNICIDAD_V4', NULL, NULL,
    'FAIL: ' || (COUNT(*) - COUNT(DISTINCT lot_id))::text || ' duplicados'
  FROM v4_report.lot_metrics
  HAVING COUNT(*) <> COUNT(DISTINCT lot_id);

  -- CHECK 2: Rowcount
  INSERT INTO parity_results
  SELECT 'ROWCOUNT', NULL, NULL,
    'FAIL: v3=' || v3_count::text || ' v4=' || v4_count::text
  FROM (
    SELECT 
      (SELECT COUNT(*) FROM v3_lot_metrics) AS v3_count,
      (SELECT COUNT(*) FROM v4_report.lot_metrics) AS v4_count
  ) c
  WHERE v3_count <> v4_count;

  -- CHECK 3: Keys faltantes
  INSERT INTO parity_results
  SELECT 
    CASE WHEN v3.lot_id IS NULL THEN 'ONLY_V4' ELSE 'ONLY_V3' END,
    COALESCE(v3.lot_id, v4.lot_id), NULL,
    CASE WHEN v3.lot_id IS NULL THEN 'Existe solo en v4' ELSE 'Existe solo en v3' END
  FROM v3_lot_metrics v3
  FULL OUTER JOIN v4_report.lot_metrics v4 ON v3.lot_id = v4.lot_id
  WHERE v3.lot_id IS NULL OR v4.lot_id IS NULL;

  -- CHECK 4: NULL_MISMATCH
  INSERT INTO parity_results
  SELECT 'NULL_MISMATCH', v3.lot_id, col_name,
    'v3_is_null=' || v3_null::text || ' v4_is_null=' || v4_null::text
  FROM v3_lot_metrics v3
  JOIN v4_report.lot_metrics v4 ON v3.lot_id = v4.lot_id
  CROSS JOIN LATERAL (
    VALUES 
      ('hectares', v3.hectares IS NULL, v4.hectares IS NULL),
      ('sowed_area_ha', v3.sowed_area_ha IS NULL, v4.sowed_area_ha IS NULL),
      ('labor_cost_usd', v3.labor_cost_usd IS NULL, v4.labor_cost_usd IS NULL),
      ('supplies_cost_usd', v3.supplies_cost_usd IS NULL, v4.supplies_cost_usd IS NULL),
      ('direct_cost_usd', v3.direct_cost_usd IS NULL, v4.direct_cost_usd IS NULL),
      ('income_net_total_usd', v3.income_net_total_usd IS NULL, v4.income_net_total_usd IS NULL),
      ('rent_per_ha_usd', v3.rent_per_ha_usd IS NULL, v4.rent_per_ha_usd IS NULL),
      ('admin_cost_per_ha_usd', v3.admin_cost_per_ha_usd IS NULL, v4.admin_cost_per_ha_usd IS NULL)
  ) AS nulls(col_name, v3_null, v4_null)
  WHERE v3_null <> v4_null;

  -- CHECK 5: Diferencias numéricas (todos los valores se castean a numeric para ROUND)
  INSERT INTO parity_results
  SELECT 'NUMERIC_DIFF', v3.lot_id, diff_col,
    'v3=' || ROUND(v3_val::numeric, 4)::text || ' v4=' || ROUND(v4_val::numeric, 4)::text || ' diff=' || ROUND(diff_val::numeric, 4)::text
  FROM v3_lot_metrics v3
  JOIN v4_report.lot_metrics v4 ON v3.lot_id = v4.lot_id
  CROSS JOIN LATERAL (
    VALUES 
      ('hectares', COALESCE(v3.hectares,0)::numeric, COALESCE(v4.hectares,0)::numeric, ABS(COALESCE(v3.hectares,0)::numeric - COALESCE(v4.hectares,0)::numeric)),
      ('sowed_area_ha', COALESCE(v3.sowed_area_ha,0)::numeric, COALESCE(v4.sowed_area_ha,0)::numeric, ABS(COALESCE(v3.sowed_area_ha,0)::numeric - COALESCE(v4.sowed_area_ha,0)::numeric)),
      ('harvested_area_ha', COALESCE(v3.harvested_area_ha,0)::numeric, COALESCE(v4.harvested_area_ha,0)::numeric, ABS(COALESCE(v3.harvested_area_ha,0)::numeric - COALESCE(v4.harvested_area_ha,0)::numeric)),
      ('yield_tn_per_ha', COALESCE(v3.yield_tn_per_ha,0)::numeric, COALESCE(v4.yield_tn_per_ha,0)::numeric, ABS(COALESCE(v3.yield_tn_per_ha,0)::numeric - COALESCE(v4.yield_tn_per_ha,0)::numeric)),
      ('labor_cost_usd', COALESCE(v3.labor_cost_usd,0)::numeric, COALESCE(v4.labor_cost_usd,0)::numeric, ABS(COALESCE(v3.labor_cost_usd,0)::numeric - COALESCE(v4.labor_cost_usd,0)::numeric)),
      ('supplies_cost_usd', COALESCE(v3.supplies_cost_usd,0)::numeric, COALESCE(v4.supplies_cost_usd,0)::numeric, ABS(COALESCE(v3.supplies_cost_usd,0)::numeric - COALESCE(v4.supplies_cost_usd,0)::numeric)),
      ('direct_cost_usd', COALESCE(v3.direct_cost_usd,0)::numeric, COALESCE(v4.direct_cost_usd,0)::numeric, ABS(COALESCE(v3.direct_cost_usd,0)::numeric - COALESCE(v4.direct_cost_usd,0)::numeric)),
      ('income_net_total_usd', COALESCE(v3.income_net_total_usd,0)::numeric, COALESCE(v4.income_net_total_usd,0)::numeric, ABS(COALESCE(v3.income_net_total_usd,0)::numeric - COALESCE(v4.income_net_total_usd,0)::numeric)),
      ('income_net_per_ha_usd', COALESCE(v3.income_net_per_ha_usd,0)::numeric, COALESCE(v4.income_net_per_ha_usd,0)::numeric, ABS(COALESCE(v3.income_net_per_ha_usd,0)::numeric - COALESCE(v4.income_net_per_ha_usd,0)::numeric)),
      ('rent_per_ha_usd', COALESCE(v3.rent_per_ha_usd,0)::numeric, COALESCE(v4.rent_per_ha_usd,0)::numeric, ABS(COALESCE(v3.rent_per_ha_usd,0)::numeric - COALESCE(v4.rent_per_ha_usd,0)::numeric)),
      ('admin_cost_per_ha_usd', COALESCE(v3.admin_cost_per_ha_usd,0)::numeric, COALESCE(v4.admin_cost_per_ha_usd,0)::numeric, ABS(COALESCE(v3.admin_cost_per_ha_usd,0)::numeric - COALESCE(v4.admin_cost_per_ha_usd,0)::numeric)),
      ('active_total_per_ha_usd', COALESCE(v3.active_total_per_ha_usd,0)::numeric, COALESCE(v4.active_total_per_ha_usd,0)::numeric, ABS(COALESCE(v3.active_total_per_ha_usd,0)::numeric - COALESCE(v4.active_total_per_ha_usd,0)::numeric)),
      ('operating_result_per_ha_usd', COALESCE(v3.operating_result_per_ha_usd,0)::numeric, COALESCE(v4.operating_result_per_ha_usd,0)::numeric, ABS(COALESCE(v3.operating_result_per_ha_usd,0)::numeric - COALESCE(v4.operating_result_per_ha_usd,0)::numeric)),
      ('rent_total_usd', COALESCE(v3.rent_total_usd,0)::numeric, COALESCE(v4.rent_total_usd,0)::numeric, ABS(COALESCE(v3.rent_total_usd,0)::numeric - COALESCE(v4.rent_total_usd,0)::numeric)),
      ('admin_total_usd', COALESCE(v3.admin_total_usd,0)::numeric, COALESCE(v4.admin_total_usd,0)::numeric, ABS(COALESCE(v3.admin_total_usd,0)::numeric - COALESCE(v4.admin_total_usd,0)::numeric)),
      ('active_total_usd', COALESCE(v3.active_total_usd,0)::numeric, COALESCE(v4.active_total_usd,0)::numeric, ABS(COALESCE(v3.active_total_usd,0)::numeric - COALESCE(v4.active_total_usd,0)::numeric)),
      ('operating_result_total_usd', COALESCE(v3.operating_result_total_usd,0)::numeric, COALESCE(v4.operating_result_total_usd,0)::numeric, ABS(COALESCE(v3.operating_result_total_usd,0)::numeric - COALESCE(v4.operating_result_total_usd,0)::numeric)),
      ('direct_cost_per_ha_usd', COALESCE(v3.direct_cost_per_ha_usd,0)::numeric, COALESCE(v4.direct_cost_per_ha_usd,0)::numeric, ABS(COALESCE(v3.direct_cost_per_ha_usd,0)::numeric - COALESCE(v4.direct_cost_per_ha_usd,0)::numeric)),
      ('project_total_hectares', COALESCE(v3.project_total_hectares,0)::numeric, COALESCE(v4.project_total_hectares,0)::numeric, ABS(COALESCE(v3.project_total_hectares,0)::numeric - COALESCE(v4.project_total_hectares,0)::numeric)),
      ('field_total_hectares', COALESCE(v3.field_total_hectares,0)::numeric, COALESCE(v4.field_total_hectares,0)::numeric, ABS(COALESCE(v3.field_total_hectares,0)::numeric - COALESCE(v4.field_total_hectares,0)::numeric))
  ) AS diffs(diff_col, v3_val, v4_val, diff_val)
  WHERE diff_val > 0.01;

  SELECT COUNT(*) INTO v_errors FROM parity_results;
  
  IF v_errors > 0 THEN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'PARIDAD FALLIDA: % errores', v_errors;
    RAISE NOTICE '========================================';
    FOR r IN SELECT * FROM parity_results LOOP
      RAISE NOTICE '[%] lot_id=% col=% msg=%', r.check_type, r.lot_id, r.column_name, r.message;
    END LOOP;
    RAISE EXCEPTION 'PARIDAD FALLIDA: % errores. Ver detalles arriba.', v_errors;
  ELSE
    RAISE NOTICE 'PARIDAD OK: v3_lot_metrics = v4_report.lot_metrics';
  END IF;
END $$;
