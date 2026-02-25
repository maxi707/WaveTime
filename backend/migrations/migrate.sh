#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PG_BIN_DIR="${PG_BIN_DIR:-$HOME/ARENADATA/github/orioledb/output_bin/bin}"
PSQL_BIN="${PSQL_BIN:-$PG_BIN_DIR/psql}"

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-wavetime}"
PGDATABASE="${PGDATABASE:-wavetime}"

CONN_ARGS=( -w -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" )

usage() {
  cat <<EOF
Usage:
  $(basename "$0") <up|down|status|check>

Connection settings (env):
  PG_BIN_DIR  (default: $HOME/ARENADATA/github/orioledb/output_bin/bin)
  PSQL_BIN    (default: \$PG_BIN_DIR/psql)
  PGHOST      (default: localhost)
  PGPORT      (default: 5432)
  PGUSER      (default: wavetime)
  PGDATABASE  (default: wavetime)
EOF
}

require_psql() {
  if [[ ! -x "$PSQL_BIN" ]]; then
    echo "ERROR: psql not found or not executable: $PSQL_BIN" >&2
    exit 1
  fi
}

psql_exec() {
  "$PSQL_BIN" "${CONN_ARGS[@]}" -v ON_ERROR_STOP=1 "$@"
}

list_up_files() {
  find "$SCRIPT_DIR" -maxdepth 1 -type f -name '*.up.sql' | sort
}

list_down_files() {
  find "$SCRIPT_DIR" -maxdepth 1 -type f -name '*.down.sql' | sort -r
}

run_up() {
  local files
  mapfile -t files < <(list_up_files)
  if [[ ${#files[@]} -eq 0 ]]; then
    echo "No .up.sql files found in $SCRIPT_DIR"
    return 0
  fi

  echo "Applying ${#files[@]} up migration(s) to $PGUSER@$PGHOST:$PGPORT/$PGDATABASE"
  for file in "${files[@]}"; do
    echo "-> $(basename "$file")"
    psql_exec -f "$file"
  done
  echo "Up migrations completed"
}

run_down() {
  local files
  mapfile -t files < <(list_down_files)
  if [[ ${#files[@]} -eq 0 ]]; then
    echo "No .down.sql files found in $SCRIPT_DIR"
    return 0
  fi

  echo "Applying ${#files[@]} down migration(s) to $PGUSER@$PGHOST:$PGPORT/$PGDATABASE"
  for file in "${files[@]}"; do
    echo "-> $(basename "$file")"
    psql_exec -f "$file"
  done
  echo "Down migrations completed"
}

migration_status_raw() {
  psql_exec -Atc "
SELECT
  (
    EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'citext')
    AND EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE n.nspname='public' AND t.typname='user_role')
  ) AS m1,
  (
    SELECT COUNT(*) = 9
    FROM pg_tables
    WHERE schemaname='public'
      AND tablename IN ('users','pools','training_types','slots','bookings','payments','attendance','admin_actions','notification_queue')
  ) AS m2,
  (
    SELECT COUNT(*) = 10
    FROM pg_indexes
    WHERE schemaname='public'
      AND indexname IN (
        'idx_slots_pool_starts_at',
        'idx_slots_status_starts_at',
        'idx_bookings_user_status_created',
        'idx_bookings_slot_status',
        'idx_bookings_reserved_until',
        'idx_payments_booking_status',
        'idx_payments_created_at',
        'idx_admin_actions_entity_created',
        'idx_notification_queue_status_scheduled',
        'idx_notification_queue_user_created'
      )
  ) AS m3,
  (
    EXISTS (
      SELECT 1
      FROM pg_proc p
      JOIN pg_namespace n ON n.oid = p.pronamespace
      WHERE n.nspname='public' AND p.proname='sync_slot_booked_count'
    )
    AND EXISTS (
      SELECT 1
      FROM pg_trigger t
      JOIN pg_class c ON c.oid = t.tgrelid
      JOIN pg_namespace n ON n.oid = c.relnamespace
      WHERE n.nspname='public'
        AND c.relname='bookings'
        AND t.tgname='trg_bookings_sync_slot_booked_count'
    )
  ) AS m4;"
}

run_status() {
  local raw m1 m2 m3 m4
  raw="$(migration_status_raw)"
  IFS='|' read -r m1 m2 m3 m4 <<< "$raw"

  echo "Status for $PGUSER@$PGHOST:$PGPORT/$PGDATABASE"
  printf "  [000001] extensions_and_types         : %s\n" "$( [[ "$m1" == "t" ]] && echo APPLIED || echo PENDING )"
  printf "  [000002] core_tables                  : %s\n" "$( [[ "$m2" == "t" ]] && echo APPLIED || echo PENDING )"
  printf "  [000003] indexes                      : %s\n" "$( [[ "$m3" == "t" ]] && echo APPLIED || echo PENDING )"
  printf "  [000004] functions_and_triggers       : %s\n" "$( [[ "$m4" == "t" ]] && echo APPLIED || echo PENDING )"

  if [[ "$m1$m2$m3$m4" == "tttt" ]]; then
    echo "Overall: OK"
  else
    echo "Overall: INCOMPLETE"
    return 1
  fi
}

run_check() {
  echo "Running schema checks on $PGUSER@$PGHOST:$PGPORT/$PGDATABASE"

  psql_exec <<'SQL'
DO $$
DECLARE
  v_tables int;
  v_types int;
  v_funcs int;
  v_indexes int;
  v_triggers int;
BEGIN
  SELECT COUNT(*) INTO v_tables
  FROM pg_tables
  WHERE schemaname='public'
    AND tablename IN ('users','pools','training_types','slots','bookings','payments','attendance','admin_actions','notification_queue');

  IF v_tables <> 9 THEN
    RAISE EXCEPTION 'Expected 9 tables, got %', v_tables;
  END IF;

  SELECT COUNT(*) INTO v_types
  FROM pg_type t
  JOIN pg_namespace n ON n.oid=t.typnamespace
  WHERE n.nspname='public'
    AND t.typname IN ('user_role','slot_status','booking_status','payment_status','notification_channel','notification_status');

  IF v_types <> 6 THEN
    RAISE EXCEPTION 'Expected 6 custom types, got %', v_types;
  END IF;

  SELECT COUNT(*) INTO v_funcs
  FROM pg_proc p
  JOIN pg_namespace n ON n.oid=p.pronamespace
  WHERE n.nspname='public'
    AND p.proname IN ('set_updated_at','is_booking_counted','increment_slot_booked_count','decrement_slot_booked_count','sync_slot_booked_count');

  IF v_funcs <> 5 THEN
    RAISE EXCEPTION 'Expected 5 functions, got %', v_funcs;
  END IF;

  SELECT COUNT(*) INTO v_indexes
  FROM pg_indexes
  WHERE schemaname='public'
    AND indexname IN (
      'idx_slots_pool_starts_at',
      'idx_slots_status_starts_at',
      'idx_bookings_user_status_created',
      'idx_bookings_slot_status',
      'idx_bookings_reserved_until',
      'idx_payments_booking_status',
      'idx_payments_created_at',
      'idx_admin_actions_entity_created',
      'idx_notification_queue_status_scheduled',
      'idx_notification_queue_user_created'
    );

  IF v_indexes <> 10 THEN
    RAISE EXCEPTION 'Expected 10 custom indexes, got %', v_indexes;
  END IF;

  SELECT COUNT(*) INTO v_triggers
  FROM pg_trigger t
  JOIN pg_class c ON c.oid=t.tgrelid
  JOIN pg_namespace n ON n.oid=c.relnamespace
  WHERE n.nspname='public'
    AND t.tgisinternal = false
    AND t.tgname IN (
      'trg_users_set_updated_at',
      'trg_pools_set_updated_at',
      'trg_training_types_set_updated_at',
      'trg_slots_set_updated_at',
      'trg_bookings_set_updated_at',
      'trg_payments_set_updated_at',
      'trg_notification_queue_set_updated_at',
      'trg_bookings_sync_slot_booked_count'
    );

  IF v_triggers <> 8 THEN
    RAISE EXCEPTION 'Expected 8 triggers, got %', v_triggers;
  END IF;
END
$$;
SQL

  local smoke_email
  smoke_email="smoke_$(date +%s%N)@example.com"

  local smoke_counts after_insert after_cancel
  smoke_counts="$(psql_exec -qAt -v smoke_email="$smoke_email" <<'SQL'
BEGIN;

INSERT INTO users (role, full_name, email, password_hash)
VALUES ('client', 'Smoke User', :'smoke_email', 'hash')
RETURNING id AS user_id \gset

INSERT INTO pools (name, address, timezone)
VALUES ('Smoke Pool', 'Smoke Address', 'Europe/Moscow')
RETURNING id AS pool_id \gset

INSERT INTO training_types (name, duration_minutes, default_capacity, price)
VALUES ('Open Swim', 60, 2, 500.00)
RETURNING id AS training_type_id \gset

INSERT INTO slots (pool_id, training_type_id, starts_at, ends_at, capacity, price)
VALUES (:pool_id, :training_type_id, '2026-03-01 07:00:00+03', '2026-03-01 08:00:00+03', 2, 500.00)
RETURNING id AS slot_id \gset

INSERT INTO bookings (user_id, slot_id, status, reserved_until, price)
VALUES (:user_id, :slot_id, 'pending_payment', NOW() + INTERVAL '15 minutes', 500.00)
RETURNING id AS booking_id \gset

SELECT booked_count
FROM slots
WHERE id = :slot_id;

UPDATE bookings
SET status='cancelled', reserved_until=NULL
WHERE id=:booking_id;

SELECT booked_count
FROM slots
WHERE id = :slot_id;

ROLLBACK;
SQL
)"

  after_insert="$(echo "$smoke_counts" | sed -n '1p')"
  after_cancel="$(echo "$smoke_counts" | sed -n '2p')"

  if [[ "$after_insert" != "1" ]]; then
    echo "ERROR: smoke test failed, expected booked_count=1 after booking insert, got: ${after_insert:-<empty>}" >&2
    exit 1
  fi

  if [[ "$after_cancel" != "0" ]]; then
    echo "ERROR: smoke test failed, expected booked_count=0 after booking cancel, got: ${after_cancel:-<empty>}" >&2
    exit 1
  fi

  echo "Check completed: schema + fat-db smoke test passed"
}

main() {
  require_psql

  local cmd="${1:-}"
  case "$cmd" in
    up) run_up ;;
    down) run_down ;;
    status) run_status ;;
    check) run_check ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
