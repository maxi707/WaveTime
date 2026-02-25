# Migrations Guide (WaveTime)

Этот каталог содержит SQL-миграции для PostgreSQL в формате `*.up.sql` и `*.down.sql`.

## 1. Предусловия

Использовать бинарники PostgreSQL (`PG_BIN_DIR`).

Чтобы не конфликтовать с другими проектами на том же сервере PostgreSQL, обязательно создать отдельные пользователя и БД:

```bash
createuser -h localhost -p 5432 wavetime
createdb  -h localhost -p 5432 -O wavetime wavetime
```

Примечание: команды выше выполняются пользователем PostgreSQL с правами на создание ролей/баз.

## 2. Порядок применения (`up`)

Миграции должны применяться строго по номеру в имени файла:

1. `000001_extensions_and_types.up.sql`
2. `000002_core_tables.up.sql`
3. `000003_indexes.up.sql`
4. `000004_functions_and_triggers.up.sql`
5. `000005_seed_default_admin.up.sql`
6. `000006_seed_reference_data.up.sql`

Пример ручного применения через `psql`:

```bash
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000001_extensions_and_types.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000002_core_tables.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000003_indexes.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000004_functions_and_triggers.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000005_seed_default_admin.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000006_seed_reference_data.up.sql
```

## 3. Откат (`down`)

Откат выполняется в обратном порядке:

1. `000006_seed_reference_data.down.sql`
2. `000005_seed_default_admin.down.sql`
3. `000004_functions_and_triggers.down.sql`
4. `000003_indexes.down.sql`
5. `000002_core_tables.down.sql`
6. `000001_extensions_and_types.down.sql`

Пример:

```bash
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000006_seed_reference_data.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000005_seed_default_admin.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000004_functions_and_triggers.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000003_indexes.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000002_core_tables.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000001_extensions_and_types.down.sql
```

Default admin seed after `up`:
- login: `admin`
- password: `admin`

## 4. Быстрая проверка после `up`

```bash
psql -h localhost -p 5432 -U wavetime -d wavetime -c "\dt"
psql -h localhost -p 5432 -U wavetime -d wavetime -c "\dT"
```

Ожидается наличие таблиц (`users`, `pools`, `training_types`, `slots`, `bookings`, `payments`, `attendance`, `admin_actions`, `notification_queue`) и пользовательских типов (`user_role`, `slot_status`, `booking_status`, `payment_status`, `notification_channel`, `notification_status`).

## 5. Запуск одной командой через `migrate.sh`

В каталоге миграций есть исполняемый скрипт:

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

Скрипт использует те же параметры подключения по умолчанию (`localhost:5432`, `wavetime/wavetime`) и путь к `psql` из `PG_BIN_DIR`.
