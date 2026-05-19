# Weather + Users API

REST API сервис на Go с PostgreSQL для управления пользователями, отслеживания городов и получения погоды. Поддерживает аутентификацию по JWT и ролевую модель.

## Стек

- **Go 1.22** — язык сервиса
- **chi** — HTTP роутер
- **pgx/v5** — PostgreSQL драйвер
- **golang-jwt/jwt** — авторизация по токенам
- **wttr.in** — внешний Weather API (бесплатно, без ключей)
- **Docker Compose** — локальная БД
- **uber-go/zap** — структурированное логирование
- **stretchr/testify** — unit-тесты и моки
- **testcontainers-go** — интеграционные тесты с реальной БД

## Архитектура (Clean Architecture)

```
HTTP Request → Middleware (Auth) → Handler → Service → Repository → PostgreSQL
                                              ↓
                                         weatherClient → wttr.in
```

Проект строго разделен на слои:
- **handler**: Обработка HTTP запросов и маппинг в DTO
- **service**: Бизнес-логика (не зависит от HTTP)
- **repository**: Работа с БД
- **dto**: Объекты для передачи данных в API (Request/Response)
- **model**: Доменные модели базы данных
- **middleware**: Авторизация и проверки прав доступа

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
| `JWT_SECRET`   | `super-secret-key`                                        |
| `PORT`         | `8080`                                                    |

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
# Запуск всех тестов (для интеграционных нужен запущенный Docker)
go test -v ./...

# Проверка покрытия кода
go test -cover ./internal/service/...
```

## Структура проекта

```
.
├── cmd/server/main.go          # точка входа, DI-сборка
├── internal/
│   ├── config/                 # конфигурация из env
│   ├── dto/                    # API Request/Response структуры
│   ├── handler/                # HTTP-хендлеры
│   ├── middleware/             # JWT и авторизация
│   ├── service/                # бизнес-логика (с интерфейсами)
│   ├── repository/             # работа с базой данных
│   ├── model/                  # доменные модели (User, City и т.д.)
│   └── weather/                # клиент к wttr.in
├── migrations/                 # SQL миграции
└── docker-compose.yml
```
