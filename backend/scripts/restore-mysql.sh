#!/usr/bin/env bash
# Restores a backup created by backup-mysql.sh. This overwrites database data.
# Required environment: MYSQL_HOST, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE,
# BACKUP_FILE and RESTORE_CONFIRM (must equal MYSQL_DATABASE).
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
require_env BACKUP_FILE
require_env RESTORE_CONFIRM

MYSQL_PORT="${MYSQL_PORT:-3306}"
if ! [[ "$MYSQL_PORT" =~ ^[0-9]+$ ]] || (( MYSQL_PORT < 1 || MYSQL_PORT > 65535 )); then
  echo "MYSQL_PORT must be between 1 and 65535" >&2
  exit 2
fi
if [[ "$MYSQL_DATABASE" =~ [^A-Za-z0-9_] ]]; then
  echo "MYSQL_DATABASE may contain only letters, digits and underscores" >&2
  exit 2
fi
if [[ "$RESTORE_CONFIRM" != "$MYSQL_DATABASE" ]]; then
  echo "RESTORE_CONFIRM must exactly match MYSQL_DATABASE" >&2
  exit 2
fi
if [[ ! -f "$BACKUP_FILE" || ( "$BACKUP_FILE" != *.sql && "$BACKUP_FILE" != *.sql.gz ) ]]; then
  echo "BACKUP_FILE must be an existing .sql or .sql.gz file" >&2
  exit 2
fi

command -v mysql >/dev/null || { echo "mysql client is required" >&2; exit 127; }
command -v sha256sum >/dev/null || { echo "sha256sum is required" >&2; exit 127; }
if [[ "$BACKUP_FILE" == *.sql.gz ]]; then
  command -v gzip >/dev/null || { echo "gzip is required" >&2; exit 127; }
fi

checksum_file="${BACKUP_FILE}.sha256"
if [[ -f "$checksum_file" ]]; then
  (cd "$(dirname "$BACKUP_FILE")" && sha256sum --check "$(basename "$checksum_file")")
fi

mysql_args=(--protocol=TCP --host="$MYSQL_HOST" --port="$MYSQL_PORT" --user="$MYSQL_USER")
MYSQL_PWD="$MYSQL_PASSWORD" mysql "${mysql_args[@]}" -e "CREATE DATABASE IF NOT EXISTS \`${MYSQL_DATABASE}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
table_count="$(MYSQL_PWD="$MYSQL_PASSWORD" mysql "${mysql_args[@]}" --skip-column-names --batch -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${MYSQL_DATABASE}'")"
if [[ "$table_count" != "0" && "${RESTORE_ALLOW_NONEMPTY:-0}" != "1" ]]; then
  echo "target database is not empty; use a fresh drill database or set RESTORE_ALLOW_NONEMPTY=1 after review" >&2
  exit 2
fi
if [[ "$BACKUP_FILE" == *.sql.gz ]]; then
  gzip -cd "$BACKUP_FILE" | MYSQL_PWD="$MYSQL_PASSWORD" mysql "${mysql_args[@]}" --database="$MYSQL_DATABASE"
else
  MYSQL_PWD="$MYSQL_PASSWORD" mysql "${mysql_args[@]}" --database="$MYSQL_DATABASE" < "$BACKUP_FILE"
fi

echo "restore completed for database: ${MYSQL_DATABASE}"
