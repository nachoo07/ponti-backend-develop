-- =============================================================================
-- PARIDAD: v3_labor_list vs v4_report.labor_list
-- Ejecutar: psql -v ON_ERROR_STOP=1 -d ponti -f parity_v3_labor_list.sql
-- =============================================================================

DO $$
DECLARE
  v_errors int;
  r record;
BEGIN
  CREATE TEMP TABLE parity_results (
    check_type text,
    workorder_id bigint,
    column_name text,
    message text
  ) ON COMMIT DROP;

  -- CHECK 1: Unicidad v4
  INSERT INTO parity_results
  SELECT 'UNICIDAD_V4', NULL, NULL,
    'FAIL: ' || (COUNT(*) - COUNT(DISTINCT workorder_id))::text || ' duplicados'
  FROM v4_report.labor_list
  HAVING COUNT(*) <> COUNT(DISTINCT workorder_id);

  -- CHECK 2: Rowcount
  INSERT INTO parity_results
  SELECT 'ROWCOUNT', NULL, NULL,
    'FAIL: v3=' || v3_count::text || ' v4=' || v4_count::text
  FROM (
    SELECT 
      (SELECT COUNT(*) FROM v3_labor_list) AS v3_count,
      (SELECT COUNT(*) FROM v4_report.labor_list) AS v4_count
  ) c
  WHERE v3_count <> v4_count;

  -- CHECK 3: Keys faltantes
  INSERT INTO parity_results
  SELECT 
    CASE WHEN v3.workorder_id IS NULL THEN 'ONLY_V4' ELSE 'ONLY_V3' END,
    COALESCE(v3.workorder_id, v4.workorder_id), NULL,
    CASE WHEN v3.workorder_id IS NULL THEN 'Existe solo en v4' ELSE 'Existe solo en v3' END
  FROM v3_labor_list v3
  FULL OUTER JOIN v4_report.labor_list v4 ON v3.workorder_id = v4.workorder_id
  WHERE v3.workorder_id IS NULL OR v4.workorder_id IS NULL;

  -- CHECK 4: NULL_MISMATCH
  INSERT INTO parity_results
  SELECT 'NULL_MISMATCH', v3.workorder_id, col_name,
    'v3_is_null=' || v3_null::text || ' v4_is_null=' || v4_null::text
  FROM v3_labor_list v3
  JOIN v4_report.labor_list v4 ON v3.workorder_id = v4.workorder_id
  CROSS JOIN LATERAL (
    VALUES 
      ('lot_id', v3.lot_id IS NULL, v4.lot_id IS NULL),
      ('lot_name', v3.lot_name IS NULL, v4.lot_name IS NULL),
      ('crop_id', v3.crop_id IS NULL, v4.crop_id IS NULL),
      ('crop_name', v3.crop_name IS NULL, v4.crop_name IS NULL),
      ('labor_category_id', v3.labor_category_id IS NULL, v4.labor_category_id IS NULL),
      ('labor_category_name', v3.labor_category_name IS NULL, v4.labor_category_name IS NULL),
      ('investor_id', v3.investor_id IS NULL, v4.investor_id IS NULL),
      ('investor_name', v3.investor_name IS NULL, v4.investor_name IS NULL),
      ('surface_ha', v3.surface_ha IS NULL, v4.surface_ha IS NULL),
      ('cost_per_ha', v3.cost_per_ha IS NULL, v4.cost_per_ha IS NULL)
  ) AS nulls(col_name, v3_null, v4_null)
  WHERE v3_null <> v4_null;

  -- CHECK 5: Diferencias numéricas
  INSERT INTO parity_results
  SELECT 'NUMERIC_DIFF', v3.workorder_id, diff_col,
    'v3=' || ROUND(v3_val, 4)::text || ' v4=' || ROUND(v4_val, 4)::text || ' diff=' || ROUND(diff_val, 4)::text
  FROM v3_labor_list v3
  JOIN v4_report.labor_list v4 ON v3.workorder_id = v4.workorder_id
  CROSS JOIN LATERAL (
    VALUES 
      ('surface_ha', COALESCE(v3.surface_ha,0)::numeric, COALESCE(v4.surface_ha,0)::numeric, ABS(COALESCE(v3.surface_ha,0) - COALESCE(v4.surface_ha,0))),
      ('cost_per_ha', COALESCE(v3.cost_per_ha,0)::numeric, COALESCE(v4.cost_per_ha,0)::numeric, ABS(COALESCE(v3.cost_per_ha,0) - COALESCE(v4.cost_per_ha,0))),
      ('total_labor_cost', COALESCE(v3.total_labor_cost,0)::numeric, COALESCE(v4.total_labor_cost,0)::numeric, ABS(COALESCE(v3.total_labor_cost,0) - COALESCE(v4.total_labor_cost,0))),
      ('dollar_average_month', COALESCE(v3.dollar_average_month,0)::numeric, COALESCE(v4.dollar_average_month,0)::numeric, ABS(COALESCE(v3.dollar_average_month,0) - COALESCE(v4.dollar_average_month,0))),
      ('usd_cost_ha', COALESCE(v3.usd_cost_ha,0)::numeric, COALESCE(v4.usd_cost_ha,0)::numeric, ABS(COALESCE(v3.usd_cost_ha,0) - COALESCE(v4.usd_cost_ha,0))),
      ('usd_net_total', COALESCE(v3.usd_net_total,0)::numeric, COALESCE(v4.usd_net_total,0)::numeric, ABS(COALESCE(v3.usd_net_total,0) - COALESCE(v4.usd_net_total,0)))
  ) AS diffs(diff_col, v3_val, v4_val, diff_val)
  WHERE diff_val > 0.01;

  SELECT COUNT(*) INTO v_errors FROM parity_results;
  
  IF v_errors > 0 THEN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'PARIDAD FALLIDA: % errores', v_errors;
    RAISE NOTICE '========================================';
    FOR r IN SELECT * FROM parity_results LOOP
      RAISE NOTICE '[%] workorder_id=% col=% msg=%', r.check_type, r.workorder_id, r.column_name, r.message;
    END LOOP;
    RAISE EXCEPTION 'PARIDAD FALLIDA: % errores. Ver detalles arriba.', v_errors;
  ELSE
    RAISE NOTICE 'PARIDAD OK: v3_labor_list = v4_report.labor_list';
  END IF;
END $$;
