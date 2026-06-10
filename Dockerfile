# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /app/checkout-service ./main.go

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/checkout-service /app/checkout-service
COPY --from=builder /app/.env.example /app/.env

EXPOSE 9000

CMD ["sh", "-c", "/app/checkout-service migrate up && exec /app/checkout-service server rest"]
