# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod ./
RUN go mod download

# Копируем исходный код
COPY main.go ./

# Компилируем приложение
RUN go build -o p2p-server main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем CA certificates для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем скомпилированный бинарник
COPY --from=builder /app/p2p-server .

# Копируем статические файлы
COPY static/ ./static/

# Копируем сертификаты
COPY certs/ ./certs/

# Даем права на привилегированные порты (80 и 443)
RUN apk --no-cache add libcap && \
    setcap cap_net_bind_service=+ep /app/p2p-server

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 -S p2p && \
    adduser -u 1000 -S p2p -G p2p && \
    chown -R p2p:p2p /app

# Переключаемся на непривилегированного пользователя
USER p2p

# Открываем порты
EXPOSE 80 443

# Запускаем приложение
CMD ["./p2p-server"]
