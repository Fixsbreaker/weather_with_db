# Weather + Users API (Микросервисная архитектура)

REST API решение на Go с разделением на два независимых сервиса, базой данных PostgreSQL, полной Docker-контейнеризацией и оркестрацией через Docker Compose.

## Стек

- **Go 1.25** — язык программирования сервисов
- **Docker & Docker Compose** — контейнеризация и оркестрация
- **chi** — HTTP роутер в API Service
- **pgx/v5** — PostgreSQL драйвер
- **golang-jwt/jwt** — авторизация по токенам
- **wttr.in** — внешний публичный Weather API (используется Gateway-сервисом)
- **uber-go/zap** — структурированное логирование
- **stretchr/testify** — unit-тесты и моки
- **testcontainers-go** — интеграционные тесты с реальной БД

## Архитектура проекта

Проект разделен на два отдельных сервиса для изоляции бизнес-логики и работы с внешними интеграциями:

```
                  ┌──────────────────────────────────────────────┐
                  │               DOCKER COMPOSE                 │
                  │                                              │
HTTP Requests ───>│ :8080 ──> [ API Service ] ───> [ PostgreSQL ]│
                  │                  │ (internal url)            │
                  │                  v                           │
                  │ :8081 ──> [ Gateway Service ] ───> wttr.in   │
                  └──────────────────────────────────────────────┘
```

1. **API Service (Порт 8080)**:
   - Содержит бизнес-логику, обрабатывает авторизацию по JWT и предоставляет REST API.
   - Подключается к PostgreSQL для сохранения пользователей, отслеживаемых городов и истории погоды.
   - Вызывает Gateway Service для получения свежей погоды по внутреннему Docker DNS: `http://gateway-service:8081`.

2. **Gateway Service (Порт 8081)**:
   - Изолирует работу с внешним API (`wttr.in`).
   - Принимает HTTP-запросы и осуществляет запросы во внешний мир с использованием таймаутов.
   - Не имеет доступа к PostgreSQL и не хранит состояние.

## Запуск проекта

Проект полностью контейнеризирован и запускается одной простой командой:

```bash
docker compose up --build
```

После успешного старта:
- **API Service** доступен по адресу `http://localhost:8080` (эпик-хэндлер здоровья: `http://localhost:8080/health`).
- **Gateway Service** доступен по адресу `http://localhost:8081` (хэндлер здоровья: `http://localhost:8081/health`).
- **PostgreSQL** доступен для внешних клиентов на порту `5433` (по умолчанию логин/пароль/бд: `weather`). Скрипты миграции применяются автоматически при первом запуске контейнера из папки `./api-service/migrations`.

### Переменные окружения API-сервиса:

| Переменная     | Значение в Compose                                         | Описание |
|----------------|------------------------------------------------------------|----------|
| `DATABASE_URL` | `postgres://weather:weather@postgres:5432/weather?sslmode=disable` | Подключение к СУБД |
| `JWT_SECRET`   | `super-secret-key`                                         | Ключ для подписи токенов |
| `PORT`         | `8080`                                                     | Порт API-сервиса |
| `GATEWAY_URL`  | `http://gateway-service:8081`                             | Внутренний URL Gateway-сервиса |

## API

### Авторизация

| Метод    | Путь             | Описание                    |
|----------|------------------|-----------------------------|
| `POST`   | `/auth/register` | Зарегистрировать пользователя |
| `POST`   | `/auth/login`    | Войти и получить JWT токен    |

### Пользователь (Нужен JWT)

| Метод    | Путь             | Описание                    |
|----------|------------------|-----------------------------|
| `GET`    | `/users/me`      | Получить свои данные        |
| `GET`    | `/cities`        | Список своих городов        |
| `POST`   | `/cities`        | Добавить город              |
| `DELETE` | `/cities/{id}`   | Удалить город               |
| `GET`    | `/weather`       | Текущая погода по всем городам |
| `GET`    | `/weather/history`| История запросов погоды    |

### Админка (Нужен JWT + роль admin)

| Метод    | Путь            | Описание                    |
|----------|-----------------|-----------------------------|
| `GET`    | `/users`        | Список всех пользователей   |
| `GET`    | `/users/{id}`   | Получить пользователя по ID |
| `PUT`    | `/users/{id}`   | Обновить пользователя       |
| `DELETE` | `/users/{id}`   | Удалить пользователя        |

## Пример использования

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alnur", "email":"alnur@example.com", "password":"password123"}'

# 2. Логин (получение токена)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alnur@example.com", "password":"password123"}'
  
# Скопируйте access_token из ответа для следующих запросов
export TOKEN="ваш_токен_здесь"

# 3. Добавление города
curl -X POST http://localhost:8080/cities \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Almaty"}'

# 4. Получение погоды
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/weather

# 5. История по городу
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/weather/history?city=Almaty&limit=10"
```

## Тестирование и Логирование

Проект полностью покрыт тестами (более 70% покрытия `service` слоя) и использует современные подходы:

- **Dependency Injection (DI)**: Все слои (handler, service, repository) связываются через интерфейсы и внедряются через конструкторы в `main.go`. Бизнес-логика не создает зависимости внутри себя.
- **Unit Тесты**: Реализованы для `handler` и `service` слоев. Для изоляции базы данных используются моки (Mock Repository) через `testify/mock`. Проверяются как позитивные (happy path), так и негативные сценарии (ошибки валидации, 404 Not Found и др.).
- **Integration Тесты**: Написаны для слоя `repository` с использованием `testcontainers-go` (поднимается временный контейнер PostgreSQL для проверки реальных SQL запросов).
- **Structured Logging**: Внедрен логгер `zap`. Добавлен `LoggingMiddleware`, который логирует все HTTP-запросы (метод, путь, статус, длительность, request_id). Ошибки в хендлерах также логируются в структурированном виде через контекст запроса.

Запуск тестов:
```bash
# Запуск тестов внутри API-сервиса
cd api-service && go test -v ./...

# Проверка покрытия кода
go test -cover ./internal/service/...
```

## Структура проекта

```
.
├── docker-compose.yml           # Единый конфигурационный файл для запуска всего проекта
│
├── api-service/                 # Основной сервис бизнес-логики и работы с БД
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/server/main.go       # Точка входа API Service
│   ├── migrations/              # SQL-миграции для PostgreSQL
│   └── internal/                # Слои чистой архитектуры (Clean Architecture)
│       ├── config/              # Конфигурация из env
│       ├── dto/                 # API DTO структуры
│       ├── handler/             # HTTP-хендлеры
│       ├── middleware/          # JWT и авторизация
│       ├── model/               # Модели БД
│       ├── repository/          # Работа с PostgreSQL
│       ├── service/             # Бизнес-логика
│       └── weather/             # HTTP-клиент, вызывающий Gateway Service
│
└── gateway-service/             # Отдельный сервис шлюза для работы с внешним wttr.in API
    ├── Dockerfile
    ├── go.mod
    └── cmd/
        └── main.go              # Точка входа Gateway Service
```
