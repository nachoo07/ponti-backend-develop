-- =============================================================================
-- SANITY CHECK: Verificar que v4_report.lot_metrics es ejecutable
-- Ejecutar: psql -d ponti -f scripts/validation/sanity_explain_v4_lot_metrics.sql
-- Propósito: Detectar errores de sintaxis, dependencias rotas, ciclos
-- =============================================================================

-- Mostrar plan de ejecución (sin costos para legibilidad)
\echo '>>> EXPLAIN v4_report.lot_metrics'
EXPLAIN (COSTS OFF, FORMAT TEXT)
SELECT * FROM v4_report.lot_metrics LIMIT 1;

-- Verificar que la vista es ejecutable
\echo ''
\echo '>>> Verificando ejecución real (LIMIT 1)'
SELECT 1 AS sanity_ok FROM v4_report.lot_metrics LIMIT 1;

-- Mostrar columnas resultantes
\echo ''
\echo '>>> Columnas de v4_report.lot_metrics'
SELECT 
  ordinal_position AS pos,
  column_name,
  data_type
FROM information_schema.columns
WHERE table_schema = 'v4_report' 
  AND table_name = 'lot_metrics'
ORDER BY ordinal_position;
