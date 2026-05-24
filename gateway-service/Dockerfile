FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /app/gateway-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway-service .

CMD ["/app/gateway-service"]
