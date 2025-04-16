# Используем официальный образ Go для сборки
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Компилируем приложение
RUN go build -o main main.go datasave.go

# Используем минимальный образ для запуска
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]