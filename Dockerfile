# Берем официальный образ Go
FROM golang:1.21-alpine

# Устанавливаем зависимости для компиляции (например, для SQLite)
RUN apk add --no-cache gcc musl-dev

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . ./

# Создаем необходимые директории
RUN mkdir -p uploads

# Собираем приложение
RUN go build -o /file-exchange-app

# Экспонируем порт, на котором работает приложение
EXPOSE 8080

# Команда для запуска приложения
CMD [ "/file-exchange-app" ]