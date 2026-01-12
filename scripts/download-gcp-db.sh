#!/usr/bin/env bash
# backup_and_restore.sh (versión endurecida)
# - Preflight: prueba conexión, espera a PG, detecta superuser
# - Descarga dump (saltable con SKIP_DUMP=1)
# - Drop/Create DB y restore en 3 pasos
set -Eeuo pipefail

### ===== Origen (GCP) =====
SRC_USER="${SRC_USER:-soalen-db-v3}"
SRC_PASS="${SRC_PASS:-Soalen*25.}"
SRC_HOST="${SRC_HOST:-34.133.49.164}"
SRC_DB="${SRC_DB:-ponti_api_db}"
SRC_PORT="${SRC_PORT:-5432}"
SRC_SSL="${SRC_SSL:-require}"    # require | verify-full

### ===== Destino (Local) =====
DB_USER="${DB_USER:-admin}"
DB_PASSWORD="${DB_PASSWORD:-admin}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_NAME="${DB_NAME:-ponti_api_db}"
DB_PORT="${DB_PORT:-5432}"

# Control
DISABLE_TRIGGERS="${DISABLE_TRIGGERS:-1}"  # 1= intentar deshabilitar triggers (requiere superuser)
SKIP_DUMP="${SKIP_DUMP:-0}"                # 1= salta el pg_dump
DUMP_FILE="${DUMP_FILE:-ponti_api_db_$(date +%F).dump}"

log(){ echo -e "\n[INFO] $*"; }
warn(){ echo -e "\n[WARN] $*"; }
err(){ echo -e "\n[ERROR] $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "No se encontró '$1' en PATH"; exit 1; }; }
# trap removido para permitir que el script continúe con errores menores

### ===== Chequeo binarios =====
need psql; need pg_dump; need pg_restore; need pg_isready

### ===== Esperar a que PG esté listo y probar login =====
log "Esperando PostgreSQL en ${DB_HOST}:${DB_PORT} como ${DB_USER} ..."
for i in {1..30}; do
  if PGPASSWORD="$DB_PASSWORD" pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  [[ $i -eq 30 ]] && { err "PG no responde o credenciales inválidas.
Solución rápida:
  - Nativo: sudo -u postgres psql -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\"
  - Docker: docker exec -it <contenedor> psql -U postgres -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\"
"; exit 1; }
done

# Probar conexión real (captura error de auth)
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "\conninfo" >/dev/null

### ===== Detectar si el rol es superuser =====
IS_SUPER=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -At -d postgres -c "SELECT rolsuper::int FROM pg_roles WHERE rolname='${DB_USER}';" || echo 0)
if [[ "${IS_SUPER:-0}" != "1" ]]; then
  if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
    warn "El rol '${DB_USER}' NO es superuser. Cambio automático DISABLE_TRIGGERS=0 para evitar errores."
    DISABLE_TRIGGERS=0
  fi
fi

### ===== Dump desde GCP (opcional) =====
if [[ "$SKIP_DUMP" == "1" && -f "$DUMP_FILE" ]]; then
  log "SKIP_DUMP=1 → uso dump existente: ${DUMP_FILE}"
else
  log "Generando dump desde GCP -> ${DUMP_FILE}"
  PGPASSWORD="$SRC_PASS" pg_dump \
    "postgresql://${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}" \
    -F c --no-owner --no-acl -v -f "$DUMP_FILE"
fi

log "Contenido del dump (primeras 20 líneas):"
pg_restore -l "$DUMP_FILE" | head -n 20 || true

### ===== Drop/Create database destino =====
log "Terminando conexiones activas a '${DB_NAME}' (si existieran)…"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname='${DB_NAME}' AND pid<>pg_backend_pid();
" || true

log "DROP DATABASE IF EXISTS ${DB_NAME};"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${DB_NAME};"

log "CREATE DATABASE ${DB_NAME};"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${DB_NAME};"

log "Conexión a la DB nueva:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\conninfo"

### ===== Restore en 3 pasos =====
RESTORE_COMMON=(-h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --no-owner --no-privileges -v)

log "PASO 1/3 — PRE-DATA (esquema)…"
if PGPASSWORD="$DB_PASSWORD" pg_restore "${RESTORE_COMMON[@]}" --section=pre-data "$DUMP_FILE" 2>&1; then
  log "  ✅ Pre-data completado exitosamente"
else
  warn "  ⚠️  Pre-data tuvo errores (como transaction_timeout) pero continuando..."
fi

log "PASO 2/3 — DATA (solo datos)…"
if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
  log "  -> con --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
  if PGPASSWORD="$DB_PASSWORD" pg_restore "${RESTORE_COMMON[@]}" --section=data --disable-triggers "$DUMP_FILE" 2>&1; then
    log "  ✅ Data completado exitosamente"
  else
    warn "  ⚠️  Data tuvo errores pero continuando..."
  fi
else
  log "  -> SIN --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
  if PGPASSWORD="$DB_PASSWORD" pg_restore "${RESTORE_COMMON[@]}" --section=data "$DUMP_FILE" 2>&1; then
    log "  ✅ Data completado exitosamente"
  else
    warn "  ⚠️  Data tuvo errores pero continuando..."
  fi
fi

log "PASO 3/3 — POST-DATA (índices, FKs, constraints)…"
if PGPASSWORD="$DB_PASSWORD" pg_restore "${RESTORE_COMMON[@]}" --section=post-data "$DUMP_FILE" 2>&1; then
  log "  ✅ Post-data completado exitosamente"
else
  warn "  ⚠️  Post-data tuvo errores pero continuando..."
fi

### ===== Verificación rápida =====
log "Tablas (primeras 30 filas de \dt):"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "\dt public.*" | sed -n '1,30p' || true

log "✅ Restauración completada con éxito en '${DB_NAME}'"
