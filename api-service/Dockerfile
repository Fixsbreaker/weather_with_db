FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/api-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-service .
COPY migrations ./migrations

CMD ["/app/api-service"]
