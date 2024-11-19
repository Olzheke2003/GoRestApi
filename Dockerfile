FROM golang:1.22.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -v ./apiserver  # Путь для компиляции

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/apiserver /usr/local/bin/apiserver  # Копируем в /usr/local/bin

EXPOSE 8080

CMD ["apiserver"]
