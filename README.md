# SSO — Single Sign-On (AuthN/AuthZ)

**SSO** — сервис централизованной аутентификации/авторизации для экосистемы приложений.
Выдаёт и валидирует **JWT**, хранит пользователей в **PostgreSQL**, ускоряет горячие чтения через **in-memory cache + Redis**.

---

## 🚀 Быстрый старт (1 команда)

```bash
# Скопировать конфиг
cp .env.example .env

# Запустить всё
docker-compose up -d
```

| Проверка | URL |
|----------|-----|
| Ping | `GET http://localhost:8080/api/v1/ping` |
| Swagger UI | `http://localhost:8080/swagger/index.html` |
| gRPC | `localhost:44044` |

---

## ✨ Уникальность / Ключевые решения

| Фича | Детали |
|------|--------|
| **Dual Transport** | Один Service Layer → HTTP (Gin) + gRPC. Zero code duplication. |
| **JWT-first Auth** | Middleware парсит токен, пробрасывает `uid/isAdmin/appId` в `context`. |
| **Redis Event Bus** | Инвалидация сессий: при смене пароля/удалении → `PUBLISH user:{id}` в Redis. |
| **PostgreSQL LISTEN/NOTIFY** | In-memory cache таблицы `apps` автоматически обновляется через триггер БД. |
| **RBAC** | Простая модель: `isAdmin` + self-access (пользователь видит/редактирует себя). |
| **Swagger as Code** | Генерируется из аннотаций, отдаётся самим HTTP-сервером. |
| **Graceful Shutdown** | Корректное завершение gRPC/HTTP серверов и Redis-соединений. |

---

## 🛠️ Стек технологий

| Категория | Технология |
|-----------|------------|
| Язык | Go 1.21+ |
| HTTP Framework | Gin |
| gRPC | grpc-go + protobuf |
| Auth | JWT HS256, bcrypt |
| Database | PostgreSQL 16 + sqlx |
| Cache L1 | In-memory (sync.Map) + PG LISTEN/NOTIFY |
| Cache L2 / Event Bus | **Redis** (go-redis/v9) |
| Migrations | golang-migrate |
| Config | Viper (env + .env file) |
| Docs | Swaggo → Swagger 2.0 |
| Deploy | Docker, Docker Compose |

---

## 🏗️ Архитектура

```mermaid
graph TD
  FE[Frontend]
  A[Service A]
  B[Service B]

  HTTPAPI[HTTP API Gin]
  GRPCAPI[gRPC API]
  MW[JWT Middleware]
  CORE[Service Layer]
  CACHE[In-memory Cache]
  REDIS[Redis]
  DB[(PostgreSQL)]
  JWTSVC[JWT Service]

  FE --> HTTPAPI
  A --> GRPCAPI
  B --> GRPCAPI

  HTTPAPI --> MW
  MW --> CORE
  GRPCAPI --> CORE

  CORE --> CACHE
  CORE --> REDIS
  CACHE --> DB
  CORE --> DB
  CORE --> JWTSVC

  DB -.->|LISTEN/NOTIFY| CACHE
  REDIS -.->|Token Invalidation| CORE
```

### Слои и ответственность

| Слой | Путь | Ответственность |
|------|------|-----------------|
| **HTTP Transport** | `internal/http/` | Gin роуты, middleware, валидация, Swagger |
| **gRPC Transport** | `internal/grpc/` | Protobuf handlers, interceptors |
| **Service Layer** | `internal/services/` | Бизнес-логика (auth, user, jwt) |
| **Repository** | `internal/repository/` | PostgreSQL, Redis, In-memory cache |
| **Domain Models** | `internal/domain/models/` | DTO, entities |
| **Config** | `internal/config/` | Viper: env + .env |
| **App Bootstrap** | `internal/app/` | DI, graceful start/stop |

### Ключевые слои (файлы)
| Назначение | Файл |
|------------|------|
| HTTP роуты | `internal/http/server.go` |
| HTTP middleware | `internal/http/middleware.go` |
| DTO | `internal/domain/models/http.go` |
| gRPC | `internal/grpc/auth/server.go` |
| Business logic | `internal/services/` |
| PostgreSQL | `internal/repository/postgres/postgres.go` |
| In-memory cache | `internal/repository/cache/in_memory_cash.go` |
| Redis | `internal/repository/redis/redis.go` |

---

## 🌐 HTTP API (основное)

BasePath: `/api/v1`

**Public**
- `GET  /ping`
- `POST /login`
- `POST /register`

