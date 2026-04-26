# Weather + Users API

REST API сервис на Go с PostgreSQL для управления пользователями, отслеживания городов и получения погоды.

## Стек

- **Go 1.22** — язык сервиса
- **chi** — HTTP роутер
- **pgx/v5** — PostgreSQL драйвер
- **wttr.in** — внешний Weather API (бесплатно, без ключей)
- **Docker Compose** — локальная БД

## Архитектура

```
handler → service → repository → PostgreSQL
                 ↘
             weather.Client → wttr.in
```

## Запуск

```bash
# 1. Поднять PostgreSQL (миграции применятся автоматически)
docker-compose up -d

# 2. Запустить сервер
go run ./cmd/server
```

По умолчанию сервер слушает на `:8080`, БД — `localhost:5434`.

Переменные окружения:

| Переменная     | По умолчанию                                              |
|----------------|-----------------------------------------------------------|
| `DATABASE_URL` | `postgres://weather:weather@localhost:5434/weather?sslmode=disable` |
| `PORT`         | `8080`                                                    |

## API

### Пользователи

| Метод    | Путь            | Описание                    |
|----------|-----------------|-----------------------------|
| `POST`   | `/users`        | Создать пользователя        |
| `GET`    | `/users`        | Список пользователей        |
| `GET`    | `/users/{id}`   | Получить пользователя по ID |
| `PUT`    | `/users/{id}`   | Обновить пользователя       |
| `DELETE` | `/users/{id}`   | Soft delete пользователя    |

**POST /users**
```json
{ "name": "Alnur", "email": "alnur@example.com" }
```

### Города пользователя

| Метод    | Путь                          | Описание              |
|----------|-------------------------------|-----------------------|
| `POST`   | `/users/{id}/cities`          | Добавить город        |
| `GET`    | `/users/{id}/cities`          | Список городов        |
| `DELETE` | `/users/{id}/cities/{city_id}`| Удалить город         |

**POST /users/{id}/cities**
```json
{ "name": "Almaty" }
```

### Погода

| Метод | Путь                               | Описание                                      |
|-------|------------------------------------|-----------------------------------------------|
| `GET` | `/users/{id}/weather`              | Текущая погода по всем городам пользователя   |
| `GET` | `/users/{id}/weather/history`      | История погодных запросов с фильтрацией       |

**GET /users/{id}/weather/history**

| Query-параметр | Обязательный | Описание                    |
|----------------|--------------|-----------------------------|
| `city`         | да           | Фильтр по городу            |
| `limit`        | нет          | Максимум записей в ответе   |
| `offset`       | нет          | Смещение для пагинации      |

Пример ответа:
```json
{
  "user_id": 1,
  "city": "Almaty",
  "history": [
    {
      "temperature": 18,
      "description": "Partly cloudy",
      "requested_at": "2026-04-27T10:00:00Z"
    }
  ]
}
```

## Пример использования

```bash
# Создать пользователя
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alnur","email":"alnur@example.com"}'

# Добавить города
curl -X POST http://localhost:8080/users/1/cities \
  -H "Content-Type: application/json" \
  -d '{"name":"Almaty"}'

curl -X POST http://localhost:8080/users/1/cities \
  -H "Content-Type: application/json" \
  -d '{"name":"London"}'

# Получить погоду (параллельные запросы)
curl http://localhost:8080/users/1/weather

# История по городу
curl "http://localhost:8080/users/1/weather/history?city=Almaty&limit=10"
```

## Структура проекта

```
.
├── cmd/server/main.go          # точка входа, DI-сборка
├── internal/
│   ├── config/                 # конфигурация из env
│   ├── handler/                # HTTP-хендлеры
│   ├── service/                # бизнес-логика
│   ├── repository/             # работа с БД
│   ├── model/                  # доменные модели
│   └── weather/                # клиент к wttr.in
├── migrations/
│   └── 001_init.sql            # схема БД + индексы
└── docker-compose.yml
```

## База данных

```sql
users            -- пользователи (soft delete через deleted_at)
user_cities      -- города пользователя
weather_history  -- история погодных запросов

-- индекс для быстрой фильтрации истории
INDEX idx_weather_history_user_city ON weather_history (user_id, city)
```
