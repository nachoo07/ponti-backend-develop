-- =============================================================================
-- CONTRACT: Validar que v3_lot_metrics y v4_report.lot_metrics tienen
--           las mismas columnas (nombre, posición, tipo)
-- Ejecutar: psql -v ON_ERROR_STOP=1 -d ponti -f parity_contract_v3_lot_metrics.sql
-- =============================================================================

DO $$
DECLARE
  v_errors int;
  r record;
BEGIN
  CREATE TEMP TABLE contract_diffs (
    check_type text,
    column_name text,
    v3_value text,
    v4_value text
  ) ON COMMIT DROP;

  -- CHECK 1: Columnas en v3 que no existen en v4
  INSERT INTO contract_diffs
  SELECT 
    'MISSING_IN_V4',
    v3.column_name,
    'exists (pos=' || v3.ordinal_position || ', type=' || v3.data_type || ')',
    'NOT FOUND'
  FROM information_schema.columns v3
  WHERE v3.table_schema = 'public' 
    AND v3.table_name = 'v3_lot_metrics'
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns v4
      WHERE v4.table_schema = 'v4_report'
        AND v4.table_name = 'lot_metrics'
        AND v4.column_name = v3.column_name
    );

  -- CHECK 2: Columnas en v4 que no existen en v3
  INSERT INTO contract_diffs
  SELECT 
    'MISSING_IN_V3',
    v4.column_name,
    'NOT FOUND',
    'exists (pos=' || v4.ordinal_position || ', type=' || v4.data_type || ')'
  FROM information_schema.columns v4
  WHERE v4.table_schema = 'v4_report' 
    AND v4.table_name = 'lot_metrics'
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns v3
      WHERE v3.table_schema = 'public'
        AND v3.table_name = 'v3_lot_metrics'
        AND v3.column_name = v4.column_name
    );

  -- CHECK 3: Columnas con posición diferente
  INSERT INTO contract_diffs
  SELECT 
    'POSITION_MISMATCH',
    v3.column_name,
    'pos=' || v3.ordinal_position::text,
    'pos=' || v4.ordinal_position::text
  FROM information_schema.columns v3
  JOIN information_schema.columns v4 
    ON v4.table_schema = 'v4_report'
    AND v4.table_name = 'lot_metrics'
    AND v4.column_name = v3.column_name
  WHERE v3.table_schema = 'public' 
    AND v3.table_name = 'v3_lot_metrics'
    AND v3.ordinal_position <> v4.ordinal_position;

  -- CHECK 4: Columnas con tipo diferente
  -- Solo permitir equivalencia integer<->bigint. numeric debe matchear numeric.
  INSERT INTO contract_diffs
  SELECT 
    'TYPE_MISMATCH',
    v3.column_name,
    'type=' || v3.data_type,
    'type=' || v4.data_type
  FROM information_schema.columns v3
  JOIN information_schema.columns v4 
    ON v4.table_schema = 'v4_report'
    AND v4.table_name = 'lot_metrics'
    AND v4.column_name = v3.column_name
  WHERE v3.table_schema = 'public' 
    AND v3.table_name = 'v3_lot_metrics'
    AND v3.data_type <> v4.data_type
    -- Solo permitir integer<->bigint como equivalentes
    AND NOT (v3.data_type IN ('integer', 'bigint') 
             AND v4.data_type IN ('integer', 'bigint'));

  SELECT COUNT(*) INTO v_errors FROM contract_diffs;
  
  IF v_errors > 0 THEN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'CONTRACT FALLIDO: % diferencias de schema', v_errors;
    RAISE NOTICE '========================================';
    FOR r IN SELECT * FROM contract_diffs ORDER BY check_type, column_name LOOP
      RAISE NOTICE '[%] col=% v3=% v4=%', r.check_type, r.column_name, r.v3_value, r.v4_value;
    END LOOP;
    RAISE EXCEPTION 'CONTRACT FALLIDO: v3_lot_metrics y v4_report.lot_metrics tienen schemas diferentes';
  ELSE
    RAISE NOTICE 'CONTRACT OK: v3_lot_metrics y v4_report.lot_metrics tienen el mismo schema';
  END IF;
END $$;
