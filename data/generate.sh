#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="$ROOT/data"

if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required but not installed" >&2
  exit 1
fi

build_db() {
  local sql_file="$1"
  local db_file="$2"
  rm -f "$db_file"
  sqlite3 "$db_file" < "$sql_file"
  echo "generated $db_file"
}

build_db "$DATA_DIR/demo.sql"     "$DATA_DIR/demo.db"
build_db "$DATA_DIR/staging.sql"  "$DATA_DIR/staging.sqlite3"
build_db "$DATA_DIR/analytics.sql" "$DATA_DIR/analytics.db"
build_db "$DATA_DIR/large.sql"    "$DATA_DIR/large.db"

echo "done — open with: make dev -- --db $DATA_DIR/demo.db"
echo "       large db:  make dev -- --db $DATA_DIR/large.db"
