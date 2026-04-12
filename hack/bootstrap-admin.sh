#!/usr/bin/env bash
set -euo pipefail

if [ -z "${SUPER_ADMIN_PHONE:-}" ]; then
  echo "SUPER_ADMIN_PHONE is required" >&2
  exit 1
fi
if [ -z "${SUPER_ADMIN_BCRYPT_HASH:-}" ]; then
  echo "SUPER_ADMIN_BCRYPT_HASH is required" >&2
  exit 1
fi

sed \
  -e "s#{{SUPER_ADMIN_PHONE}}#${SUPER_ADMIN_PHONE}#g" \
  -e "s#{{SUPER_ADMIN_BCRYPT_HASH}}#${SUPER_ADMIN_BCRYPT_HASH}#g" \
  manifest/sql/003_seed_admin.sql.tmpl > manifest/sql/003_seed_admin.sql

echo "generated manifest/sql/003_seed_admin.sql"
