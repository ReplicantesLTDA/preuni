#!/usr/bin/env bash
# Daily pg_dump at 03:00 UTC — uploads to S3-compatible storage.
# Retains 30 days of backups; older files deleted from S3.
set -euo pipefail

: "${POSTGRES_HOST:?POSTGRES_HOST required}"
: "${POSTGRES_USER:?POSTGRES_USER required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD required}"
: "${POSTGRES_DB:?POSTGRES_DB required}"
: "${BACKUP_S3_ENDPOINT:?BACKUP_S3_ENDPOINT required}"
: "${BACKUP_S3_ACCESS_KEY:?BACKUP_S3_ACCESS_KEY required}"
: "${BACKUP_S3_SECRET_KEY:?BACKUP_S3_SECRET_KEY required}"
: "${BACKUP_S3_BUCKET:?BACKUP_S3_BUCKET required}"

TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
BACKUP_FILE="/tmp/corretor-${TIMESTAMP}.dump"
RETENTION_DAYS=${RETENTION_DAYS:-30}

export PGPASSWORD="${POSTGRES_PASSWORD}"

echo "[backup] Starting pg_dump at ${TIMESTAMP}"
pg_dump \
    --host="${POSTGRES_HOST}" \
    --port="${POSTGRES_PORT:-5432}" \
    --username="${POSTGRES_USER}" \
    --dbname="${POSTGRES_DB}" \
    --format=custom \
    --file="${BACKUP_FILE}"

echo "[backup] Upload to s3://${BACKUP_S3_BUCKET}/backups/${TIMESTAMP}.dump"
s3cmd \
    --access_key="${BACKUP_S3_ACCESS_KEY}" \
    --secret_key="${BACKUP_S3_SECRET_KEY}" \
    --host="${BACKUP_S3_ENDPOINT}" \
    --host-bucket="%(bucket)s.${BACKUP_S3_ENDPOINT}" \
    put "${BACKUP_FILE}" "s3://${BACKUP_S3_BUCKET}/backups/${TIMESTAMP}.dump"

rm -f "${BACKUP_FILE}"

echo "[backup] Purging backups older than ${RETENTION_DAYS} days"
CUTOFF=$(date -u -d "${RETENTION_DAYS} days ago" +%Y%m%d 2>/dev/null || \
         date -u -v-"${RETENTION_DAYS}"d +%Y%m%d)

s3cmd \
    --access_key="${BACKUP_S3_ACCESS_KEY}" \
    --secret_key="${BACKUP_S3_SECRET_KEY}" \
    --host="${BACKUP_S3_ENDPOINT}" \
    --host-bucket="%(bucket)s.${BACKUP_S3_ENDPOINT}" \
    ls "s3://${BACKUP_S3_BUCKET}/backups/" \
  | awk '{print $4}' \
  | grep "\.dump$" \
  | while read -r key; do
      file_date=$(basename "${key}" .dump | grep -oP '\d{8}' | head -1)
      if [[ "${file_date}" < "${CUTOFF}" ]]; then
          echo "[backup] Deleting old backup: ${key}"
          s3cmd \
              --access_key="${BACKUP_S3_ACCESS_KEY}" \
              --secret_key="${BACKUP_S3_SECRET_KEY}" \
              --host="${BACKUP_S3_ENDPOINT}" \
              --host-bucket="%(bucket)s.${BACKUP_S3_ENDPOINT}" \
              del "${key}"
      fi
  done

echo "[backup] Done"
