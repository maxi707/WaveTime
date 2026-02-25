# WaveTime

WaveTime - веб-сервис сети бассейнов для записи и бронирования тренировок.

## Что реализовано
- Авторизация и регистрация пользователей.
- Личный кабинет с управлением бронями.
- Расписание слотов 07:00-23:00 и отображение доступных мест.
- Бронирование с поддержкой нескольких мест в одной брони (`seats_count`).
- Админ-раздел для управления слотами и статусами броней.
- Fat DB на PostgreSQL с триггерами и проверками целостности.

## Технологии
- Backend: Go 1.22, `net/http`, `pgx`.
- Frontend: Vue 3, Vue Router, Vite.
- DB: PostgreSQL, SQL-миграции в `backend/migrations`.

## Структура
- `backend` - Go API и доступ к PostgreSQL.
- `backend/migrations` - SQL-миграции и `migrate.sh`.
- `frontend` - SPA на Vue.

## Быстрый старт

1. Подготовить PostgreSQL и отдельные роль/БД:
```bash
createuser -h localhost -p 5432 wavetime
createdb  -h localhost -p 5432 -O wavetime wavetime
```

2. Накатить миграции:
```bash
backend/migrations/migrate.sh up
backend/migrations/migrate.sh status
backend/migrations/migrate.sh check
```

3. Запустить backend:
```bash
cd backend
export DATABASE_URL="postgres://wavetime@localhost:5432/wavetime?sslmode=disable"
export APP_SECRET="dev-secret-change-me"
go run ./cmd/api
```

4. Запустить frontend:
```bash
cd frontend
npm install
npm run dev
```

5. Открыть `http://127.0.0.1:5173`.

## Доступы по умолчанию
- Админ создается миграцией `000005`.
- Логин: `admin`
- Пароль: `admin`

## Важные бизнес-правила в текущей реализации
- Один пользователь имеет одну запись брони на слот (`UNIQUE (user_id, slot_id)`).
- Повторное нажатие "Забронировать" увеличивает `seats_count` в этой же брони.
- Отмена или уменьшение `seats_count` корректно освобождают места в расписании.
- Доступные места считаются как `capacity - SUM(seats_count)` по активным статусам.

## Документация по подпроектам
- [Backend README](backend/README.md)
- [Frontend README](frontend/README.md)
- [Migrations README](backend/migrations/README.md)
