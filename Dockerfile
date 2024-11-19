FROM golang:1.22.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -v ./apiserver  

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/apiserver /usr/local/bin/apiserver

EXPOSE 8080

CMD ["apiserver"]
