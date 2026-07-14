#!/usr/bin/env bash
# Creates an encrypted-permission MySQL logical backup for TIKU-ZONG.
# Required environment: MYSQL_HOST, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE.
set -Eeuo pipefail
umask 077

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "missing required environment variable: ${name}" >&2
    exit 2
  fi
}

require_env MYSQL_HOST
require_env MYSQL_USER
require_env MYSQL_PASSWORD
require_env MYSQL_DATABASE

MYSQL_PORT="${MYSQL_PORT:-3306}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/tiku-zong}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

if ! [[ "$MYSQL_PORT" =~ ^[0-9]+$ && "$BACKUP_RETENTION_DAYS" =~ ^[0-9]+$ ]] || (( MYSQL_PORT < 1 || MYSQL_PORT > 65535 )); then
  echo "MYSQL_PORT must be between 1 and 65535; BACKUP_RETENTION_DAYS must be a non-negative integer" >&2
  exit 2
fi
if [[ "$MYSQL_DATABASE" =~ [^A-Za-z0-9_] ]]; then
  echo "MYSQL_DATABASE may contain only letters, digits and underscores" >&2
  exit 2
fi
if [[ "$BACKUP_DIR" != /* || "$BACKUP_DIR" == "/" || "$BACKUP_DIR" == /www/wwwroot/* ]]; then
  echo "BACKUP_DIR must be an absolute path outside the web root and must not be /" >&2
  exit 2
fi

command -v mysqldump >/dev/null || { echo "mysqldump is required" >&2; exit 127; }
command -v gzip >/dev/null || { echo "gzip is required" >&2; exit 127; }
command -v sha256sum >/dev/null || { echo "sha256sum is required" >&2; exit 127; }

mkdir -p "$BACKUP_DIR"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
filename="${MYSQL_DATABASE}_${timestamp}.sql.gz"
target="${BACKUP_DIR%/}/${filename}"
temporary="${target}.tmp"

trap 'rm -f "$temporary"' EXIT
MYSQL_PWD="$MYSQL_PASSWORD" mysqldump \
  --protocol=TCP --host="$MYSQL_HOST" --port="$MYSQL_PORT" --user="$MYSQL_USER" \
  --single-transaction --routines --events --triggers --default-character-set=utf8mb4 \
  "$MYSQL_DATABASE" | gzip -c > "$temporary"
mv "$temporary" "$target"
(cd "$BACKUP_DIR" && sha256sum "$filename" > "${filename}.sha256")
find "$BACKUP_DIR" -maxdepth 1 -type f -name "${MYSQL_DATABASE}_*.sql.gz" -mtime "+${BACKUP_RETENTION_DAYS}" -delete
find "$BACKUP_DIR" -maxdepth 1 -type f -name "${MYSQL_DATABASE}_*.sql.gz.sha256" -mtime "+${BACKUP_RETENTION_DAYS}" -delete

echo "backup created: ${target}"
