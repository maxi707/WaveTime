# WaveTime Backend (Go REST API)

## Требования
- Go 1.22+
- PostgreSQL со схемой из `backend/migrations`

## Переменные окружения
- `DATABASE_URL` (обязательно), пример: `postgres://wavetime@localhost:5432/wavetime?sslmode=disable`
- `APP_SECRET` (обязательно), секрет для подписи токенов
- `HTTP_ADDR` (опционально, по умолчанию `:8080`)
- `PGHOST/PGPORT/PGUSER/PGDATABASE` (для `migrations/migrate.sh`, не для API)

## Быстрый запуск локального PG-кластера (для проверки)
```bash
# 1) Бинарники PostgreSQL
export PG_BIN_DIR="..."
export PATH="$PG_BIN_DIR:$PATH"

# 2) Временный data-dir
export PGDATA="/tmp/pgdata-wavetime"
rm -rf "$PGDATA"

# 3) Инициализировать и запустить кластер
initdb -D "$PGDATA" --encoding=UTF8 --locale=C.UTF-8
pg_ctl -D "$PGDATA" -l "$PGDATA/server.log" start

# 4) Создать роль/БД приложения
createuser -h localhost -p 5432 wavetime
createdb  -h localhost -p 5432 -O wavetime wavetime

# 5) Подключение для API и миграций
export DATABASE_URL="postgres://wavetime@localhost:5432/wavetime?sslmode=disable"
```

Остановить и удалить кластер после проверки:
```bash
pg_ctl -D "$PGDATA" stop -m fast
rm -rf "$PGDATA"
```

## Запуск
```bash
cd backend
GOTOOLCHAIN=local go run ./cmd/api
```

## Сборка
```bash
cd backend
GOTOOLCHAIN=local go build ./...
```

## Быстрый цикл БД
```bash
cd backend
./migrations/migrate.sh up
./migrations/migrate.sh status
./migrations/migrate.sh check
```

## REST API (MVP)

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh` (Bearer)
- `POST /auth/logout` (Bearer)

### Profile
- `GET /me` (Bearer)
- `PATCH /me` (Bearer)

### Schedule
- `GET /pools` (Bearer)
- `GET /schedule?pool_id&date_from=YYYY-MM-DD&date_to=YYYY-MM-DD` (Bearer)

### Bookings
- `POST /bookings` (Bearer)
- `GET /bookings/my` (Bearer)
- `DELETE /bookings/{id}` (Bearer)

### Payments
- `POST /payments/init` (Bearer)
- `POST /payments/webhook`

### Admin
- `GET /admin/bookings` (Bearer, role=admin)
- `PATCH /admin/bookings/{id}` (Bearer, role=admin)
- `POST /admin/slots` (Bearer, role=admin)
- `PATCH /admin/slots/{id}` (Bearer, role=admin)

## Примечания
- Авторизация сделана через подписанный bearer token (HMAC).
- `refresh/logout` реализованы в stateless-варианте (без хранения сессий в БД).
- Логика `booked_count` и защита от овербукинга опираются на триггеры/ограничения БД (Fat DB).
