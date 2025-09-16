# Стадия сборки
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY . .

# Устанавливаем совместимую версию go-sqlite3
RUN go get github.com/mattn/go-sqlite3@v1.14.22

RUN CGO_ENABLED=1 go build -o file-exchange-app

# Стадия запуска  
FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY --from=builder /app/file-exchange-app .
COPY --from=builder /app/uploads ./uploads
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./file-exchange-app"]