**Private (нужен Authorization: Bearer {JWT})**
- `GET    /user/is_admin/:id`
- `GET    /user/info/:id`
- `GET    /user/info` (обычно admin)
- `POST   /user/update_token`
- `PUT    /user/` (update password)
- `DELETE /user/` (remove user)

**Error contract**
```json
{ "error": "message" }
```

---

## 📚 Swagger

Swagger генерируется из аннотаций в коде и отдаётся самим HTTP сервером.

### Где открыть
- UI: `http://{host}:{HTTP_PORT}/swagger/index.html`
- OpenAPI JSON: `http://{host}:{HTTP_PORT}/swagger/doc.json`

### Как пользоваться (чтобы не страдать)
1) Откройте UI.
2) Нажмите **Authorize**.
3) Вставьте токен в формате:

```
Bearer {JWT}
```

После этого можно вызывать приватные ручки прямо из Swagger UI.

### Обновить документацию

```bash
swag init -g internal/http/docs.go -o ./docs
```

Артефакты генерации:
- `docs/swagger.json`
- `docs/swagger.yaml`
- `docs/docs.go`

---

## 🗄️ Хранилища данных

### PostgreSQL — источник правды

| Таблица | Назначение |
|---------|------------|
| `users` | Логин, email, bcrypt-хэш пароля, isAdmin, timestamps |
| `apps` | Зарегистрированные приложения (appId, secret) |

**Фичи:**
- Миграции через `golang-migrate`
- Триггер `NOTIFY update_cache` при изменении `apps`
- Connection pooling через `sqlx`

### In-memory Cache (L1)

```
internal/repository/cache/in_memory_cash.go
```

| Аспект | Реализация |
|--------|------------|
| Структура | `sync.RWMutex` + `map[int32]AppRepos` |
| Инвалидация | PostgreSQL `LISTEN/NOTIFY` → автоматический reload |
| Область | Таблица `apps` (горячие данные для JWT валидации) |
| Ограничения | Не распределённый, при рестарте — холодный старт |

**Как работает:**
1. При старте — `SELECT * FROM apps` → заполнение map
2. Фоновая горутина слушает канал `update_cache`
3. При `NOTIFY` — полный reload таблицы

### Redis (L2 / Event Bus)

```
internal/repository/redis/redis.go
```

| Аспект | Реализация |
|--------|------------|
| Клиент | `github.com/redis/go-redis/v9` |
| Назначение | **Token invalidation** при смене пароля / удалении пользователя |
| Ключи | `user:{userId}` |
| TTL | Совпадает с `JWT_TOKEN_TTL` |

**Сценарий инвалидации:**
1. Пользователь меняет пароль → `Redis.PublishUserUpdated(userId)`
2. Ключ `user:123` = `1` с TTL = JWT_TOKEN_TTL
3. При каждом запросе: `HasUserChanges(userId)` → если `true`, токен невалиден
4. После истечения TTL ключ удаляется автоматически

**Зачем это нужно:**
- JWT stateless, но при компрометации токена нужна возможность его отозвать
- Redis — быстрая проверка без похода в PostgreSQL

---

## 🗄️ Миграции

- лежат в `migrations/`
- в `docker-compose` применяются контейнером `migrate` до старта `sso`

Локально:
```bash
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

---

## ⚙️ Конфигурация

Конфиг читается из **переменных окружения** (приоритет) и файла **`.env`** (fallback).

### Все переменные

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `ENV` | `local` | Окружение: `local` / `dev` / `prod` |
| `HTTP_PORT` | `8080` | Порт HTTP API |
| `GRPC_PORT` | `44044` | Порт gRPC API |
| `GRPC_TIMEOUT` | `10s` | Таймаут gRPC вызовов |
| `POSTGRES_HOST` | `localhost` | Хост PostgreSQL |
| `POSTGRES_EXTERNAL_PORT` | `5436` | Порт PostgreSQL |
| `POSTGRES_DB` | `postgres` | Имя базы данных |
| `POSTGRES_USER` | `postgres` | Пользователь БД |
| `POSTGRES_PASSWORD` | `qwerty` | Пароль БД |
| `POSTGRES_SSL_MODE` | `disable` | SSL режим |
| `POSTGRES_URI` | — | Полный URI (если задан, остальные PG-переменные игнорируются) |
| `REDIS_HOST` | `redis` | Хост Redis |
| `REDIS_PORT` | `6379` | Порт Redis |
| `JWT_SECRET` | `my-secret` | Секрет для подписи JWT (**сменить в prod!**) |
| `JWT_TOKEN_TTL` | `12h` | Время жизни токена |

### Пример `.env`

```dotenv
ENV=local
HTTP_PORT=8080
GRPC_PORT=44044
POSTGRES_HOST=localhost
POSTGRES_EXTERNAL_PORT=5436
POSTGRES_DB=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=qwerty
POSTGRES_SSL_MODE=disable
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
JWT_SECRET=super-secret-change-me
JWT_TOKEN_TTL=12h
```

> ⚠️ **Важно для Docker:** внутри `docker-compose` хосты переопределяются:
> - `POSTGRES_HOST=postgres`
> - `REDIS_HOST=redis`

См. `internal/config/config.go`.

---

## ✅ Quality Gates

```bash
# Тесты
go test ./...

