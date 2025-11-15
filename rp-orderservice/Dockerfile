FROM golang:1.25.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /orderservice ./cmd/orderservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /orderservice /app/orderservice

ENTRYPOINT ["/app/orderservice"]