#!/bin/bash
# ========================================
# SCRIPT SIMPLE: Snapshot rápido del schema
# ========================================
# 
# Versión simplificada que solo exporta lo esencial

set -e

SNAPSHOT_NAME="${1:-snapshot_$(date +%Y%m%d_%H%M%S)}"
OUTPUT_FILE="snapshots/${SNAPSHOT_NAME}.sql"

mkdir -p snapshots

echo "📸 Creando snapshot rápido: ${SNAPSHOT_NAME}"

# Exportar solo schemas SSOT y vistas v3_*
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    pg_dump -U admin -d ponti_api_db \
    --schema=v3_calc \
    --schema=v3_core_ssot \
    --schema=v3_lot_ssot \
    --schema=v3_dashboard_ssot \
    --schema=v3_report_ssot \
    --schema=v3_workorder_ssot \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "${OUTPUT_FILE}"

# Agregar vistas v3_* al final
echo "" >> "${OUTPUT_FILE}"
echo "-- ========================================" >> "${OUTPUT_FILE}"
echo "-- VISTAS v3_*" >> "${OUTPUT_FILE}"
echo "-- ========================================" >> "${OUTPUT_FILE}"

docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    psql -U admin -d ponti_api_db -t -c "
    SELECT 'CREATE OR REPLACE VIEW ' || schemaname || '.' || viewname || ' AS ' || 
           pg_get_viewdef(schemaname || '.' || viewname, true) || ';'
    FROM pg_views
    WHERE schemaname = 'public' AND viewname LIKE 'v3_%'
    ORDER BY viewname;
    " >> "${OUTPUT_FILE}"

echo "✅ Snapshot guardado en: ${OUTPUT_FILE}"
echo "📊 Tamaño: $(du -h ${OUTPUT_FILE} | cut -f1)"





