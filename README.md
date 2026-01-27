# SSO — Single Sign-On Service (AuthN/AuthZ)

> **TL;DR**: Сервис централизованной аутентификации и авторизации.
> Выдаёт и проверяет JWT, хранит пользователей в PostgreSQL, ускоряет горячие чтения через in-memory cache.
> Даёт **HTTP API (Gin)** для фронта и интеграций + **gRPC API** для микросервисов.

---

## 📌 Суть проекта

### Что такое SSO в рамках этого репозитория
SSO (Single Sign-On) в данном проекте — это **единая точка входа**, которая:

1. Регистрирует пользователя.
2. Проверяет логин/пароль.
3. Выдаёт **JWT** как “пропуск” в экосистему сервисов.
4. Позволяет другим сервисам:
   - быстро понять, кто пользователь (`uid`),
   - имеет ли он роль admin,
   - выполнить операции управления пользователем.

### Для кого этот сервис
- Для **фронтенда**: логин/регистрация по HTTP + Swagger документация.
- Для **бэкенд‑сервисов**: проверка токена/прав через gRPC или самостоятельная валидация JWT (в зависимости от политики безопасности).

### Что вы получаете «из коробки»
- JWT‑аутентификация
- RBAC на минималках (`isAdmin`)
- Сервисный слой с повторным использованием логики
- Postgres‑хранилище + in-memory cache
- gRPC + HTTP (Gin)
- Swagger UI
- SQL миграции

---

## ✨ Уникальность продукта (pragmatic)

- **Два транспорта — одна бизнес‑логика**.
  HTTP и gRPC используют один и тот же сервисный слой, что снижает дублирование и риск расхождения поведения.

- **JWT‑first дизайн**.
  Сервер выдаёт токен сразу при логине, middleware валидирует токен и складывает `uid/isAdmin/appId` в context.

- **Кэш поверх Postgres**.
  Postgres — источник правды, кэш — ускоритель чтений.

- **Документация — часть продукта**.
  Swagger генерируется из кода и отдаётся самим HTTP сервером.

---

## 🛠️ Стек технологий

### Язык / платформа
- **Go** (см. `go.mod`)

### Транспорт
- HTTP: **Gin** (`github.com/gin-gonic/gin`)
- gRPC: `google.golang.org/grpc`

### Crypto / безопасность
- JWT: `github.com/golang-jwt/jwt/v5`
- bcrypt: `golang.org/x/crypto/bcrypt`

### Данные
- PostgreSQL: `github.com/lib/pq`
- SQL helper: `github.com/jmoiron/sqlx`

### Валидация
- `github.com/go-playground/validator/v10`

### Документация
- Swagger (swag): `github.com/swaggo/swag` + `github.com/swaggo/gin-swagger`

### Инфраструктура
- Docker
- Docker Compose
- migrate: `migrate/migrate` (через контейнер или локально)

---

## 🏗️ Архитектура

### Высокоуровневая схема

Сервис предоставляет **две внешние поверхности**:
- HTTP API для клиентов (browser/mobile/frontend).
- gRPC API для сервисов внутри кластера.

При этом ядро обработки (Service layer) едино.

```mermaid
graph TD

    subgraph "Clients"
        FE[Frontend / Mobile / Browser]
        BE[Other Backend Services]
    end

    subgraph "SSO Service"
        HTTP[HTTP API (Gin)]
        GRPC[gRPC API]
        MW[JWT Middleware]
        SVC[Service Layer]
        JWT[JWT Service]
        CACHE[(In-memory cache)]
        DB[(PostgreSQL)]
    end

    FE -->|JSON over HTTP| HTTP
    BE -->|gRPC calls| GRPC

    HTTP --> MW
    MW --> SVC
    GRPC --> SVC

    SVC --> JWT
    SVC --> CACHE
    CACHE --> DB
    SVC --> DB
```

### Поток логина и выдачи токена

```mermaid
graph LR
    A[POST /api/v1/login] --> B[Validate JSON]
    B --> C[Check password (bcrypt)]
    C --> D[Generate JWT]
    D --> E[Return {token}]
```

### Поток приватного запроса (JWT middleware)

```mermaid
graph TD
    R[Request with Authorization: Bearer JWT] --> M[authMiddleware]
    M -->|Parse + Verify| P[JWT Parse]
    P -->|OK| Ctx[ctx: uid/isAdmin/appId]
    Ctx --> H[Handler]
    P -->|Fail| U[401 Unauthorized]
```

### Слои проекта