# Тесты с покрытием
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Линтер
golangci-lint run

# Обновить Swagger
swag init -g internal/http/docs.go -o ./docs
```

---

## 🐳 Docker Compose — что внутри

```yaml
services:
  redis       # Redis Alpine — event bus для token invalidation
  postgres    # PostgreSQL 16 — основное хранилище
  migrate     # Применяет миграции до старта sso
  sso         # Само приложение (HTTP + gRPC)
```

### Порядок запуска (depends_on + healthcheck)

```
redis (healthy) ─┐
                 ├─► sso
postgres (healthy) ─► migrate (completed) ─┘
```

### Healthchecks

| Сервис | Проверка |
|--------|----------|
| `redis` | `redis-cli ping` |
| `postgres` | `pg_isready` |
| `sso` | `wget /api/v1/ping` |

---

## 🔐 JWT — детали реализации

| Аспект | Значение |
|--------|----------|
| Алгоритм | HS256 |
| Payload | `uid`, `appId`, `isAdmin`, `exp` |
| Хранение секрета | Переменная `JWT_SECRET` |
| Валидация | Middleware проверяет подпись + expiration |
| Инвалидация | Redis-ключ `user:{id}` при смене пароля |

### Структура токена (payload)

```json
{
  "uid": 123,
  "app_id": 1,
  "is_admin": false,
  "exp": 1706500000
}
```

---

## 🔄 Graceful Shutdown

При получении `SIGINT` / `SIGTERM`:

1. HTTP сервер: `server.Shutdown(ctx)` с таймаутом
2. gRPC сервер: `server.GracefulStop()`
3. Redis: закрытие connection pool
4. PostgreSQL: закрытие пула соединений

---

## 🧑‍💻 API без боли: шпаргалка для Frontend и Backend

Ниже — **коротко по каждой ручке**: что делает, что отправлять, что получать.

### Общие правила (очень важно)

**Base URL**
- Локально через compose: `http://localhost:{HTTP_PORT}`

**JSON**
- Всегда ставьте заголовок: `Content-Type: application/json`

**JWT для приватных методов**
- Передавайте JWT так:
  - `Authorization: Bearer {JWT}`

**Единый формат ошибки**
```json
{ "error": "message" }
```

---

### 1) Healthcheck

#### `GET /api/v1/ping`
- **Зачем**: проверить, что HTTP сервис жив.
- **Что отправлять**: ничего.
- **Что получишь**: `200 OK` + `{ "message": "pong" }`.

Frontend: используйте для health-check в dev окружении.

---

### 2) Auth

#### `POST /api/v1/register`
- **Зачем**: создать нового пользователя.
- **Тело запроса**:
```json
{ "login": "alice", "password": "StrongPass123", "email": "alice@example.com", "fullName": "Alice Doe" }
```
- **Успех**: `200 OK`
```json
{ "userId": 123 }
```
- **Типовые ошибки**:
  - `400` — невалидный JSON или не прошла валидация (например пароль < 8)
  - `409` — пользователь уже существует

Frontend совет:
- показывайте `error` из ответа пользователю;
- пароль делайте минимум 8 символов.

---

#### `POST /api/v1/login`
- **Зачем**: получить JWT.
- **Тело запроса**:
```json
{ "appId": 1, "login": "alice", "password": "StrongPass123" }
```
- **Успех**: `200 OK`
```json
{ "token": "{jwt}" }
```
- **Типовые ошибки**:
  - `400` — невалидный JSON/валидация
  - `401` — неверный логин/пароль

Frontend совет:
- храните JWT в памяти/secure storage (для браузера часто: httpOnly cookie на периметре; если храните в localStorage — осознавайте XSS риски).

---

### 3) User (private)

> Для всех ручек ниже требуется заголовок `Authorization: Bearer {JWT}`.

