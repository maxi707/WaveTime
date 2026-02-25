# WaveTime Migrations

SQL-миграции PostgreSQL в формате `*.up.sql` и `*.down.sql`.

## Предусловия
- Нужен `psql`.
- По умолчанию скрипт ищет бинарь в `PG_BIN_DIR`:
  - `PG_BIN_DIR=${PG_BIN_DIR:-$HOME/ARENADATA/github/orioledb/output_bin/bin}`
- Для изоляции от других проектов использовать отдельные роль и БД:

```bash
createuser -h localhost -p 5432 wavetime
createdb  -h localhost -p 5432 -O wavetime wavetime
```

## Миграции

Порядок `up`:
1. `000001_extensions_and_types.up.sql`
2. `000002_core_tables.up.sql`
3. `000003_indexes.up.sql`
4. `000004_functions_and_triggers.up.sql`
5. `000005_seed_default_admin.up.sql`
6. `000006_seed_reference_data.up.sql`
7. `000007_booking_seats.up.sql`
8. `000008_seed_test_month_schedule.up.sql` (опционально, тестовое расписание на месяц)

Порядок `down`:
1. `000008_seed_test_month_schedule.down.sql` (если применяли)
2. `000007_booking_seats.down.sql`
3. `000006_seed_reference_data.down.sql`
4. `000005_seed_default_admin.down.sql`
5. `000004_functions_and_triggers.down.sql`
6. `000003_indexes.down.sql`
7. `000002_core_tables.down.sql`
8. `000001_extensions_and_types.down.sql`

## Использование `migrate.sh`

```bash
backend/migrations/migrate.sh <up|down|status|check>
```

Примеры:
```bash
backend/migrations/migrate.sh up
backend/migrations/migrate.sh status
backend/migrations/migrate.sh check
backend/migrations/migrate.sh down
```

Переменные подключения:
- `PGHOST` (default `localhost`)
- `PGPORT` (default `5432`)
- `PGUSER` (default `wavetime`)
- `PGDATABASE` (default `wavetime`)
- `PG_BIN_DIR` (default `$HOME/ARENADATA/github/orioledb/output_bin/bin`)
- `PSQL_BIN` (опционально, полный путь к `psql`)

Важно:
- `up` применяет все найденные `*.up.sql`, включая `000008`.
- `status` и `check` в текущем скрипте валидируют базовую схему до `000007`.
- `000008` считается опциональным seed и не влияет на `Overall: OK`.

## Ручное применение через `psql`

```bash
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000001_extensions_and_types.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000002_core_tables.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000003_indexes.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000004_functions_and_triggers.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000005_seed_default_admin.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000006_seed_reference_data.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000007_booking_seats.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000008_seed_test_month_schedule.up.sql
```

## Smoke-проверка

```bash
backend/migrations/migrate.sh status
backend/migrations/migrate.sh check
psql -h localhost -p 5432 -U wavetime -d wavetime -c "SELECT COUNT(*) FROM slots;"
```

Default admin после `000005`:
- login: `admin`
- password: `admin`