- `internal/http` — HTTP слой: роутинг, middleware, JSON binding, swagger endpoint.
- `internal/grpc` — gRPC слой: handlers + interceptors.
- `internal/services` — бизнес‑логика: регистрация, логин, user management.
- `internal/repository/postgres` — запросы к PostgreSQL.
- `internal/repository/cache` — in-memory кэш.
- `internal/config` — конфигурация через env/yaml.
- `migrations` — SQL миграции.

---

## 📁 Структура репозитория (расширенно)

```text
.
├── cmd/
│   ├── sso/                      # Entry point: поднимает HTTP + gRPC
│   └── client/                   # Пример клиента (может быть устаревшим)
├── internal/
│   ├── app/
│   │   ├── grpc/                 # gRPC app wiring
│   │   └── http/                 # HTTP app wiring
│   ├── config/                   # viper + переменные окружения
│   ├── domain/
│   │   └── models/               # DTO и доменные модели
│   ├── grpc/
│   │   ├── auth/                 # gRPC сервер для auth
│   │   └── middleware/           # gRPC interceptor
│   ├── http/                     # Gin server (routes + middleware)
│   ├── repository/
│   │   ├── cache/                # In-memory cache
│   │   └── postgres/             # Postgres repository
│   ├── services/                 # Сервисный слой
│   │   └── jwt/                  # JWT parse/generate
│   └── logs/                     # slog handlers
├── migrations/                   # SQL миграции
├── docs/                         # swagger.json/yaml (генерируется)
└── tests/
    ├── http_test/                # unit-тесты HTTP слоя
    ├── http_integration_test/    # интеграционные тесты HTTP (нужен поднятый сервис)
    └── grpc_test/                # gRPC тесты (может требовать синхронизации с proto)
```

---

## 🔐 Модель безопасности

### JWT
JWT — это токен, который содержит (минимально):
- `uid` — идентификатор пользователя
- `app_id` — идентификатор приложения (если применимо)
- `is_admin` — признак роли
- `exp` — срок жизни

JWT подписывается секретом (HMAC).

### Как передавать токен
HTTP:
- `Authorization: Bearer {JWT}`

gRPC:
- как metadata (зависит от конкретных proto; в проекте уже есть middleware/интерсепторы)

### RBAC
В проекте используется простая модель:
- `isAdmin = true` — администратор
- `isAdmin = false` — обычный пользователь

Правила:
- Пользователь может получать/обновлять свои данные.
- Администратор может получать список пользователей и управлять любым пользователем.

---

## 🧠 Postgres + Cache (как это работает)

### Postgres — источник правды
Postgres хранит первичные данные:
- пользователи
- хэши паролей
- роль/admin flag
- аудиты/таймстемпы (если есть)

### In-memory cache — ускорение чтений
Кэш нужен для горячих операций, где:
- нагрузка чтения >> нагрузка записи,
- допустима eventual consistency на коротком окне,
- выгодно уменьшить round-trip в Postgres.

⚠️ Ограничения:
- кэш не распределённый
- кэш очищается при рестарте
- при нескольких репликах кэш у каждой реплики свой

Сценарии, где кэш помогает:
- частая проверка admin
- частое получение user info одного пользователя

Сценарии, где Postgres обязательно:
- регистрация
- смена пароля
- удаление пользователя

---

## 🌐 HTTP API (Gin)

### BasePath
- `/api/v1`

### Публичные эндпоинты
- `GET  /api/v1/ping`
- `POST /api/v1/login`
- `POST /api/v1/register`

### Приватные эндпоинты (JWT обязателен)
- `GET    /api/v1/user/is_admin/:id`
- `GET    /api/v1/user/info/:id`
- `GET    /api/v1/user/info` (обычно admin)
- `POST   /api/v1/user/update_token`
- `PUT    /api/v1/user/`
- `DELETE /api/v1/user/`

### Контракты JSON
Фактические DTO лежат в `internal/domain/models/http.go`.

Примеры:

**Login**
```json
{
  "appId": 1,
  "login": "alice",
  "password": "super-secret"
}
```

**Login Response**
```json
{
  "token": "{jwt}"
}
```

**Register**
```json
{
  "login": "alice",
  "password": "super-secret-123",
  "email": "alice@example.com",
  "fullName": "Alice Doe"
}
```

### Ошибки
Текущий контракт ошибок:
```json
{
  "error": "Permission denied"
}
```

Статусы:
- `400` — неверный JSON/валидация
- `401` — невалидный токен/нет заголовка
- `403` — недостаточно прав
- `404` — сущность не найдена
- `409` — конфликт (например, user already exists)
- `500` — внутренняя ошибка

---

## 📡 gRPC API

