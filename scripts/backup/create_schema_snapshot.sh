#!/bin/bash
# ========================================
# SCRIPT: Crear snapshot completo del schema
# ========================================
# 
# Propósito: Generar backup completo del schema antes de cambios
# Uso: ./create_schema_snapshot.sh [nombre_snapshot]
# 
# Ejemplo: ./create_schema_snapshot.sh antes_fase_1

set -e

# Configuración
SNAPSHOT_NAME="${1:-snapshot_$(date +%Y%m%d_%H%M%S)}"
OUTPUT_DIR="snapshots/${SNAPSHOT_NAME}"
DB_NAME="ponti_api_db"
DB_USER="admin"
DB_HOST="localhost"
DB_PORT="5432"

# Crear directorio de salida
mkdir -p "${OUTPUT_DIR}"

echo "📸 Creando snapshot del schema: ${SNAPSHOT_NAME}"
echo "📁 Directorio de salida: ${OUTPUT_DIR}"

# 1. Schema completo (solo estructura, sin datos)
echo "📋 1. Exportando schema completo..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    pg_dump -U ${DB_USER} -d ${DB_NAME} \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "${OUTPUT_DIR}/01_schema_completo.sql"

# 2. Solo schemas SSOT
echo "📋 2. Exportando schemas SSOT..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    pg_dump -U ${DB_USER} -d ${DB_NAME} \
    --schema=v3_calc \
    --schema=v3_core_ssot \
    --schema=v3_lot_ssot \
    --schema=v3_dashboard_ssot \
    --schema=v3_report_ssot \
    --schema=v3_workorder_ssot \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "${OUTPUT_DIR}/02_schemas_ssot.sql"

# 3. Solo vistas v3_*
echo "📋 3. Exportando vistas v3_*..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    psql -U ${DB_USER} -d ${DB_NAME} -t -c "
    SELECT 'CREATE OR REPLACE VIEW ' || schemaname || '.' || viewname || ' AS ' || 
           pg_get_viewdef(schemaname || '.' || viewname, true) || ';'
    FROM pg_views
    WHERE schemaname = 'public' AND viewname LIKE 'v3_%'
    ORDER BY viewname;
    " > "${OUTPUT_DIR}/03_vistas_v3.sql"

# 4. Lista de funciones SSOT
echo "📋 4. Exportando definiciones de funciones SSOT..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    psql -U ${DB_USER} -d ${DB_NAME} -t -c "
    SELECT pg_get_functiondef(p.oid) || ';'
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname IN ('v3_calc', 'v3_core_ssot', 'v3_lot_ssot', 'v3_dashboard_ssot', 'v3_report_ssot', 'v3_workorder_ssot')
    ORDER BY n.nspname, p.proname;
    " > "${OUTPUT_DIR}/04_funciones_ssot.sql"

# 5. Inventario de objetos
echo "📋 5. Creando inventario de objetos..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    psql -U ${DB_USER} -d ${DB_NAME} << 'EOF' > "${OUTPUT_DIR}/05_inventario_objetos.txt"
-- Esquemas
SELECT 'SCHEMA: ' || nspname FROM pg_namespace WHERE nspname LIKE 'v3_%' ORDER BY nspname;

-- Funciones por esquema
SELECT 'FUNCTION: ' || n.nspname || '.' || p.proname || '(' || pg_get_function_arguments(p.oid) || ')'
FROM pg_proc p
JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname LIKE 'v3_%'
ORDER BY n.nspname, p.proname;

-- Vistas
SELECT 'VIEW: ' || schemaname || '.' || viewname
FROM pg_views
WHERE schemaname = 'public' AND viewname LIKE 'v3_%'
ORDER BY viewname;
EOF

# 6. Metadata adicional
echo "📋 6. Exportando metadata..."
docker compose -f projects/ponti-api/docker-compose.yml exec -T ponti-db \
    psql -U ${DB_USER} -d ${DB_NAME} << 'EOF' > "${OUTPUT_DIR}/06_metadata.txt"
-- Versión de PostgreSQL
SELECT version();

-- Fecha y hora del snapshot
SELECT NOW() AS snapshot_date;

-- Tamaño de la base de datos
SELECT pg_size_pretty(pg_database_size('ponti_api_db')) AS database_size;

-- Número de objetos por tipo
SELECT 
    'Schemas SSOT' AS tipo,
    COUNT(*) AS cantidad
FROM pg_namespace WHERE nspname LIKE 'v3_%'
UNION ALL
SELECT 
    'Funciones SSOT',
    COUNT(*)
FROM pg_proc p
JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname LIKE 'v3_%'
UNION ALL
SELECT 
    'Vistas v3_*',
    COUNT(*)
FROM pg_views
WHERE schemaname = 'public' AND viewname LIKE 'v3_%';
EOF

# 7. Crear README del snapshot
cat > "${OUTPUT_DIR}/README.md" << EOF
# 📸 Snapshot del Schema: ${SNAPSHOT_NAME}

**Fecha:** $(date +"%Y-%m-%d %H:%M:%S")
**Base de datos:** ${DB_NAME}
**Propósito:** Backup del schema antes de cambios

## 📁 Archivos incluidos

1. **01_schema_completo.sql** - Schema completo de la BD (solo estructura)
2. **02_schemas_ssot.sql** - Solo schemas SSOT (v3_calc, v3_core_ssot, etc.)
3. **03_vistas_v3.sql** - Solo vistas v3_*
4. **04_funciones_ssot.sql** - Definiciones completas de funciones SSOT
5. **05_inventario_objetos.txt** - Lista de todos los objetos
6. **06_metadata.txt** - Información de versión y estadísticas
7. **README.md** - Este archivo

## 🔄 Restaurar snapshot

\`\`\`bash
# Restaurar schema completo
docker compose exec ponti-db psql -U admin -d ponti_api_db < 01_schema_completo.sql

# O restaurar solo schemas SSOT
docker compose exec ponti-db psql -U admin -d ponti_api_db < 02_schemas_ssot.sql
\`\`\`

## ⚠️ Notas

- Este snapshot contiene **solo estructura**, NO datos
- Para restaurar datos, usar backup completo de la BD
- Verificar que no haya conflictos antes de restaurar
EOF

echo ""
echo "✅ Snapshot creado exitosamente en: ${OUTPUT_DIR}"
echo "📊 Archivos generados:"
ls -lh "${OUTPUT_DIR}" | tail -n +2