#### `GET /api/v1/user/is_admin/:id`
- **Зачем**: узнать, является ли пользователь админом.
- **Параметры**: `id` в path.
- **Успех**: `200 OK`
```json
{ "isAdmin": true }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `403` — вы не admin и пытаетесь проверить не себя
  - `404` — пользователь не найден

---

#### `GET /api/v1/user/info/:id`
- **Зачем**: получить профиль пользователя.
- **Параметры**: `id` в path.
- **Успех**: `200 OK` (пример)
```json
{ "userId": 123, "login": "alice", "email": "alice@example.com", "fullName": "Alice Doe", "isAdmin": false }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `403` — не admin и запрашиваете не себя
  - `404` — пользователь не найден

---

#### `GET /api/v1/user/info`
- **Зачем**: список всех пользователей.
- **Успех**: `200 OK`
```json
{ "users": [ {"userId": 1, "login": "root", "email": "root@local", "fullName": "Root", "isAdmin": true} ] }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `403` — требуется admin

---

#### `POST /api/v1/user/update_token`
- **Зачем**: обновить токен (refresh) для приложения.
- **Тело запроса**:
```json
{ "appId": 1 }
```
- **Успех**: `200 OK`
```json
{ "token": "{newJwt}" }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `400` — невалидный JSON

Frontend совет:
- если вы используете этот endpoint как refresh, заменяйте сохранённый токен на новый сразу.

---

#### `PUT /api/v1/user/` (update password)
- **Зачем**: смена пароля (сам себе или admin).
- **Тело запроса**:
```json
{ "userId": 123, "newPassword": "NewStrongPass123" }
```
- **Успех**: `200 OK`
```json
{ "message": "Password updated" }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `403` — не admin и пытаетесь сменить пароль другому

---

#### `DELETE /api/v1/user/` (remove user)
- **Зачем**: удалить пользователя (сам себя или admin).
- **Тело запроса**:
```json
{ "userId": 123 }
```
- **Успех**: `200 OK`
```json
{ "message": "User deleted" }
```
- **Ошибки**:
  - `401` — нет/невалидный токен
  - `403` — не admin и пытаетесь удалить другого

---

## 🧩 Примеры для Frontend (JS)

### Базовый helper
```js
const BASE = `http://localhost:${process.env.HTTP_PORT ?? 8081}`;

async function api(path, { method = 'GET', token, body } = {}) {
  const headers = {};
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const msg = data?.error ?? `HTTP ${res.status}`;
    throw new Error(msg);
  }

  return data;
}
```

### Register → Login → Profile
```js
const reg = await api('/api/v1/register', {
  method: 'POST',
  body: { login: 'alice', password: 'StrongPass123', email: 'alice@example.com', fullName: 'Alice Doe' }
});

const login = await api('/api/v1/login', {
  method: 'POST',
  body: { appId: 1, login: 'alice', password: 'StrongPass123' }
});

const me = await api(`/api/v1/user/info/${reg.userId}`, {
  method: 'GET',
  token: login.token,
});

console.log(me);
```

---

## 🧰 Примеры для Backend

### Вариант 1: Go (gRPC) — verify token

> gRPC реально удобен для внутренних сервисов: меньше накладных расходов на JSON, строгий контракт.

```go
package main

