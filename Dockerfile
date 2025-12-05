FROM golang:1.25-bookworm AS builder

WORKDIR /src

# Зависимости отдельно для кеша
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Сборка бинарника (статически, без CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sso ./cmd/sso


FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Режим gin = release
ENV GIN_MODE=release

# Бинарник
COPY --from=builder /out/sso /app/sso

# Порт из конфигурации
EXPOSE 8080

# Запуск
ENTRYPOINT ["/app/sso"]
