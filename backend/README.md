# WaveTime Backend

Go REST API для WaveTime.

## Требования
- Go 1.22+
- PostgreSQL с накаченными миграциями из `backend/migrations`

## Переменные окружения
- `DATABASE_URL` (обязательно), пример: `postgres://wavetime@localhost:5432/wavetime?sslmode=disable`
- `APP_SECRET` (обязательно), секрет подписи токенов
- `HTTP_ADDR` (опционально), по умолчанию `:8080`

## Запуск
```bash
cd backend
export DATABASE_URL="postgres://wavetime@localhost:5432/wavetime?sslmode=disable"
export APP_SECRET="dev-secret-change-me"
go run ./cmd/api
```

## Сборка
```bash
cd backend
go build ./...
```

## API

### Service
- `GET /healthz`

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
- `GET /training-types` (Bearer)
- `GET /schedule?pool_id&date_from=YYYY-MM-DD&date_to=YYYY-MM-DD` (Bearer)

### Bookings
- `POST /bookings` (Bearer)
- `GET /bookings/my` (Bearer)
- `PATCH /bookings/{id}/seats` (Bearer, body: `{"action":"inc"|"dec"}`)
- `DELETE /bookings/{id}` (Bearer)

### Payments
- `POST /payments/init` (Bearer)
- `POST /payments/webhook`

### Admin
- `GET /admin/bookings?status=&limit=` (Bearer, admin)
- `PATCH /admin/bookings/{id}` (Bearer, admin)
- `POST /admin/slots` (Bearer, admin)
- `PATCH /admin/slots/{id}` (Bearer, admin)

## Замечания по реализации
- Токены stateless (HMAC bearer token), хранения refresh-сессий в БД нет.
- Логин выполняется по полю `login`, которое проверяется как `email` или `phone`.
- Админ по умолчанию: `admin/admin` (создается миграцией `000005`).
- Повторное бронирование того же слота пользователем увеличивает `seats_count`.
- Расчет доступных мест в расписании учитывает сумму `seats_count` по активным броням.
