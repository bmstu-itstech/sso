# SSO — Single Sign-On сервис (AuthN/AuthZ)

SSO — микросервис централизованной **аутентификации** и **авторизации** для ваших приложений. Он выдаёт и валидирует **JWT**, хранит пользователей в **PostgreSQL**, ускоряет горячие операции через **in-memory cache**, и предоставляет **два транспорта** для интеграции:

- **HTTP API (Gin)** — удобно для фронтенда и быстрых интеграций.
- **gRPC API** — удобно для межсервисного взаимодействия и высокой производительности.

> Философия проекта: один источник правды по пользователям/ролям + быстрые проверки токена/ролей в рантайме.

---

## Что в этом продукте «уникального»

- **Два интерфейса из одного ядра**: бизнес-логика в сервисном слое переиспользуется для HTTP и gRPC.
- **JWT-first**: выдача токена на логине, проверка токена в middleware, извлечение `uid`/`isAdmin`.
- **RBAC на минималках**: роль `isAdmin` как базовый уровень разграничения прав.
- **Кэш поверх Postgres**: снижает нагрузку на БД на частых запросах (например, проверка прав/получение пользователя).

---

## Технологический стек

**Язык/платформа**
- Go (см. `go.mod`)

**Транспорт**
- HTTP: `gin-gonic/gin`
- gRPC: `google.golang.org/grpc`

**Auth/Security**
- JWT: `golang-jwt/jwt/v5`
- Хэш паролей: `bcrypt` (`golang.org/x/crypto`)

**Хранилище**
- PostgreSQL (`lib/pq`, `sqlx`)

**Кэш**
- In-memory cache (реализация в `internal/repository/cache/in_memory_cash.go`)

**Документация API**
- Swagger (swaggo): генерируется в `docs/`, UI отдаётся сервером

---

## Архитектура (коротко)

- `internal/services` — бизнес-логика (регистрация, логин, валидация токена, права, операции с пользователями)
- `internal/repository/postgres` — работа с Postgres (источник правды)
- `internal/repository/cache` — быстрый кэш для горячих данных
- `internal/http` — HTTP сервер на Gin (роуты + middleware)
- `internal/grpc` — gRPC сервер (handlers + interceptor)
- `internal/app` — сборка приложения (wire зависимостей)

---

## Структура проекта

```text
.
├── cmd/
│   ├── sso/                 # entrypoint приложения (поднимает gRPC + HTTP)
│   └── client/              # пример клиента (может быть неактуален/в разработке)
├── internal/
│   ├── app/                 # сборка приложения (grpc/http)
│   ├── config/              # конфиг (viper)
│   ├── domain/models/       # модели домена и HTTP DTO
│   ├── grpc/                # gRPC handlers + middleware
│   ├── http/                # Gin сервер: роуты + middleware
│   ├── repository/
│   │   ├── postgres/        # репозиторий Postgres
│   │   └── cache/           # in-memory cache
│   └── services/            # сервисный слой + JWT сервис
├── migrations/              # SQL миграции
├── docs/                    # swagger.json/yaml (генерируется swag)
└── tests/                   # тесты (grpc/http)
```

---

## Документация API (Swagger)

Swagger генерируется через swaggo и отдаётся самим HTTP сервером.

- UI: `http://{host}:{port}/swagger/index.html`
- JSON: `http://{host}:{port}/swagger/doc.json`

### Обновить Swagger

```bash
swag init -g internal/http/docs.go -o ./docs
```

> Если вы хотите, чтобы документация открывалась по `/docx` (а не `/swagger/index.html`) — добавьте редирект `/docx -> /docx/index.html` и смонтируйте swagger UI на `/docx/*any`.

---

## Как устроены Postgres и cache (и зачем оба)

### Postgres — источник правды
PostgreSQL хранит пользователей/учётные данные/роль. Любые изменения состояния (регистрация, смена пароля, удаление пользователя) фиксируются в БД.

### In-memory cache — ускоритель
Кэш (в `internal/repository/cache/in_memory_cash.go`) — это быстрый слой в памяти процесса. Он нужен, чтобы:

- снижать количество запросов в Postgres на **частых чтениях**;
- ускорять **проверки** (например, роль/пользователь);
- сгладить пики нагрузки (когда много запросов на одни и те же данные).

Важно понимать: in-memory cache **не является распределённым**. Это значит:
- кэш живёт в рамках одного инстанса сервиса;
- при рестарте сервиса кэш очищается;
- при горизонтальном масштабировании кэш будет «разным» у разных реплик.

В этой архитектуре это нормальный компромисс: Postgres остаётся системой истины, а кэш — оптимизация для скорости.

---

## Быстрый старт (максимально быстро)

Ниже — два рабочих сценария: через Docker Compose (рекомендуется) и вручную.

### Вариант A. Docker Compose (recommended)

1) Поднять инфраструктуру:

```bash
docker-compose up -d
```

2) Накатить миграции (если compose их не накатывает автоматически):

```bash
migrate -path ./migrations -database "postgres://postgres:{PASSWORD}@localhost:{PORT}/{DB}?sslmode=disable" up
```

3) Запустить сервис:

```bash
go run ./cmd/sso
```

4) Проверка:

- `GET http://localhost:{HTTP_PORT}/api/v1/ping`
- Swagger: `http://localhost:{HTTP_PORT}/swagger/index.html`

### Вариант B. Поднять Postgres вручную + миграции + запуск

> Ваш предыдущий баг `failed to open database: EOF` почти всегда из-за того, что порт/контейнер не готовы или проброшен неверный порт. Важно пробрасывать **5432 внутри контейнера**.

1) Запустить Postgres:

```bash
docker run --name sso-db -e POSTGRES_PASSWORD=qwerty -p 5436:5432 -d postgres
```

2) Подождать, пока БД поднимется (пара секунд) и накатить миграции:

```bash
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

3) Запустить сервис:

```bash
go run ./cmd/sso
```

---

## Переменные окружения и конфигурация

Конфиг собирается через `viper`. Ключевые параметры смотрите в `internal/config/config.go`.

На практике вам понадобятся:
- `POSTGRES_HOST`
- `POSTGRES_EXTERNAL_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `GRPC_PORT`
- `HTTP_PORT`

---

## Полезные эндпоинты (HTTP)

Публичные:
- `GET  /api/v1/ping`
- `POST /api/v1/login`
- `POST /api/v1/register`

Приватные (нужен `Authorization: Bearer {JWT}`):
- `GET    /api/v1/user/is_admin/:id`
- `GET    /api/v1/user/info/:id`
- `GET    /api/v1/user/info` (обычно только для admin)
- `POST   /api/v1/user/update_token`
- `PUT    /api/v1/user/`
- `DELETE /api/v1/user/`

Детальный контракт и примеры — в Swagger UI.

---

## Тесты

- HTTP unit-тесты: `tests/http_test`
- HTTP интеграционные: `tests/http_integration_test` (ожидают поднятый сервис)
- gRPC тесты: `tests/grpc_test` (в репозитории могут быть в процессе синхронизации с proto)

Запуск выборочно:

```bash
go test ./tests/http_test
```

---

## Development заметки

- Swagger обновляется командой: `swag init -g internal/http/docs.go -o ./docs`
- HTTP сервер отдаёт Swagger UI по `/swagger/*any`.

---

## Лицензия

См. репозиторий проекта.

