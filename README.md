# SSO — Single Sign-On (AuthN/AuthZ)

**SSO** — сервис централизованной аутентификации/авторизации для экосистемы приложений.
Выдаёт и валидирует **JWT**, хранит пользователей в **PostgreSQL**, ускоряет горячие чтения через **in-memory cache**.

## 🚀 Быстрый старт (1 команда)

```bash
docker-compose up -d
```

- Ping: `GET http://localhost:{HTTP_PORT}/api/v1/ping`
- Swagger UI: `http://localhost:{HTTP_PORT}/swagger/index.html`

---

## ✨ Чем проект реально полезен (уникальность)

- **Одна бизнес-логика → два транспорта**: HTTP (Gin) + gRPC используют один сервисный слой.
- **JWT-first**: middleware валидирует токен и пробрасывает `uid/isAdmin/appId` в контекст.
- **RBAC без оверхеда**: базовая модель прав `isAdmin`.
- **Cache поверх Postgres**: Postgres — источник правды, кэш — ускорение чтений.
- **Документация как код**: Swagger генерируется из аннотаций и отдаётся самим сервисом.

---

## 🛠️ Стек

- Go
- HTTP: Gin
- gRPC
- JWT (HS256)
- bcrypt
- PostgreSQL + sqlx
- in-memory cache
- migrations: migrate
- Docker / Docker Compose

---

## 🏗️ Архитектура (коротко)

```mermaid
graph TD
  FE[Frontend] -->|HTTP JSON| HTTP[HTTP API (Gin)]
  SVC1[Service A] -->|gRPC| GRPC[gRPC API]
  SVC2[Service B] -->|gRPC| GRPC

  HTTP --> MW[JWT middleware]
  MW --> Core[Service layer]
  GRPC --> Core

  Core --> Cache[(In-memory cache)]
  Cache --> DB[(PostgreSQL)]
  Core --> DB
  Core --> Jwt[JWT service]
```

### Ключевые слои
- HTTP роуты: `internal/http/server.go`
- HTTP middleware: `internal/http/middleware.go`
- DTO: `internal/domain/models/http.go`
- gRPC: `internal/grpc/auth` + `internal/app/grpc`
- Business logic: `internal/services`
- DB: `internal/repository/postgres`
- Cache: `internal/repository/cache`

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

## 🗄️ Postgres + Cache

- Postgres — **источник правды** (регистрация/пароли/роль/пользователи).
- In-memory cache — **ускорение чтений** (не распределённый, очищается при рестарте, у каждой реплики свой).

---

## 🗄️ Миграции

- лежат в `migrations/`
- в `docker-compose` применяются контейнером `migrate` до старта `sso`

Локально:
```bash
migrate -path ./migrations -database "postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable" up
```

---

## ⚙️ Конфигурация (ключевое)

- `HTTP_PORT`, `GRPC_PORT`
- `POSTGRES_HOST`, `POSTGRES_EXTERNAL_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_SSL_MODE`
- `JWT_SECRET`, `JWT_TOKEN_TTL`

См. `internal/config/config.go`.

---

## ✅ Quality gates

```bash
go test ./...
swag init -g internal/http/docs.go -o ./docs
```

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

> Если хотите самый правильный backend UX: добавьте HTTP endpoint `/api/v1/verify_token` для сервисов без gRPC.