import (
	"context"
	"fmt"
	"net"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cc, err := grpc.Dial(net.JoinHostPort("localhost", "44044"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	client := ssov1.NewAuthClient(cc)

	resp, err := client.VerifyToken(context.Background(), &ssov1.VerifyTokenRequest{
		Token: "{jwt}",
		AppId: 1,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("uid=%d isAdmin=%v\n", resp.UserId, resp.IsAdmin)
}
```

### Вариант 2: Node.js (HTTP) — middleware для Express

```js
import express from 'express';

const SSO = process.env.SSO_HTTP_BASE_URL ?? 'http://localhost:8081';

async function requireJWT(req, res, next) {
  const auth = req.headers.authorization || '';
  if (!auth.startsWith('Bearer ')) return res.status(401).json({ error: 'Unauthorized' });

  // В этом проекте нет отдельного /verify по HTTP.
  // Практичный вариант для Node: хранить JWT secret и валидировать токен локально,
  // либо дергать gRPC VerifyToken из Node (через grpc-js), либо добавить HTTP /verify.
  return next();
}

const app = express();
app.get('/private', requireJWT, (req, res) => res.json({ ok: true }));
app.listen(3000);
```

### Вариант 3: Python (HTTP) — requests

```py
import os
import requests

BASE = os.getenv('SSO_HTTP_BASE_URL', 'http://localhost:8081')

# login
r = requests.post(f'{BASE}/api/v1/login', json={'appId': 1, 'login': 'alice', 'password': 'StrongPass123'})
r.raise_for_status()
token = r.json()['token']

# user info
uid = 123
r = requests.get(f'{BASE}/api/v1/user/info/{uid}', headers={'Authorization': f'Bearer {token}'})
print(r.status_code, r.json())
```

---

## 🛠️ Локальная разработка

### Вариант 1: Всё в Docker (рекомендуется)

```bash
cp .env.example .env
docker-compose up -d --build
```

### Вариант 2: Приложение локально, инфраструктура в Docker

```bash
# 1. Запустить только инфраструктуру
docker-compose up -d redis postgres migrate

# 2. Настроить .env для локального запуска
REDIS_HOST=127.0.0.1
POSTGRES_HOST=127.0.0.1

# 3. Запустить приложение
go run cmd/sso/main.go
```

### Вариант 3: Запуск через GoLand

1. Создайте Run Configuration → Go Build
2. Package path: `github.com/bmstu-itstech/sso/cmd/sso`
3. Working directory: корень проекта
4. Environment: скопируйте из `.env` или укажите путь к файлу

---

## 🐛 Troubleshooting

### Redis: connection refused

**Симптом:**
```
dial tcp 127.0.0.1:6379: connect: connection refused
```

**Причина:** Redis не запущен или недоступен.

**Решение:**
```bash
# Проверить статус
docker ps | grep redis

# Если не запущен
docker-compose up -d redis

# Проверить доступность порта (Windows PowerShell)
Test-NetConnection -ComputerName 127.0.0.1 -Port 6379
```

### Redis: IPv6 vs IPv4

**Симптом:**
```
dial tcp [::1]:6379: connect: connection refused
```

**Причина:** `localhost` резолвится в IPv6 `::1`, Docker слушает на IPv4.

**Решение:** В `.env` укажите `REDIS_HOST=127.0.0.1` вместо `localhost`.

### PostgreSQL: connection refused

**Решение аналогично Redis:**
```bash
docker-compose up -d postgres
# Дождаться healthcheck
docker-compose ps
```

### Миграции не применились

```bash
# Проверить логи
docker-compose logs migrate

# Применить вручную
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

### Контейнер sso постоянно перезагружается

**Причина:** Неверные хосты для Redis/PostgreSQL внутри Docker.

**Решение:** В `docker-compose.yml` должны быть переопределены:
```yaml
environment:
  REDIS_HOST: redis
  POSTGRES_HOST: postgres
```

---

## 📁 Структура проекта

```
sso/
├── cmd/sso/                    # Entrypoint
│   └── main.go
├── internal/
│   ├── app/                    # Bootstrap: DI, graceful shutdown
│   │   ├── app.go
│   │   ├── grpc/app.go
│   │   └── http/app.go
│   ├── config/                 # Viper config
│   │   └── config.go
│   ├── domain/
│   │   ├── models/             # DTO, entities
│   │   └── storage/            # Storage interfaces/errors
│   ├── grpc/                   # gRPC handlers
│   │   ├── auth/server.go
│   │   └── middleware/
│   ├── http/                   # HTTP handlers (Gin)
│   │   ├── server.go
│   │   ├── middleware.go
│   │   ├── validate.go
│   │   └── docs.go             # Swagger annotations entry
│   ├── logs/                   # Structured logging (slog)
│   ├── repository/
│   │   ├── cache/              # In-memory cache + PG LISTEN
│   │   ├── postgres/           # PostgreSQL repository
│   │   └── redis/              # Redis client
│   └── services/               # Business logic
│       ├── auth.go
│       ├── user.go
│       ├── services.go
│       └── jwt/jwt.go
├── migrations/                 # SQL миграции
├── docs/                       # Сгенерированный Swagger
├── tests/                      # Интеграционные тесты
├── benchmark/                  # Бенчмарки
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── README.md
```

---

## 📊 Производительность

| Операция | Среднее время | Примечание |
|----------|---------------|------------|
| Login | ~5ms | bcrypt verify (CPU-bound) |
| Token validation | <1ms | In-memory + HMAC |
| User info (cached) | <1ms | In-memory lookup |
| User info (DB) | ~2ms | PostgreSQL query |

> Бенчмарки: `go test -bench=. ./benchmark/`

---

## 🔒 Security Checklist

- [ ] Сменить `JWT_SECRET` на сложный секрет (32+ символов)
- [ ] Включить SSL для PostgreSQL в production
- [ ] Настроить Redis AUTH (пароль)
- [ ] Использовать HTTPS (через reverse proxy: nginx, traefik)
- [ ] Ограничить CORS в production
- [ ] Настроить rate limiting

---

## 📝 License

MIT
