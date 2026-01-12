#!/bin/bash
# ========================================
# SCRIPT: Dump completo del schema (estructura)
# ========================================
# 
# Propósito: Generar un archivo SQL único con TODO el schema
# para recrear la BD desde cero (sin datos)
# 
# Uso: ./create_full_schema_dump.sh [nombre]
# Ejemplo: ./create_full_schema_dump.sh schema_completo

set -e

SNAPSHOT_NAME="${1:-schema_completo_$(date +%Y%m%d_%H%M%S)}"
OUTPUT_FILE="snapshots/${SNAPSHOT_NAME}.sql"

mkdir -p snapshots

echo "📸 Generando dump completo del schema: ${SNAPSHOT_NAME}"
echo "📁 Archivo de salida: ${OUTPUT_FILE}"

# Crear dump completo del schema (sin datos)
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    pg_dump -U admin -d ponti_api_db \
    --schema-only \
    --no-owner \
    --no-privileges \
    --verbose \
    > "${OUTPUT_FILE}"

# Verificar que el archivo se creó correctamente
if [ ! -s "${OUTPUT_FILE}" ]; then
    echo "❌ Error: El archivo está vacío"
    exit 1
fi

# Agregar header informativo al inicio del archivo
HEADER=$(cat << 'EOF'
-- ========================================
-- DUMP COMPLETO DEL SCHEMA
-- ========================================
-- 
-- Este archivo contiene TODO el schema de la base de datos:
--   - Schemas (v3_calc, v3_core_ssot, etc.)
--   - Tablas (con columnas, tipos, defaults)
--   - Constraints (PK, FK, UNIQUE, CHECK)
--   - Índices
--   - Vistas (v3_*)
--   - Funciones SSOT (157 funciones)
--   - Tipos (enums, tipos custom)
--   - Secuencias
--   - Triggers
-- 
-- NO contiene datos (filas/INSERT)
-- 
-- Para restaurar: psql -U admin -d ponti_api_db < este_archivo.sql
-- 
-- Generado: TIMESTAMP_PLACEHOLDER
-- ========================================

EOF
)

# Reemplazar placeholder con timestamp real
HEADER=$(echo "$HEADER" | sed "s/TIMESTAMP_PLACEHOLDER/$(date +"%Y-%m-%d %H:%M:%S")/")

# Crear archivo temporal con header + contenido
TEMP_FILE="${OUTPUT_FILE}.tmp"
echo "$HEADER" > "${TEMP_FILE}"
cat "${OUTPUT_FILE}" >> "${TEMP_FILE}"
mv "${TEMP_FILE}" "${OUTPUT_FILE}"

# Estadísticas del archivo
FILE_SIZE=$(du -h "${OUTPUT_FILE}" | cut -f1)
LINE_COUNT=$(wc -l < "${OUTPUT_FILE}")

echo ""
echo "✅ Dump completado exitosamente"
echo "📊 Estadísticas:"
echo "   - Archivo: ${OUTPUT_FILE}"
echo "   - Tamaño: ${FILE_SIZE}"
echo "   - Líneas: ${LINE_COUNT}"
echo ""
echo "📋 Primeras líneas del archivo:"
head -n 30 "${OUTPUT_FILE}" | head -n 20
echo "..."
echo ""
echo "🔍 Para verificar contenido:"
echo "   head -n 50 ${OUTPUT_FILE}"
echo "   grep -E '^(CREATE|ALTER)' ${OUTPUT_FILE} | head -n 20"





