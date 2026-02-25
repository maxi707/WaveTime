# WaveTime Frontend

## Требования
- Node.js 20+
- Backend запущен на `http://127.0.0.1:8080`

## Установка и запуск
```bash
cd frontend
npm install
npm run dev
```

Vite dev server: `http://127.0.0.1:5173`.

## Маршруты
- `/` - стартовая страница с крупным логотипом и кнопками `Login` / `Register`
- `/login` - отдельная страница входа
- `/register` - отдельная страница регистрации
- `/app` - основная страница после авторизации

## Интеграция с backend
Во время разработки запросы идут через proxy:
- frontend -> `/api/...`
- Vite proxy -> `http://127.0.0.1:8080/...`

Это позволяет работать без CORS-настроек в backend.
