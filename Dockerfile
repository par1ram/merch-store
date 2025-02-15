FROM golang:1.23 AS builder

WORKDIR /app

# Копируем go.mod и go.sum, скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN go build -o /build/merch-store ./cmd

# Финальный образ для минимального размера
FROM debian:stable-slim

WORKDIR /app

# Копируем бинарник из builder‑образа
COPY --from=builder /build/merch-store /app/merch-store

# (Если нужны миграции, схемы - скопируйте и их)
COPY --from=builder /app/internal/sql /app/internal/sql

# Порт, на котором будет слушать ваше приложение
EXPOSE 8080

# Запуск приложения
CMD ["/app/merch-store"]