> gRPC контракт зависит от proto (подключён через модуль сгенерированных протобуфов).
> См. `internal/grpc/auth`.

Рекомендованный подход:
- использовать gRPC для межсервисных вызовов
- использовать HTTP как “public edge”

---

## 📚 Документация (Swagger)

Swagger генерируется автоматически на основе аннотаций в коде.

- UI: `http://{host}:{HTTP_PORT}/swagger/index.html`
- JSON: `http://{host}:{HTTP_PORT}/swagger/doc.json`

### Как обновить Swagger

```bash
swag init -g internal/http/docs.go -o ./docs
```

### Сгенерированные файлы
- `docs/swagger.json`
- `docs/swagger.yaml`
- `docs/docs.go`

---

## ⚙️ Конфигурация

Конфиг читается через `viper` (см. `internal/config/config.go`).

### Ключевые переменные

#### HTTP
- `HTTP_PORT` — порт HTTP сервера

#### gRPC
- `GRPC_PORT` — порт gRPC
- `GRPC_TIMEOUT` — таймауты (если используются)

#### Postgres
- `POSTGRES_HOST`
- `POSTGRES_EXTERNAL_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_SSL_MODE`
- `POSTGRES_URI` (если используется как full DSN)

#### JWT
- `JWT_SECRET`
- `JWT_TOKEN_TTL`

---

## 🚀 Быстрый старт (Docker Compose)

Если вы хотите “максимально быстро поднять всё” — достаточно:

```bash
docker-compose up -d
```

или (новый синтаксис Docker):

```bash
docker compose up -d
```

После запуска:
- Postgres поднимется и станет Healthy
- контейнер `migrate` накатит миграции
- контейнер `sso` стартанёт HTTP + gRPC

### Проверка (Smoke)

- Ping: `GET http://localhost:{HTTP_PORT}/api/v1/ping`
- Swagger: `http://localhost:{HTTP_PORT}/swagger/index.html`

---

## 🧰 Запуск без Compose (ручной)

### Поднять Postgres

```bash
docker run --name sso-db -e POSTGRES_PASSWORD=qwerty -p 5436:5432 -d postgres
```

### Миграции

```bash
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

### Запуск сервиса

```bash
go run ./cmd/sso
```

---

## 🧪 Тестирование

### Unit tests
```bash
go test ./tests/http_test
```

### Integration tests (HTTP)
Эти тесты ожидают, что сервис уже поднят на `SSO_HTTP_BASE_URL` (по умолчанию `http://localhost:8080`).

> Если вы меняете HTTP порт — задайте env:

```bash
SSO_HTTP_BASE_URL=http://localhost:{HTTP_PORT} go test ./tests/http_integration_test
```

---

## 🔭 Observability (логирование)

В проекте используется `slog` и кастомные хендлеры форматирования.

Рекомендации:
- выставлять ENV=prod на production
- собирать логи в централизованное хранилище

---

## 🧯 Troubleshooting

### 1) `failed to open database: EOF`
Чаще всего причины:
- Postgres контейнер ещё не готов
- проброшен неправильный порт (внутри контейнера всегда 5432)
- неверный DSN

Проверка:
- `docker ps` — контейнер postgres должен быть Running
- `docker logs sso-postgres` — нет бесконечных перезапусков

### 2) Swagger не открывается
Проверьте:
- HTTP порт (`HTTP_PORT`) — туда ли вы стучитесь
- URL: `/swagger/index.html`

### 3) Интеграционные тесты HTTP не видят ping
- Поднимите сервис заранее (compose или go run)
- Проверьте, что `SSO_HTTP_BASE_URL` совпадает с вашим портом

---

## 🧭 Runbook (операционные заметки)

### Обновить swagger
```bash
swag init -g internal/http/docs.go -o ./docs
```

### Обновить миграции
- добавить новый файл в `migrations/`
- проверить `docker compose up -d` (migrate контейнер применит изменения)

---

## 🔒 Безопасность и best practices

- В проде не используйте дефолтный `JWT_SECRET`.
- Пароль хранится в виде bcrypt hash.
- Включайте TLS на периметре (ingress/reverse proxy).

---

## 📜 Лицензия
См. репозиторий проекта.

---

## 🧾 Контракты API: правила и соглашения (важно для фронта)

### 1) Формат ошибок
Сервис возвращает единый формат ошибки:

```json
{
  "error": "human readable message"
}
```

Рекомендации фронту:
- не парсить текст ошибки, ориентируйтесь на HTTP status code
- текст ошибки используйте только для UI/логирования

### 2) Content-Type
- Все JSON запросы отправляйте с заголовком `Content-Type: application/json`

