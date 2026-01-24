# SSO Service

Микросервис централизованной аутентификации и авторизации, реализующий протокол SSO (Single Sign-On). Сервис предоставляет gRPC API для регистрации, входа пользователей, управления правами доступа и валидации JWT-токенов. Написан на Go, использует PostgreSQL в качестве хранилища данных.

## 🚀 Ключевые особенности

- **gRPC API**: Высокопроизводительный интерфейс на базе Protocol Buffers для взаимодействия между сервисами.
- **JWT Authentication**: Выпуск и валидация Access-токенов (HMAC SHA256).
- **Secure Storage**: Безопасное хранение паролей с использованием bcrypt.
- **Role-Based Access Control (RBAC)**: Поддержка ролей (Admin/User) для разграничения доступа.
- **PostgreSQL**: Надежное хранение данных пользователей и приложений.
- **Interceptors**: Встроенные механизмы логирования и восстановления после паники.
- **Configurable**: Гибкая настройка через YAML-конфиг и переменные окружения.

## 🛠 Предварительные требования

Для запуска и разработки вам понадобятся:

- **Go** (версия 1.25+)
- **Docker** и **Docker Compose**
- **Make** (опционально, для удобства запуска команд)
- **gRPC Client** (например, [BloomRPC](https://github.com/bloomrpc/bloomrpc) или [grpcurl](https://github.com/fullstorydev/grpcurl)) для тестирования.

## ⚡ Быстрый старт

### 1. Клонирование и настройка

```bash
git clone https://github.com/bmstu-itstech/sso.git
cd sso
# Создайте файл конфигурации (если используется локальный запуск без Docker)
# cp config/local.yaml config/config.yaml
```

### 2. Запуск через Docker Compose

Это рекомендуемый способ для развертывания локального окружения вместе с базой данных.

```bash
docker-compose up -d
```

Сервис будет доступен на порту, указанном в конфигурации (по умолчанию `44044`).

### 3. Проверка работы

```bash
# Пример проверки порта (если установлен netcat)
nc -zv localhost 44044
```

## 🔌 Интеграция с вашим микросервисом

Ниже приведен пример того, как другой Go-сервис может использовать клиент gRPC для взаимодействия с SSO. В данном примере мы проверяем права администратора для пользователя.

> **Важно:** Для полноценной валидации токена рекомендуется либо использовать общий секретный ключ (для локальной проверки подписи JWT), либо реализовать метод `ValidateToken` в SSO. В примере ниже показан вызов метода `IsAdmin`, который требует валидного токена.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type SSOClient struct {
	api ssov1.AuthClient
}

func NewSSOClient(addr string) (*SSOClient, error) {
	const op = "grpc.NewSSOClient"

	// Используем insecure credentials только для тестов/локальной разработки
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &SSOClient{
		api: ssov1.NewAuthClient(cc),
	}, nil
}

func (c *SSOClient) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, err
	}
	return resp.IsAdmin, nil
}

// AuthMiddleware пример middleware, который извлекает токен и делает запрос к SSO
// Примечание: В реальном сценарии лучше валидировать JWT локально публичным ключом для производительности.
func (c *SSOClient) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Добавляем токен в метаданные для gRPC запроса
		md := metadata.New(map[string]string{
			"authorization": "Bearer " + token,
		})
		ctx := metadata.NewOutgoingContext(r.Context(), md)

		// Пример: проверяем, является ли пользователь админом.
		// Внимание: для этого нужно знать UserID. Обычно он извлекается из claims токена.
		// Здесь для примера мы используем хардкод или извлекаем из заголовка (небезопасно без проверки подписи).
		// userID := extractUserIdFromToken(token)
		var userID int64 = 1 // Заглушка

		isAdmin, err := c.IsAdmin(ctx, userID)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !isAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func main() {
	sso, err := NewSSOClient("localhost:44044")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/admin", sso.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome, Admin!"))
	}))

	log.Println("Service started on :8080")
	http.ListenAndServe(":8080", nil)
}
```

## 📦 Структура проекта

```
.
├── cmd/sso/            # Точка входа (main.go)
├── internal/
│   ├── app/            # Приложение (gRPC сервер)
│   ├── config/         # Конфигурация
│   ├── domain/         # Бизнес-логика и модели
│   ├── grpc/           # Реализация gRPC хендлеров
│   ├── services/       # Сервисный слой (Auth, JWT)
│   └── storage/        # Работа с БД (PostgreSQL)
├── migrations/         # SQL миграции
└── tests/              # E2E и интеграционные тесты

## для запуска быстро
docker run --name=sso-db -e POSTGRES_PASSWORD=qwerty -p 5436:5432 -d postgres
migrate -path ./migrations -database 'postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable' up
```
