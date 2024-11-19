FROM golang:1.22.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -v ./cmd/apiserver

FROM alpine:latest

RUN apk --no-cache add ca-certificates


EXPOSE 8080

CMD ["apiserver"]