### 3) Authorization
- для приватных эндпоинтов добавляйте `Authorization: Bearer {JWT}`

### 4) Валидации
- валидация реализована через `go-playground/validator`
- поля с `validate:"required"` должны быть обязательно заполнены
- min-ограничения применяются только при наличии значения

---

## 🔑 JWT: структура, сроки жизни, best practices

### Что лежит внутри токена
Точный набор claim-ов зависит от реализации JWT сервиса, но логика проекта опирается на:
- `uid` (int64)
- `app_id` (int32)
- `is_admin` (bool)
- `exp` (timestamp)

### TTL
TTL задаётся через `JWT_TOKEN_TTL`.

Рекомендации:
- для прод окружения делайте TTL коротким (например, 15m–2h)
- если нужно «долго жить» — добавляйте refresh механизм на периметре

---

## 🧩 Взаимодействие сервисов (схема в стиле "Consumer → Router")

SSO выступает как централизованный провайдер идентичности.
Клиенты проходят логин/регистрацию по HTTP, а внутренние сервисы используют gRPC для быстрых проверок прав.

```mermaid
graph TD

  subgraph "External Ecosystem"
    Web[Frontend / Web]
    Mobile[Mobile]
  end

  subgraph "Internal Ecosystem"
    A[Service A]
    B[Service B]
  end

  subgraph "SSO Service"
    HTTP[HTTP API (Gin)]
    GRPC[gRPC API]
    MW[JWT Middleware]
    Auth[Auth Service]
    UserSvc[User Service]
    JwtSvc[JWT Service]
    Cache[(In-memory Cache)]
    DB[(PostgreSQL)]
  end

  Web -->|login/register (JSON)| HTTP
  Mobile -->|login/register (JSON)| HTTP

  HTTP --> MW
  MW --> Auth

  A -->|Verify/IsAdmin (gRPC)| GRPC
  B -->|Verify/IsAdmin (gRPC)| GRPC

  GRPC --> Auth

  Auth --> JwtSvc
  Auth --> Cache
  Cache --> DB
  Auth --> DB

  UserSvc --> DB
```

---

## 🗄️ Миграции: жизненный цикл схемы БД

Миграции лежат в `migrations/`.

### При запуске через docker-compose
Контейнер `migrate` автоматически применяет миграции до старта `sso`.

### При локальном запуске
Вы можете применить миграции вручную:

```bash
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

---

## 🧪 How-to: быстрый end-to-end сценарий руками

### 1) Запустить сервис
Самый быстрый путь:

```bash
docker-compose up -d
```

### 2) Зарегистрировать пользователя
`POST /api/v1/register`

### 3) Войти
`POST /api/v1/login` → получите `{jwt}`

### 4) Дёрнуть приватный endpoint
Пример: `GET /api/v1/user/info/:id` с заголовком `Authorization: Bearer {jwt}`

---

## ⚡ Производительность и масштабирование (практика)

### Где узкие места
- bcrypt: сравнительно дорогая операция (на логине/регистрации)
- Postgres: соединения/индексы/пул
- JWT parse: дешёво, но на высоком RPS важна оптимизация аллокаций

### Горизонтальное масштабирование
Сервис можно поднимать в нескольких репликах.
Ограничение: in-memory cache будет отдельным на каждой реплике.

Если потребуется единый кэш:
- добавьте Redis/Memcached
- или используйте кэш только как per-instance оптимизацию (как сейчас)

---

## 🧱 docker-compose: порты и сервисы

Состав (типовой):
- `postgres` — БД
- `migrate` — применяет миграции
- `sso` — приложение (HTTP + gRPC)

Порты:
- `HTTP_PORT` → HTTP API + Swagger
- `GRPC_PORT` → gRPC

---

## ✅ Quality gates (как проверять изменения)

Минимальный набор проверок перед PR:

```bash
# формат/линт у вас может быть настроен дополнительно

# бил��

go test ./...

# генерация swagger
swag init -g internal/http/docs.go -o ./docs
```

---

## 📎 Приложение: где что искать в коде (быстрые ссылки)

- HTTP роуты: `internal/http/server.go`
- HTTP middleware: `internal/http/middleware.go`
- DTO для HTTP: `internal/domain/models/http.go`
- Сборка HTTP app: `internal/app/http/app.go`
- gRPC app: `internal/app/grpc/app.go`
- JWT сервис: `internal/services/jwt/*`
- Репозитории: `internal/repository/*`

---

## 🧾 Примечание про одну команду старта

Для максимально быстрого старта достаточно:

```bash
docker-compose up -d
```

Если у вас современный Docker:

```bash
docker compose up -d
```
