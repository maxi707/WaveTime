# Migrations Guide (WaveTime)

Этот каталог содержит SQL-миграции для PostgreSQL в формате `*.up.sql` и `*.down.sql`.

## 1. Предусловия

Используй бинарники PostgreSQL из каталога:

```bash
$HOME/ARENADATA/github/orioledb/output_bin/bin
```

Чтобы не конфликтовать с другими проектами на том же сервере PostgreSQL, обязательно создай отдельные пользователя и БД:

```bash
$HOME/ARENADATA/github/orioledb/output_bin/bin/createuser -h localhost -p 5432 wavetime
$HOME/ARENADATA/github/orioledb/output_bin/bin/createdb  -h localhost -p 5432 -O wavetime wavetime
```

Примечание: команды выше выполняются пользователем PostgreSQL с правами на создание ролей/баз.

## 2. Порядок применения (`up`)

Миграции должны применяться строго по номеру в имени файла:

1. `000001_extensions_and_types.up.sql`
2. `000002_core_tables.up.sql`
3. `000003_indexes.up.sql`
4. `000004_functions_and_triggers.up.sql`

Пример ручного применения через `psql`:

```bash
export PATH="$HOME/ARENADATA/github/orioledb/output_bin/bin:$PATH"

psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000001_extensions_and_types.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000002_core_tables.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000003_indexes.up.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000004_functions_and_triggers.up.sql
```

## 3. Откат (`down`)

Откат выполняется в обратном порядке:

1. `000004_functions_and_triggers.down.sql`
2. `000003_indexes.down.sql`
3. `000002_core_tables.down.sql`
4. `000001_extensions_and_types.down.sql`

Пример:

```bash
export PATH="$HOME/ARENADATA/github/orioledb/output_bin/bin:$PATH"

psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000004_functions_and_triggers.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000003_indexes.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000002_core_tables.down.sql
psql -h localhost -p 5432 -U wavetime -d wavetime -v ON_ERROR_STOP=1 -f backend/migrations/000001_extensions_and_types.down.sql
```

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
